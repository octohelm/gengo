package types

import (
	"go/ast"
	"go/constant"
	"go/parser"
	gotypes "go/types"
	"maps"
	"path/filepath"
	"testing"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/gengo/pkg/sumfile"
)

func TestHelpers(t *testing.T) {
	t.Run("StringifyNode", func(t *testing.T) {
		expr := MustValue(t, func() (ast.Expr, error) {
			return parser.ParseExpr(`pkg.Value(arg1, "x")`)
		})

		u := MustValue(t, func() (*Universe, error) {
			return Load([]string{"github.com/octohelm/gengo/testdata/a"})
		})

		p := u.Package("github.com/octohelm/gengo/testdata/a")

		Then(t, "表达式应被格式化回源码",
			Expect(StringifyNode(p.FileSet(), expr), Equal(`pkg.Value(arg1, "x")`)),
		)
	})

	t.Run("Result.String", func(t *testing.T) {
		t.Run("优先输出常量值", func(t *testing.T) {
			r := Result{
				Value: constant.MakeString("demo"),
				Type:  gotypes.Typ[gotypes.String],
			}

			Then(t, "应输出常量值",
				Expect(r.String(), Equal(`"demo"`)),
			)
		})

		t.Run("无常量值时输出类型", func(t *testing.T) {
			r := Result{
				Type: gotypes.Typ[gotypes.Int],
			}

			Then(t, "应输出类型名",
				Expect(r.String(), Equal("int")),
			)
		})

		t.Run("无值无类型时输出 invalid", func(t *testing.T) {
			Then(t, "应输出 invalid",
				Expect((Result{}).String(), Equal("invalid")),
			)
		})
	})

	t.Run("Results.String", func(t *testing.T) {
		results := Results{
			{Value: constant.MakeString("a")},
			{Type: gotypes.Typ[gotypes.Int]},
		}

		Then(t, "结果应按 | 串联",
			Expect(results.String(), Equal(`"a" | int`)),
		)
	})

	t.Run("FuncResults.String", func(t *testing.T) {
		results := FuncResults{
			{{Value: constant.MakeString("a")}},
			{{Type: gotypes.Typ[gotypes.Int]}},
		}

		Then(t, "返回位应按逗号串联",
			Expect(results.String(), Equal(`("a", int)`)),
		)
	})

	t.Run("FuncResults.Concat", func(t *testing.T) {
		t.Run("长度一致时应按位拼接", func(t *testing.T) {
			left := FuncResults{
				{{Value: constant.MakeInt64(1)}},
				{{Value: constant.MakeString("a")}},
			}
			right := FuncResults{
				{{Value: constant.MakeInt64(2)}},
				{{Value: constant.MakeString("b")}},
			}

			merged := left.Concat(right)

			Then(t, "应返回按位合并后的结果",
				Expect(merged.String(), Equal(`(1 | 2, "a" | "b")`)),
			)
		})

		t.Run("长度不一致时应保持原结果", func(t *testing.T) {
			left := FuncResults{
				{{Value: constant.MakeInt64(1)}},
			}
			right := FuncResults{
				{{Value: constant.MakeInt64(2)}},
				{{Value: constant.MakeString("b")}},
			}

			merged := left.Concat(right)

			Then(t, "应返回原始结果",
				Expect(merged.String(), Equal(`(1)`)),
			)
		})
	})
}

