package gengo

import (
	"errors"
	"fmt"
	"go/token"
	"go/types"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"

	corecontext "context"
	"github.com/go-courier/logr"
	"github.com/octohelm/gengo/pkg/gengo/snippet"
	gengotypes "github.com/octohelm/gengo/pkg/types"
	reflectx "github.com/octohelm/x/reflect"
)

type Executor interface {
	Execute(ctx corecontext.Context, generators ...Generator) error
}

func NewContext(args *GeneratorArgs) (Executor, error) {
	u, err := gengotypes.Load(args.Entrypoint)
	if err != nil {
		return nil, err
	}
	c := &gengoCtx{
		universe: u,
		args:     args,
		l:        newLogger(),
	}
	return c, nil
}

type Context interface {
	IsZero() bool

	Logger() logr.Logger
	Defer(func(c Context) error)

	LocateInPackage(pos token.Pos) gengotypes.Package
	Package(importPath string) gengotypes.Package
	Doc(typ types.Object) (Tags, []string)

	Render(snippet snippet.Snippet)
	RenderT(template string, args ...snippet.TArg)
}

type gengoCtx struct {
	args     *GeneratorArgs
	universe *gengotypes.Universe

	pkgTags map[string][]string
	pkg     gengotypes.Package
	genfile *genfile

	ignore bool

	defers []func(ctx Context) error

	l logr.Logger
}

func (c *gengoCtx) IsZero() bool {
	return c.genfile.IsZero() && !c.ignore
}

func (c *gengoCtx) Defer(fn func(c Context) error) {
	c.defers = append(c.defers, fn)
}

func (c *gengoCtx) Logger() logr.Logger {
	return c.l
}

func (c *gengoCtx) RenderT(template string, args ...snippet.TArg) {
	c.genfile.Render(snippet.T(template, args...))
}

func (c *gengoCtx) Render(snippet snippet.Snippet) {
	c.genfile.Render(snippet)
}

func (c *gengoCtx) Writer() SnippetWriter {
	return c.genfile
}

func (c *gengoCtx) Execute(ctx corecontext.Context, generators ...Generator) error {
	for pkgPath, direct := range c.universe.LocalPkgPaths() {
		if !c.args.All && !direct {
			continue
		}

		if err := c.pkgExecute(logr.LoggerInjectContext(ctx, c.l), pkgPath, generators...); err != nil {
			return err
		}
	}

	return nil
}

