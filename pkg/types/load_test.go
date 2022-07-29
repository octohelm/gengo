package types

import (
	"go/types"
	"testing"

	testingx "github.com/octohelm/x/testing"
)

func TestLoad(t *testing.T) {
	u, _, err := Load([]string{
		"github.com/octohelm/gengo/testdata/a",
	})

	testingx.Expect(t, err, testingx.Be[error](nil))

	p := u.Package("github.com/octohelm/gengo/testdata/a")

	t.Run("Comments", func(t *testing.T) {
		t.Run("Const", func(t *testing.T) {
			c := p.Constant("GENDER__MALE")
			comments := p.Comment(c.Pos())
			testingx.Expect(t, comments, testingx.Equal([]string{
				"男",
			}))
		})

		t.Run("Struct", func(t *testing.T) {
			tpe := p.Type("Struct")
			_, lines := p.Doc(tpe.Pos())
			testingx.Expect(t, lines, testingx.Equal([]string{
				"Struct",
			}))

			s := tpe.Type().(*types.Named).Underlying().(*types.Struct)

			for i := 0; i < s.NumFields(); i++ {
				f := s.Field(i)

				if f.Name() == "ID" {
					_, lines := p.Doc(f.Pos())
					testingx.Expect(t, lines, testingx.Equal([]string{
						"StructID",
					}))
				}

				if f.Name() == "Slice" {
					_, lines := p.Doc(f.Pos())
					testingx.Expect(t, len(lines), testingx.Be(0))
				}
			}
		})
	})

	tpe := p.Type("FakeBool")
	testingx.Expect(t, p.MethodsOf(tpe.Type().(*types.Named), false), testingx.HaveLen[[]*types.Func](1))
	testingx.Expect(t, p.MethodsOf(tpe.Type().(*types.Named), true), testingx.HaveLen[[]*types.Func](1))

	t.Run("ResultsOf", func(t *testing.T) {
		funcResults := map[string]string{
			"FuncSingleReturn":              `(2)`,
			"FuncSelectExprReturn":          `(string)`,
			"FuncWillCall":                  `(2, github.com/octohelm/gengo/testdata/a.String)`,
			"FuncReturnWithCallDirectly":    `(2, github.com/octohelm/gengo/testdata/a.String)`,
			"FuncWithNamedReturn":           `(2, github.com/octohelm/gengo/testdata/a.String)`,
			"FuncSingleNamedReturnByAssign": `("1", "2", *github.com/pkg/errors.fundamental)`,
			"FunWithSwitch":                 `("a1" | "a2" | "a3", "b1" | "b2" | "b3")`,
			"FuncWithIf":                    `("a0" | "a1" | string)`,
			"FuncCallReturnAssign":          `(2, github.com/octohelm/gengo/testdata/a.String)`,
			"FuncCallWithFuncLit":           `(1, "s")`,
			"FuncWithImportedCall":          `(int)`,
			"FuncCurryCall":                 `(int)`,
		}

		for k, r := range funcResults {
			t.Run(k, func(t *testing.T) {
				fn := p.Function(k)
				ar, _ := p.ResultsOf(fn)
				testingx.Expect(t, ar.String(), testingx.Equal(r))
			})
		}
	})
}
