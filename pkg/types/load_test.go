package types

import (
	"go/types"
	"testing"

	"github.com/onsi/gomega"
)

func TestLoad(t *testing.T) {
	u, _, err := Load([]string{
		"github.com/octohelm/gengo/testdata/a",
	})

	gomega.NewWithT(t).Expect(err).To(gomega.BeNil())

	p := u.Package("github.com/octohelm/gengo/testdata/a")

	t.Run("Comments", func(t *testing.T) {
		tpe := p.Type("Struct")
		_, lines := p.Doc(tpe.Pos())
		gomega.NewWithT(t).Expect(lines).To(gomega.Equal([]string{
			"Struct",
		}))

		s := tpe.Type().(*types.Named).Underlying().(*types.Struct)

		for i := 0; i < s.NumFields(); i++ {
			f := s.Field(i)

			if f.Name() == "ID" {
				_, lines := p.Doc(f.Pos())
				gomega.NewWithT(t).Expect(lines).To(gomega.Equal([]string{
					"StructID",
				}))
			}

			if f.Name() == "Slice" {
				_, lines := p.Doc(f.Pos())
				gomega.NewWithT(t).Expect(lines).To(gomega.BeNil())
			}
		}
	})

	tpe := p.Type("FakeBool")
	gomega.NewWithT(t).Expect(p.MethodsOf(tpe.Type().(*types.Named), false)).To(gomega.HaveLen(1))
	gomega.NewWithT(t).Expect(p.MethodsOf(tpe.Type().(*types.Named), true)).To(gomega.HaveLen(1))

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
				gomega.NewWithT(t).Expect(ar.String()).To(gomega.Equal(r))
			})
		}
	})
}
