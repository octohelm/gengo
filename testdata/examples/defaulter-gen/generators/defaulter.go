package generators

import (
	"go/types"

	"github.com/go-courier/gengo/pkg/gengo"
)

func init() {
	gengo.Register(&defaulterGen{})
}

type defaulterGen struct {
	gengo.SnippetWriter
}

func (defaulterGen) Name() string {
	return "defaulter"
}

func (*defaulterGen) New() gengo.Generator {
	return &defaulterGen{}
}

func (g *defaulterGen) Init(c *gengo.Context, s gengo.GeneratorCreator) (gengo.Generator, error) {
	return s.Init(c, g, func(g gengo.Generator, sw gengo.SnippetWriter) error {
		g.(*defaulterGen).SnippetWriter = sw
		return nil
	})
}

func (g defaulterGen) GenerateType(c *gengo.Context, t *types.Named) error {
	g.Do(`
func(v *[[ .type | raw ]]) SetDefault() {
}
`, gengo.Args{
		"type": t.Obj(),
	})
	return nil
}
