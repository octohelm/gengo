package gengo

import (
	corecontext "context"
	"errors"
	"fmt"
	"go/token"
	"go/types"
	"log/slog"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/go-courier/logr"
	gengotypes "github.com/octohelm/gengo/pkg/types"
	reflectx "github.com/octohelm/x/reflect"
	"golang.org/x/sync/errgroup"
)

type Executor interface {
	Execute(ctx corecontext.Context, generators ...Generator) error
}

func NewContext(args *GeneratorArgs) (Executor, error) {
	u, pkgPaths, err := gengotypes.Load(args.Entrypoint)
	if err != nil {
		return nil, err
	}
	c := &context{
		pkgPaths: pkgPaths,
		args:     args,
		universe: u,
	}
	return c, nil
}

type Context interface {
	LocateInPackage(pos token.Pos) gengotypes.Package
	Package(importPath string) gengotypes.Package
	Doc(typ types.Object) (Tags, []string)
	Render(snippet Snippet)
	Logger() logr.Logger
}

type context struct {
	pkgPaths map[string]bool
	args     *GeneratorArgs

	universe gengotypes.Universe
	pkgTags  map[string][]string
	pkg      gengotypes.Package
	genfile  *genfile
	l        logr.Logger
}

func (c *context) Logger() logr.Logger {
	return c.l
}

func (c *context) Render(snippet Snippet) {
	c.genfile.Render(snippet)
}

func (c *context) Writer() SnippetWriter {
	return c.genfile
}

func (c *context) Execute(ctx corecontext.Context, generators ...Generator) error {
	cc, l := logr.FromContext(ctx).Start(ctx, "Gen")
	defer l.End()

	for pkgPath := range c.pkgPaths {
		if err := c.pkgExecute(cc, pkgPath, generators...); err != nil {
			return err
		}
	}

	l.Info("all done.")
	return nil
}

func (c *context) pkgExecute(ctx corecontext.Context, pkg string, generators ...Generator) error {
	ctx, l := logr.FromContext(ctx).Start(ctx, pkg)
	defer l.End()

	p, ok := c.universe[pkg]
	if !ok {
		return fmt.Errorf("invalid pkg `%s`", pkg)
	}

	pkgCtx := &context{
		universe: c.universe,
		args:     c.args,
		pkg:      p,
		pkgTags:  map[string][]string{},
	}

	for _, f := range p.Files() {
		if f.Doc != nil && len(f.Doc.List) > 0 {
			tags, _ := gengotypes.ExtractCommentTags(strings.Split(f.Doc.Text(), "\n"))
			for k := range tags {
				pkgCtx.pkgTags[k] = tags[k]
			}
		}
	}

	gfs := sync.Map{}
	eg := &errgroup.Group{}

	for i := range generators {
		eg.Go(func() error {
			pkgCtxForGen := &context{
				args:     pkgCtx.args,
				universe: pkgCtx.universe,
				pkg:      pkgCtx.pkg,
				pkgTags:  pkgCtx.pkgTags,
				genfile:  newGenfile(),
			}

			if err := pkgCtxForGen.genfile.InitWith(pkgCtxForGen); err != nil {
				return err
			}

			g := pkgCtxForGen.New(generators[i])

			pkgCtxForGen.genfile.generator = g
			pkgCtxForGen.l = logr.FromContext(ctx).WithValues("gengo", g.Name())

			if e := pkgCtxForGen.doGenerate(ctx, g); e != nil {
				return fmt.Errorf("`%s` generate failed for %s: %w", g.Name(), pkgCtx.pkg.Pkg().Path(), e)
			}

			gfs.Store(g.Name(), pkgCtxForGen.genfile)

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	for _, w := range gfs.Range {
		if err := w.(*genfile).WriteToFile(pkgCtx, c.args); err != nil {
			return err
		}
	}

	return nil
}

func (c *context) Package(importPath string) gengotypes.Package {
	if importPath == "" {
		return c.pkg
	}
	return c.universe.Package(importPath)
}

func (c *context) LocateInPackage(pos token.Pos) gengotypes.Package {
	return c.universe.LocateInPackage(pos)
}

func (c *context) Doc(typ types.Object) (Tags, []string) {
	tags, doc := c.universe.Package(typ.Pkg().Path()).Doc(typ.Pos())
	return merge(c.args.Globals, c.pkgTags, tags), doc
}

func (c *context) doGenerate(ctx corecontext.Context, g Generator) error {
	if c.pkg == nil {
		return nil
	}

	pkgTypes := c.pkg.Types()

	names := make([]string, 0)

	for n := range pkgTypes {
		names = append(names, n)
	}

	sort.Strings(names)

	for _, n := range names {
		tpe := pkgTypes[n].Type()

		if named, ok := tpe.(*types.Named); ok {
			// 	skip alias other pkg type XXX = XXX2
			if named.Obj().Pkg() != c.pkg.Pkg() {
				continue
			}

			tags, _ := c.Doc(named.Obj())

			if IsGeneratorEnabled(g, tags) {
				if err := func() error {
					_, l := logr.FromContext(ctx).WithValues(slog.String("target", named.String())).Start(ctx, g.Name())
					defer l.End()

					if err := g.GenerateType(c, named); err != nil {
						if errors.Is(err, ErrSkip) {
							return nil
						}
						if errors.Is(err, ErrIgnore) {
							l.Warn(err)
							return nil
						}
						return err
					}

					l.Debug("generated.")

					return nil
				}(); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *context) New(generator Generator) Generator {
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
