package defaultergen

import (
	"go/types"

	"github.com/octohelm/gengo/pkg/gengo"
)

func init() {
	gengo.Register(&defaulterGen{})
}

type defaulterGen struct{}

func (*defaulterGen) Name() string {
	return "defaulter"
}

func (g *defaulterGen) GenerateType(c gengo.Context, t *types.Named) error {
	c.Render(gengo.Snippet{gengo.T: `
func(v *@Type) SetDefault() {
	// TODO
}
`,
		"Type": gengo.ID(t.Obj()),
	})
	return nil
}
