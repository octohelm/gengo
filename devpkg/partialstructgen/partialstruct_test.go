package partialstructgen

import (
	"regexp"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/gengo/pkg/gengo/testingutil"
)

func TestPartialStructGen(t *testing.T) {
	m := MustValue(t, func() (*testingutil.Module, error) {
		return testingutil.NewModule(t, map[string]string{
			"origin/types.go": `package origin

type Origin struct {
	Keep  string ` + "`" + `json:"keep"` + "`" + `
	Count int    ` + "`" + `json:"count"` + "`" + `
	Skip  string ` + "`" + `json:"skip"` + "`" + `
}
`,
			"sample/types.go": `package sample

import "example.com/gengo-test/origin"

// +gengo:partialstruct
// +gengo:partialstruct:omit=Skip
// +gengo:partialstruct:replace=Count:*int json:"count,omitempty"
type OriginPatch origin.Origin
`,
		})
	})

	files := MustValue(t, func() (map[string]string, error) {
		return m.Generate(gengo.GeneratorArgs{
			Entrypoint:         []string{m.ImportPath("sample")},
			OutputFileBaseName: "zz_generated_test",
			Force:              true,
		}, &partialStructGen{})
	})

	Then(t, "应生成 partial struct 并应用字段选项",
		Expect(files, Be(testingutil.File("sample/zz_generated_test.partialstruct.go",
			testingutil.Contains(
				"type OriginPatch struct",
				"Keep string `json:\"keep\"`",
				"Count *int `json:\"count,omitempty\"`",
				"func (in *OriginPatch) DeepCopyAs() *origin.Origin",
				"func (in *OriginPatch) DeepCopyIntoAs(out *origin.Origin)",
				"out.Count = in.Count",
			),
			testingutil.NotContains("Skip"),
		))),
	)
}

func TestPartialStructGenErrors(t *testing.T) {
	t.Run("非 struct 类型应返回错误", func(t *testing.T) {
		m := MustValue(t, func() (*testingutil.Module, error) {
			return testingutil.NewModule(t, map[string]string{
				"sample/types.go": `package sample

// +gengo:partialstruct
type Bad string
`,
			})
		})

		Then(t, "应返回非 struct 错误",
			ExpectDo(func() error {
				return m.GenerateError(gengo.GeneratorArgs{
					Entrypoint:         []string{m.ImportPath("sample")},
					OutputFileBaseName: "zz_generated_test",
					Force:              true,
				}, &partialStructGen{})
			}, ErrorMatch(regexp.MustCompile("must be struct type"))),
		)
	})

	t.Run("缺少来源类型声明应返回错误", func(t *testing.T) {
		m := MustValue(t, func() (*testingutil.Module, error) {
			return testingutil.NewModule(t, map[string]string{
				"sample/types.go": `package sample

// +gengo:partialstruct
type Bad struct {
	Name string
}
`,
			})
		})

		Then(t, "应返回缺少来源类型声明错误",
			ExpectDo(func() error {
				return m.GenerateError(gengo.GeneratorArgs{
					Entrypoint:         []string{m.ImportPath("sample")},
					OutputFileBaseName: "zz_generated_test",
					Force:              true,
				}, &partialStructGen{})
			}, ErrorMatch(regexp.MustCompile("need to define type"))),
		)
	})
}