func TestPackageQueries(t *testing.T) {
	u := MustValue(t, func() (*Universe, error) {
		return Load([]string{"github.com/octohelm/gengo/testdata/a"})
	})

	p := u.Package("github.com/octohelm/gengo/testdata/a")

	t.Run("Constants 与 Constant", func(t *testing.T) {
		constants := p.Constants()

		Then(t, "应包含性别常量",
			Expect(constants["GENDER__MALE"], Be(cmp.NotNil[*gotypes.Const]())),
			Expect(constants["GENDER__FEMALE"], Be(cmp.NotNil[*gotypes.Const]())),
		)

		Then(t, "Constant 查询结果应与 map 中一致",
			Expect(p.Constant("GENDER__MALE"), Equal(constants["GENDER__MALE"])),
		)
	})

	t.Run("Types 与 Type", func(t *testing.T) {
		typeSet := p.Types()

		Then(t, "应包含公开类型",
			Expect(typeSet["Struct"], Be(cmp.NotNil[*gotypes.TypeName]())),
			Expect(typeSet["FakeBool"], Be(cmp.NotNil[*gotypes.TypeName]())),
		)

		Then(t, "Type 查询结果应与 map 中一致",
			Expect(p.Type("Struct"), Equal(typeSet["Struct"])),
		)
	})

	t.Run("Functions 与 Function", func(t *testing.T) {
		functions := p.Functions()

		Then(t, "应包含顶层函数",
			Expect(functions["FuncSingleReturn"], Be(cmp.NotNil[*gotypes.Func]())),
			Expect(functions["FuncWithIf"], Be(cmp.NotNil[*gotypes.Func]())),
		)

		Then(t, "Function 查询结果应与 map 中一致",
			Expect(p.Function("FuncWithIf"), Equal(functions["FuncWithIf"])),
		)
	})

	t.Run("Position 与 Decl", func(t *testing.T) {
		structType := p.Type("Struct")
		pos := p.Position(structType.Pos())

		Then(t, "位置信息应落在 a.go 中",
			Expect(filepath.Base(pos.Filename), Equal("a.go")),
			Expect(pos.Line, Be(cmp.Gt(0))),
		)

		decl := p.Decl(structType.Pos())

		Then(t, "声明应为类型声明",
			Expect(decl, Be(cmp.NotNil[ast.Decl]())),
			ExpectMustValue(func() (bool, error) {
				_, ok := decl.(*ast.GenDecl)
				return ok, nil
			}, Be(cmp.True())),
		)
	})

	t.Run("Eval 与 ObjectOf", func(t *testing.T) {
		var evalExpr ast.Expr
		var structIdent *ast.Ident

		for _, file := range p.Files() {
			ast.Inspect(file, func(node ast.Node) bool {
				switch x := node.(type) {
				case *ast.BinaryExpr:
					if evalExpr == nil {
						evalExpr = x
					}
				case *ast.TypeSpec:
					if x.Name.Name == "Struct" && structIdent == nil {
						structIdent = x.Name
					}
				}
				return evalExpr == nil || structIdent == nil
			})

			if evalExpr != nil && structIdent != nil {
				break
			}
		}

		Then(t, "应找到可求值表达式",
			Expect(evalExpr, Be(cmp.NotNil[ast.Expr]())),
		)

		tv := MustValue(t, func() (gotypes.TypeAndValue, error) {
			return p.Eval(evalExpr)
		})

		Then(t, "Eval 应解析出字符串类型和值",
			Expect(tv.Type.String(), Equal("untyped string")),
			Expect(tv.Value.ExactString(), Equal(`"1"`)),
		)

		Then(t, "应找到 Struct 标识符",
			Expect(structIdent, Be(cmp.NotNil[*ast.Ident]())),
		)

		obj := p.ObjectOf(structIdent)

		Then(t, "ObjectOf 应返回对应类型对象",
			Expect(obj, Be(cmp.NotNil[gotypes.Object]())),
			Expect(obj.Name(), Equal("Struct")),
			Expect(obj.Pkg().Path(), Equal("github.com/octohelm/gengo/testdata/a")),
		)
	})
}

func TestUniverseQueries(t *testing.T) {
	u := MustValue(t, func() (*Universe, error) {
		return Load([]string{"github.com/octohelm/gengo/testdata/a"})
	})

	p := u.Package("github.com/octohelm/gengo/testdata/a")

	t.Run("SumFile 与 LocalPkgPaths", func(t *testing.T) {
		Then(t, "应生成 sumfile",
			Expect(u.SumFile(), Be(cmp.NotNil[*sumfile.File]())),
		)

		localPkgPaths := maps.Collect(u.LocalPkgPaths())

		Then(t, "应包含直接加载包",
			Expect(localPkgPaths["github.com/octohelm/gengo/testdata/a"], Be(cmp.True())),
		)

		Then(t, "应包含同模块下的间接本地包",
			Expect(localPkgPaths["github.com/octohelm/gengo/testdata/a/b"], Be(cmp.False())),
			Expect(localPkgPaths["github.com/octohelm/gengo/testdata/a/x"], Be(cmp.False())),
		)
	})

	t.Run("LocateInPackage 与 SourceDir", func(t *testing.T) {
		structType := p.Type("Struct")

		Then(t, "SourceDir 应指向 testdata/a 目录",
			Expect(filepath.Base(p.SourceDir()), Equal("a")),
		)

		Then(t, "LocateInPackage 应定位到当前包",
			Expect(u.LocateInPackage(structType.Pos()), Equal(p)),
		)
	})

	t.Run("Imports", func(t *testing.T) {
		imports := p.Imports()

		Then(t, "应包含本模块导入包路径键",
			ExpectMustValue(func() (bool, error) {
				_, ok := imports["github.com/octohelm/gengo/testdata/a/b"]
				return ok, nil
			}, Be(cmp.True())),
			ExpectMustValue(func() (bool, error) {
				_, ok := imports["github.com/octohelm/gengo/testdata/a/x"]
				return ok, nil
			}, Be(cmp.True())),
			ExpectMustValue(func() (bool, error) {
				for _, imp := range imports {
					if imp == nil {
						return false, nil
					}
				}
				return true, nil
			}, Be(cmp.True())),
		)
	})
}

func TestCommentLinesFrom(t *testing.T) {
	t.Run("应跳过 go 指令并保留普通注释", func(t *testing.T) {
		commentGroup := &ast.CommentGroup{
			List: []*ast.Comment{
				{Text: "//go:generate test"},
				{Text: "// hello"},
				{Text: "// world"},
			},
		}

		Then(t, "只应保留普通注释行",
			Expect(commentLinesFrom(commentGroup), Equal([]string{"hello", "world"})),
		)
	})

	t.Run("空输入应返回 nil", func(t *testing.T) {
		Then(t, "无注释组时返回 nil",
			Expect(commentLinesFrom(), Be(cmp.Nil[[]string]())),
		)
	})
}
