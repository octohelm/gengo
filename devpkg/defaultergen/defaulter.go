package defaultergen

import (
	"go/types"

	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/gengo/pkg/gengo/snippet"
)

func init() {
	gengo.Register(&defaulterGen{})
}

type defaulterGen struct{}

func (*defaulterGen) Name() string {
	return "defaulter"
}

func (g *defaulterGen) GenerateType(c gengo.Context, t *types.Named) error {
	if _, isInterface := t.Obj().Type().(*types.Interface); isInterface {
		return gengo.ErrSkip
	}

	c.RenderT(`
func(v *@Type) SetDefault() {
	// TODO
}
`,

		snippet.IDArg("Type", t.Obj()),
	)

	return nil
}
