package defaultergen

import (
	"github.com/octohelm/gengo/pkg/gengo/snippet"
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
	c.RenderT(`
func(v *@Type) SetDefault() {
	// TODO
}
`,

		snippet.IDArg("Type", t.Obj()),
	)
	return nil
}
