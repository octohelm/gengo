package defaultergen

import (
	"go/types"

	"github.com/octohelm/gengo/pkg/gengo"
)

func init() {
	gengo.Register(&defaulterGen{})
}

type defaulterGen struct {
	gengo.SnippetWriter
}

func (*defaulterGen) Name() string {
	return "defaulter"
}

func (*defaulterGen) New(c gengo.Context) gengo.Generator {
	return &defaulterGen{
		SnippetWriter: c.Writer(),
	}
}

func (g *defaulterGen) GenerateType(c gengo.Context, t *types.Named) error {
	g.Render(gengo.Snippet{gengo.T: `
func(v *@Type) SetDefault() {
	// TODO
}
`,
		"Type": gengo.ID(t.Obj()),
	})
	return nil
}