func (c *gengoCtx) pkgExecute(pctx corecontext.Context, pkg string, generators ...Generator) (finalErr error) {
	ctx, l := logr.FromContext(pctx).Start(pctx, "generate", slog.String("scope", pkg))
	defer l.End()

	defer func() {
		if finalErr != nil {
			l.Error(finalErr)
		}
	}()

	p := c.universe.Package(pkg)
	if p == nil {
		return fmt.Errorf("invalid pkg `%s`", pkg)
	}

	generatedFiles := make(map[string]string)

	pkgCtx := &gengoCtx{
		universe: c.universe,
		args:     c.args,
		pkg:      p,
		pkgTags:  map[string][]string{},
	}

	for _, f := range p.Files() {
		fileFullname := p.FileSet().File(f.FileStart).Name()
		filename := filepath.Base(fileFullname)
		if strings.HasPrefix(filename, c.args.OutputFileBaseName+".") {
			generatedFiles[filename] = fileFullname
		}

		if f.Doc != nil && len(f.Doc.List) > 0 {
			tags, _ := gengotypes.ExtractCommentTags(strings.Split(f.Doc.Text(), "\n"))
			for k := range tags {
				pkgCtx.pkgTags[k] = tags[k]
			}
		}
	}

	gfs := sync.Map{}

	for _, gen := range generators {
		pkgCtxForGen := &gengoCtx{
			args:     pkgCtx.args,
			universe: pkgCtx.universe,
			pkg:      pkgCtx.pkg,
			pkgTags:  pkgCtx.pkgTags,
			genfile:  newGenfile(gen.Name()),
		}

		if err := pkgCtxForGen.genfile.InitWith(pkgCtxForGen); err != nil {
			return err
		}

		g := pkgCtxForGen.New(gen)

		pkgCtxForGen.l = l.WithValues("gengo", g.Name())

		if err := pkgCtxForGen.doGenerate(ctx, g); err != nil {
			return fmt.Errorf("`%s` generate failed for %s: %w", g.Name(), pkgCtx.pkg.Pkg().Path(), err)
		}

		for _, fn := range pkgCtxForGen.defers {
			if err := fn(pkgCtxForGen); err != nil {
				return fmt.Errorf("`%s` defer generate failed for %s: %w", g.Name(), pkgCtx.pkg.Pkg().Path(), err)
			}
		}

		if !pkgCtxForGen.IsZero() {
			gfs.Store(g.Name(), pkgCtxForGen.genfile)
		}
	}

	for _, w := range gfs.Range {
		gfile := w.(*genfile)

		if err := gfile.WriteToFile(pkgCtx, c.args); err != nil {
			return err
		}

		delete(generatedFiles, gfile.Filename(c.args))
	}

	if len(generatedFiles) > 0 {
		for _, fullFilename := range generatedFiles {
			if err := os.RemoveAll(fullFilename); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *gengoCtx) Package(importPath string) gengotypes.Package {
	if importPath == "" {
		return c.pkg
	}
	return c.universe.Package(importPath)
}

func (c *gengoCtx) LocateInPackage(pos token.Pos) gengotypes.Package {
	return c.universe.LocateInPackage(pos)
}

func (c *gengoCtx) Doc(typ types.Object) (Tags, []string) {
	tags, doc := c.universe.Package(typ.Pkg().Path()).Doc(typ.Pos())

	if len(doc) > 0 {
		doc[0] = strings.TrimSpace(strings.TrimPrefix(doc[0], typ.Name()))
		if len(doc[0]) == 0 {
			doc = doc[1:]
		}
	}

	return merge(c.args.Globals, c.pkgTags, tags), doc
}

func (c *gengoCtx) doGenerate(ctx corecontext.Context, g Generator) error {
	if c.pkg == nil {
		return nil
	}

	defer func() {
		if e := recover(); e != nil {
			c.l.Error(fmt.Errorf("doGenerate panic: %#v", e))
		}
	}()

	pkgTypes := c.pkg.Types()

	names := make([]string, 0)
	for n := range pkgTypes {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, n := range names {
		tpe := pkgTypes[n].Type()

		switch x := tpe.(type) {
		case *types.Alias:
			tags, _ := c.Doc(x.Obj())

			if IsGeneratorEnabled(g, tags) {
				if a, ok := g.(AliasGenerator); ok {
					if err := c.doGenerateAliasType(ctx, a, x); err != nil {
						return err
					}
				}
			}
		case *types.Named:
			tags, _ := c.Doc(x.Obj())

			if IsGeneratorEnabled(g, tags) {
				if err := c.doGenerateNamedType(ctx, g, x); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *gengoCtx) doGenerateNamedType(pctx corecontext.Context, g Generator, x *types.Named) error {
	_, l := c.l.Start(pctx, "debug: generate named", slog.String("scope", x.Obj().Pkg().Path()), slog.String("type", x.Obj().Name()))
	defer l.End()

	if err := g.GenerateType(c, x); err != nil {
		if errors.Is(err, ErrSkip) {
			return nil
		}
		if errors.Is(err, ErrIgnore) {
			l.Warn(err)
			// mark ignore to avoid remove previous generated
			c.ignore = true
			return nil
		}
		return err
	}

	return nil
}

func (c *gengoCtx) doGenerateAliasType(pctx corecontext.Context, g AliasGenerator, x *types.Alias) error {
	_, l := c.l.Start(pctx, "debug: generate alias", slog.String("scope", x.Obj().Pkg().Path()), slog.String("type", x.Obj().Name()))
	defer l.End()

	if err := g.GenerateAliasType(c, x); err != nil {
		if errors.Is(err, ErrSkip) {
			return nil
		}
		if errors.Is(err, ErrIgnore) {
			l.Warn(err)
			return nil
		}
		return err
	}

	return nil
}

func (c *gengoCtx) New(generator Generator) Generator {
	if creator, ok := generator.(GeneratorNewer); ok {
		return creator.New(c)
	}
	return reflect.New(reflectx.Indirect(reflect.ValueOf(generator)).Type()).Interface().(Generator)
}

func IsGeneratorEnabled(g Generator, tags map[string][]string) bool {
	prefix := "gengo:" + g.Name()

	enabled := false

	for k, values := range tags {
		if k == prefix {
			enabled = strings.Join(values, "") != "false"
			return enabled
		}

		if strings.HasPrefix(k, prefix+":") {
			enabled = true
		}
	}

	return enabled
}

type Tags map[string][]string
