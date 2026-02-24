package gengo

import (
	"context"
	"errors"
	"go/types"
	"testing"

	"github.com/octohelm/x/cmp"
	"github.com/octohelm/x/logr"
	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/gengo/pkg/gengo/snippet"
)

type namedOnlyGenerator struct {
	name string
	err  error
}

func (g *namedOnlyGenerator) Name() string { return g.name }
func (g *namedOnlyGenerator) GenerateType(c Context, named *types.Named) error {
	return g.err
}

type aliasOnlyGenerator struct {
	name string
	err  error
}

func (g *aliasOnlyGenerator) Name() string { return g.name }
func (g *aliasOnlyGenerator) GenerateType(c Context, named *types.Named) error {
	return nil
}

func (g *aliasOnlyGenerator) GenerateAliasType(c Context, alias *types.Alias) error {
	return g.err
}

type newerGenerator struct{}

func (*newerGenerator) Name() string { return "newer" }
func (*newerGenerator) GenerateType(c Context, named *types.Named) error {
	return nil
}

func (*newerGenerator) New(c Context) Generator {
	return &namedOnlyGenerator{name: "newer"}
}

func mustCtxOnPackage(t *testing.T, entrypoint string) *gengoCtx {
	executor := MustValue(t, func() (Executor, error) {
		return NewContext(&GeneratorArgs{
			Entrypoint:         []string{entrypoint},
			OutputFileBaseName: "zz_generated_api_test",
			Force:              true,
		})
	})

	c := executor.(*gengoCtx)
	c.pkg = c.universe.Package(entrypoint)
	c.l = newLogger()

	c.genfile = newGenfile("demo")
	Must(t, func() error {
		return c.genfile.InitWith(c)
	})

	return c
}

func TestGengoCtxHelpers(t *testing.T) {
	c := mustCtxOnPackage(t, "github.com/octohelm/gengo/testdata/a/c")

	t.Run("Package 与 LocateInPackage", func(t *testing.T) {
		Then(t, "空 importPath 应返回当前包",
			Expect(c.Package(""), Equal(c.pkg)),
		)

		Then(t, "指定 importPath 应返回对应包",
			Expect(c.Package("github.com/octohelm/gengo/testdata/a/c"), Equal(c.pkg)),
		)

		kubeType := c.pkg.Type("KubePkg")
		Then(t, "LocateInPackage 应定位到所属包",
			Expect(c.LocateInPackage(kubeType.Pos()), Equal(c.pkg)),
		)
	})

	t.Run("Doc 与 OptsOf", func(t *testing.T) {
		kubeObj := c.pkg.Type("KubePkg")

		opts := c.OptsOf(kubeObj, "deepcopy")
		v, ok := opts.Get("interfaces")

		Then(t, "应提取到 deepcopy 的参数选项",
			Expect(v, Equal("Object")),
			Expect(ok, Be(cmp.True())),
		)

		tags, doc := c.Doc(kubeObj)
		Then(t, "应包含 gengo 标签",
			Expect(tags["gengo:deepcopy"], Equal([]string{""})),
			Expect(doc != nil, Be(cmp.True())),
		)
	})

	t.Run("Render 系列", func(t *testing.T) {
		Then(t, "Logger/Writer 应可用",
			Expect(c.Logger(), Be(cmp.NotNil[logr.Logger]())),
			Expect(c.Writer(), Be(cmp.NotNil[SnippetWriter]())),
		)

		c.Render(snippet.Block("const A = 1\n"))
		c.RenderT("const @name = @value\n", snippet.IDArg("name", "B"), snippet.ValueArg("value", 2))

		Then(t, "Render 输出应写入 genfile",
			Expect(c.genfile.body.String(), Be(func(s string) error {
				if s == "" {
					return errors.New("expected rendered output")
				}
				return nil
			})),
		)
	})

	t.Run("Defer", func(t *testing.T) {
		c.Defer(func(c Context) error { return nil })
		Then(t, "Defer 应记录回调",
			Expect(len(c.defers), Be(cmp.Gt(0))),
		)
	})

	t.Run("New", func(t *testing.T) {
		Then(t, "默认应按类型创建新实例",
			Expect(c.New(&namedOnlyGenerator{name: "x"}).Name(), Equal("")),
		)

		Then(t, "实现 GeneratorNewer 时应走自定义 New",
			Expect(c.New(&newerGenerator{}).Name(), Equal("newer")),
		)
	})
}

func TestGenerateBranches(t *testing.T) {
	t.Run("doGenerateNamedType", func(t *testing.T) {
		c := mustCtxOnPackage(t, "github.com/octohelm/gengo/testdata/a/c")
		named := c.pkg.Type("KubePkg").Type().(*types.Named)

		Must(t, func() error {
			return c.doGenerateNamedType(context.Background(), &namedOnlyGenerator{name: "x", err: ErrSkip}, named)
		})

		Must(t, func() error {
			return c.doGenerateNamedType(context.Background(), &namedOnlyGenerator{name: "x", err: ErrIgnore}, named)
		})
		Then(t, "ErrIgnore 分支应标记 ignore",
			Expect(c.ignore, Be(cmp.True())),
		)

		Then(t, "普通错误应向上返回",
			ExpectMust(func() error {
				err := c.doGenerateNamedType(context.Background(), &namedOnlyGenerator{name: "x", err: errors.New("boom")}, named)
				if err == nil {
					return errors.New("expected error")
				}
				return nil
			}),
		)
	})

	t.Run("doGenerateAliasType", func(t *testing.T) {
		c := mustCtxOnPackage(t, "github.com/octohelm/gengo/testdata/a")
		alias := c.pkg.Type("TimeAlias").Type().(*types.Alias)

		Must(t, func() error {
			return c.doGenerateAliasType(context.Background(), &aliasOnlyGenerator{name: "x", err: ErrSkip}, alias)
		})

		Must(t, func() error {
			return c.doGenerateAliasType(context.Background(), &aliasOnlyGenerator{name: "x", err: ErrIgnore}, alias)
		})

		Then(t, "普通错误应向上返回",
			ExpectMust(func() error {
				err := c.doGenerateAliasType(context.Background(), &aliasOnlyGenerator{name: "x", err: errors.New("boom")}, alias)
				if err == nil {
					return errors.New("expected error")
				}
				return nil
			}),
		)
	})
}
