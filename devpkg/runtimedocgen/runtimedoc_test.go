package runtimedocgen

import (
	"testing"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/gengo/pkg/gengo/testingutil"
)

func TestRuntimeDocGen(t *testing.T) {
	m := MustValue(t, func() (*testingutil.Module, error) {
		return testingutil.NewModule(t, map[string]string{
			"sample/doc/article.md": "article detail",
			"sample/types.go": `package sample

// Article
// [[doc/article.md]]
// +gengo:runtimedoc
type Article struct {
	// Title
	Title string

	// DetailsPrefix
	Details Details

	Embedded

	empty Empty
}

// Details
// +gengo:runtimedoc
type Details struct {
	// Summary
	Summary string
}

// Embedded
// +gengo:runtimedoc
type Embedded struct {
	// Code
	Code string
}

// Empty
// +gengo:runtimedoc
type Empty struct{}

// hidden
// +gengo:runtimedoc
type hidden struct {
	Name string
}

// Reader
// +gengo:runtimedoc
type Reader interface {
	Read()
}
`,
		})
	})

	files := MustValue(t, func() (map[string]string, error) {
		return m.Generate(gengo.GeneratorArgs{
			Entrypoint:         []string{m.ImportPath("sample")},
			OutputFileBaseName: "zz_generated_test",
			Force:              true,
		}, &runtimedocGen{})
	})

	Then(t, "应生成 runtime doc 方法和 helper",
		Expect(files, Be(testingutil.File("sample/zz_generated_test.runtimedoc.go",
			testingutil.Contains(
				"//go:embed doc/article.md",
				"func (v *Article) RuntimeDoc(names ...string) ([]string, bool)",
				"case \"Title\":",
				"return []string{}, true",
				"case \"Details\":",
				"\"Prefix\"",
				"if doc, ok := runtimeDoc(&v.Embedded, \"\", names...); ok",
				"func (v *Details) RuntimeDoc(names ...string) ([]string, bool)",
				"func (v *Embedded) RuntimeDoc(names ...string) ([]string, bool)",
			),
			testingutil.Count("func runtimeDoc(", cmp.Eq(1)),
			testingutil.NotContains(
				"func (v *hidden) RuntimeDoc",
				"func (v *Reader) RuntimeDoc",
			),
		))),
	)
}
