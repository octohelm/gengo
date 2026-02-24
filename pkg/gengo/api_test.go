package gengo

import (
	"context"
	"errors"
	"go/types"
	"slices"
	"testing"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"
)

type testGenerator struct {
	name     string
	called   []string
	returned error
}

func (g *testGenerator) Name() string {
	return g.name
}

func (g *testGenerator) GenerateType(c Context, named *types.Named) error {
	g.called = append(g.called, named.Obj().Name())
	return g.returned
}

func (g *testGenerator) New(c Context) Generator {
	return g
}

func TestHelperAPI(t *testing.T) {
	t.Run("ImportGoPath", func(t *testing.T) {
		Then(t, "包含 vendor 前缀时应裁剪到 vendor 之后",
			Expect(ImportGoPath("github.com/example/project/vendor/github.com/acme/lib"), Equal("/vendor/github.com/acme/lib")),
		)

		Then(t, "无 vendor 前缀时应原样返回",
			Expect(ImportGoPath("github.com/example/project/pkg"), Equal("github.com/example/project/pkg")),
		)
	})

	t.Run("PkgImportPathAndExpose", func(t *testing.T) {
		pkg, expose := PkgImportPathAndExpose("github.com/acme/lib.Item")

		Then(t, "普通限定名应拆成包路径和类型名",
			Expect(pkg, Equal("github.com/acme/lib")),
			Expect(expose, Equal("Item")),
		)

		pkg, expose = PkgImportPathAndExpose("github.com/acme/lib.List[string]")

		Then(t, "泛型限定名应忽略类型参数",
			Expect(pkg, Equal("github.com/acme/lib")),
			Expect(expose, Equal("List")),
		)

		pkg, expose = PkgImportPathAndExpose("string")

		Then(t, "内置类型应返回空包路径",
			Expect(pkg, Equal("")),
			Expect(expose, Equal("string")),
		)
	})
}

func TestOptsAndTags(t *testing.T) {
	t.Run("Opts.Get 与 GetAll", func(t *testing.T) {
		opts := Opts{
			"output-file": {"a.go", "b.go"},
			"force":       {"true"},
		}

		v, ok := opts.Get("OutputFile")
		Then(t, "Get 应按 lower-kebab-case 取首个值",
			Expect(v, Equal("a.go")),
			Expect(ok, Be(cmp.True())),
		)

		all, ok := opts.GetAll("OutputFile")
		Then(t, "GetAll 应返回全部值",
			Expect(all, Equal([]string{"a.go", "b.go"})),
			Expect(ok, Be(cmp.True())),
		)

		v, ok = opts.Get("Missing")
		all, okAll := opts.GetAll("Missing")
		Then(t, "不存在键应返回 false",
			Expect(v, Equal("")),
			Expect(ok, Be(cmp.False())),
			Expect(all, Equal([]string(nil))),
			Expect(okAll, Be(cmp.False())),
		)
	})

	t.Run("IsGeneratorEnabled", func(t *testing.T) {
		g := &testGenerator{name: "demo"}

		Then(t, "显式 true 应启用生成器",
			Expect(IsGeneratorEnabled(g, map[string][]string{"gengo:demo": {""}}), Be(cmp.True())),
		)

		Then(t, "显式 false 应禁用生成器",
			Expect(IsGeneratorEnabled(g, map[string][]string{"gengo:demo": {"false"}}), Be(cmp.False())),
		)

		Then(t, "存在子选项时也应启用生成器",
			Expect(IsGeneratorEnabled(g, map[string][]string{"gengo:demo:output": {"x.go"}}), Be(cmp.True())),
		)

		Then(t, "无相关标签时应禁用生成器",
			Expect(IsGeneratorEnabled(g, map[string][]string{"gengo:other": {""}}), Be(cmp.False())),
		)
	})
}

func TestRegistryAPI(t *testing.T) {
	t.Run("Register 与 GetRegisteredGenerators", func(t *testing.T) {
		original := registeredGenerators
		registeredGenerators = map[string]Generator{}
		defer func() {
			registeredGenerators = original
		}()

		first := &testGenerator{name: "first"}
		second := &testGenerator{name: "second"}

		Register(first)
		Register(second)

		Then(t, "应返回全部已注册生成器",
			Expect(len(GetRegisteredGenerators()), Equal(2)),
		)

		selected := GetRegisteredGenerators("second", "missing")

		Then(t, "按名称查询时应仅返回存在项",
			Expect(len(selected), Equal(1)),
			Expect(selected[0].Name(), Equal("second")),
		)
	})
}

func TestContextAndExecute(t *testing.T) {
	t.Run("NewContext", func(t *testing.T) {
		t.Run("非法入口应返回错误", func(t *testing.T) {
			ExpectMust(func() error {
				_, err := NewContext(&GeneratorArgs{
					Entrypoint: []string{"/definitely/not/exist"},
				})
				if err == nil {
					return errors.New("expected error")
				}
				return nil
			})
		})

		t.Run("合法入口应创建执行器并可执行生成器", func(t *testing.T) {
			executor := MustValue(t, func() (Executor, error) {
				return NewContext(&GeneratorArgs{
					Entrypoint:         []string{"github.com/octohelm/gengo/testdata/a/b"},
					OutputFileBaseName: "zz_generated_api_test",
					Force:              true,
				})
			})

			Then(t, "应成功创建执行器",
				Expect(executor, Be(cmp.NotNil[Executor]())),
			)

			g := &testGenerator{name: "defaulter"}

			Must(t, func() error {
				return executor.Execute(context.Background(), g)
			})

			Then(t, "启用标签命中的包类型应被执行",
				Expect(len(g.called), Be(cmp.Gt(0))),
				ExpectMustValue(func() (bool, error) {
					if slices.Contains(g.called, "Obj") {
						return true, nil
					}
					return false, nil
				}, Be(cmp.True())),
			)
		})
	})
}
