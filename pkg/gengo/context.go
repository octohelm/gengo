package gengo

import (
	corecontext "context"
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

	"github.com/octohelm/x/logr"
	reflectx "github.com/octohelm/x/reflect"

	"github.com/octohelm/gengo/pkg/gengo/internal/cache"
	"github.com/octohelm/gengo/pkg/gengo/snippet"
	gengotypes "github.com/octohelm/gengo/pkg/types"
)

// Executor 会在 Context 已加载的包上执行生成器。
type Executor interface {
	Execute(ctx corecontext.Context, generators ...Generator) error
}

var noCached = sync.OnceValue(func() bool {
	return os.Getenv("GENGO_NO_CACHE") == "1"
})

// NewExecutor 加载配置的入口并返回一个执行器。
func NewExecutor(args *GeneratorArgs) (Executor, error) {
	if noCached() {
		args.NoCache = true
	}

	var genCache *cache.Cache

	if args.CacheDir != "" {
		genCache = cache.NewWithDir(args.CacheDir)
	} else {
		c, err := cache.New()
		if err != nil {
			return nil, err
		}
		genCache = c
	}

	u, err := gengotypes.Load(args.Entrypoint)
	if err != nil {
		return nil, err
	}

	c := &gengoCtx{
		universe:            u,
		args:                args,
		cache:               genCache,
		pkgContentHashCache: map[string]string{},
		l:                   newLogger(),
	}
	return c, nil
}

// NewContext 加载配置的入口并返回一个执行器。
//
// Deprecated: use NewExecutor.
//
//go:fix inline
func NewContext(args *GeneratorArgs) (Executor, error) {
	return NewExecutor(args)
}

// Context 在生成器执行期间暴露包元信息、指令和渲染辅助能力。
type Context interface {
	IsZero() bool

	Logger() logr.Logger
	Defer(func(c Context) error)

	LocateInPackage(pos token.Pos) gengotypes.Package
	Package(importPath string) gengotypes.Package
	Doc(typ types.Object) (Tags, []string)

	OptsOf(typ types.Object, generatorName string) Opts

	Render(snippet snippet.Snippet)
	RenderT(template string, args ...snippet.TArg)
}

// Opts 保存从注释标签解析并归一化后的生成器选项。
type Opts map[string][]string

// Get 返回 name 对应的第一个归一化选项值。
func (opts Opts) Get(name string) (string, bool) {
	if v, ok := opts[LowerKebabCase(name)]; ok {
		if len(v) > 0 {
			return v[0], true
		}
	}
	return "", false
}

// GetAll 返回 name 对应的全部归一化选项值。
func (opts Opts) GetAll(name string) ([]string, bool) {
	if v, ok := opts[LowerKebabCase(name)]; ok {
		return v, true
	}
	return nil, false
}

type gengoCtx struct {
	args     *GeneratorArgs
	universe *gengotypes.Universe

	pkgTags map[string][]string
	pkg     gengotypes.Package
	genfile *genfile

	ignore bool
	cache  *cache.Cache

	pkgContentHashCache map[string]string

	fullLoaded bool

	defers []func(ctx Context) error

	l logr.Logger
}

func (c *gengoCtx) OptsOf(typ types.Object, generatorName string) Opts {
	tags, _ := c.Doc(typ)
	if tags == nil {
		return Opts{}
	}
	opts := Opts{}

	prefix := "gengo:" + generatorName + ":"

	for tag, values := range tags {
		if strings.HasPrefix(tag, prefix) {
			opts[LowerKebabCase(tag[len(prefix):])] = values[:]
		}
	}

	return opts
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

func (c *gengoCtx) pkgContentHash(pkgPath string) string {
	cacheKey := pkgPath + "\x00" + c.args.OutputFileBaseName
	if h, ok := c.pkgContentHashCache[cacheKey]; ok {
		return h
	}
	h := computePkgContentHash(c.universe, pkgPath, c.args.OutputFileBaseName)
	c.pkgContentHashCache[cacheKey] = h
	return h
}

func (c *gengoCtx) Execute(ctx corecontext.Context, generators ...Generator) error {
	for pkgPath, direct := range c.universe.LocalPkgPaths() {
		if !direct {
			continue
		}

		if err := c.pkgExecute(logr.WithLogger(ctx, c.l), pkgPath, generators...); err != nil {
			return err
		}
	}

	return nil
}

func (c *gengoCtx) pkgExecute(pctx corecontext.Context, pkg string, generators ...Generator) (finalErr error) {
	p := c.universe.Package(pkg)
	if p == nil {
		return fmt.Errorf("invalid pkg `%s`", pkg)
	}

	// 先检测是否有需要执行的生成器
	regenerated := false

	ctx, l := logr.FromContext(pctx).Start(pctx, "generate", slog.String("scope", pkg))
	defer func() {
		if regenerated {
			l.End()
		} else {
			l.WithValues(slog.Bool("cached", true)).End()
		}
	}()

	defer func() {
		if finalErr != nil {
			l.Error(finalErr)
		}
	}()

	pkgContentHash := sync.OnceValue(func() string {
		return c.pkgContentHash(pkg)
	})

	p = c.universe.Package(pkg)

	generatedFiles := make(map[string]string)

	pkgCtx := &gengoCtx{
		universe:            c.universe,
		args:                c.args,
		pkg:                 p,
		pkgTags:             map[string][]string{},
		cache:               c.cache,
		pkgContentHashCache: c.pkgContentHashCache,
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
		genName := gen.Name()
		genVer := generatorVersion(gen)
		actionID := computeActionID(genName, genVer, pkgContentHash())

		pkgCtxForGen := &gengoCtx{
			args:                pkgCtx.args,
			universe:            pkgCtx.universe,
			pkg:                 pkgCtx.pkg,
			pkgTags:             pkgCtx.pkgTags,
			genfile:             newGenfile(genName),
			cache:               pkgCtx.cache,
			pkgContentHashCache: pkgCtx.pkgContentHashCache,
		}

		g := pkgCtxForGen.New(gen)

		// 检查缓存
		if !c.args.NoCache && c.cache.Exists(actionID) {
			gfs.Store(g.Name(), pkgCtxForGen.genfile)
			continue
		}

		if err := pkgCtxForGen.genfile.InitWith(pkgCtxForGen); err != nil {
			return err
		}

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

		// 写入缓存
		if err := c.cache.Mark(actionID); err != nil {
			l.Warn(fmt.Errorf("cache: mark failed: %w", err))
		}

		regenerated = true
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

// IsGeneratorEnabled 判断 tags 是否启用了 g。
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

// Tags 保存从包注释或声明注释解析出的指令。
type Tags map[string][]string
