package deepcopygen

import (
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/gengo/pkg/gengo/testingutil"
)

func TestDeepcopyGen(t *testing.T) {
	m := MustValue(t, func() (*testingutil.Module, error) {
		return testingutil.NewModule(t, map[string]string{
			"sample/types.go": `package sample

type Object interface {
	DeepCopyObject() Object
}

// +gengo:deepcopy
// +gengo:deepcopy:interfaces=Object
type Item struct {
	Name   string
	Child  Child
	Labels map[string]string
	Values []string
}

// +gengo:deepcopy
type Child struct {
	Count int
}

// +gengo:deepcopy
type Lookup map[string]string

// +gengo:deepcopy
type Size int64

// +gengo:deepcopy
type Skipped interface {
	Do()
}
`,
		})
	})

	files := MustValue(t, func() (map[string]string, error) {
		return m.Generate(gengo.GeneratorArgs{
			Entrypoint:         []string{m.ImportPath("sample")},
			OutputFileBaseName: "zz_generated_test",
			Force:              true,
		}, &deepcopyGen{})
	})

	Then(t, "应生成 deepcopy 方法",
		Expect(files, Be(testingutil.File("sample/zz_generated_test.deepcopy.go",
			testingutil.Contains(
				"func (in *Item) DeepCopyObject() Object",
				"func (in *Item) DeepCopy() *Item",
				"func (in *Item) DeepCopyInto(out *Item)",
				"in.Child.DeepCopyInto(&out.Child)",
				"*o = make(map[string]string, len(*i))",
				"*o = make([]string, len(*i))",
				"func (in *Child) DeepCopy() *Child",
				"func (in Lookup) DeepCopy() Lookup",
				"func (in *Size) DeepCopy() *Size",
			),
		))),
	)
}
