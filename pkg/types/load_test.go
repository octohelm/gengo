package types

import (
	"go/types"
	"os"
	"testing"

	gopackages "golang.org/x/tools/go/packages"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"
)

func TestLoad(t *testing.T) {
	t.Run("GIVEN 包路径列表", func(t *testing.T) {
		patterns := []string{
			"github.com/octohelm/gengo/testdata/a",
		}

		t.Run("WHEN 加载包", func(t *testing.T) {
			u := MustValue(t, func() (*Universe, error) {
				return Load(patterns, func(c *gopackages.Config) {
					c.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
				})
			})

			t.Run("THEN 获取指定包", func(t *testing.T) {
				p := u.Package("github.com/octohelm/gengo/testdata/a")

				Then(t, "包应该存在",
					Expect(p, Be(cmp.NotNil[Package]())),
				)

				t.Run("注释相关测试", func(t *testing.T) {
					t.Run("常量注释", func(t *testing.T) {
						c := p.Constant("GENDER__MALE")

						Then(t, "常量应该存在",
							Expect(c, Be(cmp.NotNil[*types.Const]())),
						)

						comments := p.Comment(c.Pos())

						Then(t, "注释应该正确",
							Expect(comments, Equal([]string{"男"})),
						)
					})

					t.Run("结构体注释", func(t *testing.T) {
						tpe := p.Type("Struct")

						Then(t, "类型应该存在",
							Expect(tpe, Be(cmp.NotNil[*types.TypeName]())),
						)

						_, lines := p.Doc(tpe.Pos())

						Then(t, "结构体文档应该正确",
							Expect(lines, Equal([]string{"Struct"})),
						)

						s := tpe.Type().(*types.Named).Underlying().(*types.Struct)

						Then(t, "结构体字段数量正确",
							Expect(s.NumFields(), Be(cmp.Gt(0))),
						)

						t.Run("检查各个字段", func(t *testing.T) {
							foundID := false
							foundSlice := false

							for f := range s.Fields() {
								t.Run(f.Name(), func(t *testing.T) {
									if f.Name() == "ID" {
										foundID = true
										_, lines := p.Doc(f.Pos())

										Then(t, "ID字段文档应该正确",
											Expect(lines, Equal([]string{"StructID"})),
										)
									}

									if f.Name() == "Slice" {
										foundSlice = true
										_, lines := p.Doc(f.Pos())

										Then(t, "Slice 字段应该没有文档",
											Expect(len(lines), Equal(0)),
										)
									}
								})
							}

							Then(t, "应该找到 ID 字段",
								Expect(foundID, Be(cmp.True())),
							)

							Then(t, "应该找到 Slice 字段",
								Expect(foundSlice, Be(cmp.True())),
							)
						})
					})
				})

				t.Run("方法测试", func(t *testing.T) {
					tpe := p.Type("FakeBool")

					Then(t, "类型应该存在",
						Expect(tpe, Be(cmp.NotNil[*types.TypeName]())),
					)

					namedType := tpe.Type().(*types.Named)

					Then(t, "不包括嵌入式方法时",
						Expect(
							len(p.MethodsOf(namedType, false)),
							Be(cmp.Eq(1)),
						),
					)

					Then(t, "包括嵌入式方法时",
						Expect(
							len(p.MethodsOf(namedType, true)),
							Be(cmp.Eq(1)),
						),
					)
				})

				t.Run("函数结果测试", func(t *testing.T) {
					funcResults := map[string]string{
						"FuncWithFuncArg":                   `(*errors.errorString)`,
						"FuncReturnWithInterfaceCallSingle": `(string)`,
						"FuncReturnWithInterfaceCall":       `(string, error)`,
						"FuncWithCallChain":                 `(untyped nil | *string, untyped nil | untyped nil)`,
						"FuncSingleReturn":                  `(2)`,
						"FuncSelectExprReturn":              `(string | "2")`,
						"FuncWillCall":                      `(2, github.com/octohelm/gengo/testdata/a.String)`,
						"FuncReturnWithCallDirectly":        `(2, github.com/octohelm/gengo/testdata/a.String)`,
						"FuncWithNamedReturn":               `(2, github.com/octohelm/gengo/testdata/a.String)`,
						"FuncSingleNamedReturnByAssign":     `("1", "2", *errors.errorString)`,
						"FuncWithSwitch":                    `("a1" | "a2" | "a3", "b1" | "b2" | "b3")`,
						"FuncWithIf":                        `("a0" | "a1" | string)`,
						"FuncCallReturnAssign":              `(2, github.com/octohelm/gengo/testdata/a.String)`,
						"FuncCallWithFuncLit":               `(1, "s")`,
						"FuncWithImportedCall":              `(int)`,
						"FuncCurryCall":                     `(int)`,
						"FuncWithGenerics":                  `(*github.com/octohelm/gengo/testdata/a.Node, untyped nil)`,
					}

					for funcName, expectedResult := range funcResults {
						t.Run(funcName, func(t *testing.T) {
							fn := p.Function(funcName)

							Then(t, "函数应该存在",
								Expect(fn, Be(cmp.NotNil[*types.Func]())),
							)

							ar, _ := p.ResultsOf(fn)

							Then(t, "结果字符串应该匹配",
								Expect(ar.String(), Equal(expectedResult)),
							)
						})
					}
				})
			})
		})
	})
}
