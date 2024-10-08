package deepcopygen

import (
	"go/types"

	"github.com/octohelm/gengo/devpkg/deepcopygen/helper"

	"github.com/octohelm/gengo/pkg/gengo"
)

func init() {
	gengo.Register(&deepcopyGen{})
}

func (*deepcopyGen) Name() string {
	return "deepcopy"
}

type deepcopyGen struct {
	processed map[*types.Named]bool
}

func (g *deepcopyGen) GenerateType(c gengo.Context, named *types.Named) error {
	if g.processed == nil {
		g.processed = map[*types.Named]bool{}
	}

	return g.generateType(c, named)
}

func (g *deepcopyGen) generateType(c gengo.Context, named *types.Named) error {
	if _, ok := g.processed[named]; ok {
		return nil
	}

	g.processed[named] = true

	if _, ok := named.Obj().Type().Underlying().(*types.Interface); ok {
		return gengo.ErrSkip
	}

	tags, _ := c.Doc(named.Obj())
	if !gengo.IsGeneratorEnabled(g, tags) {
		return nil
	}

	interfaces := ""

	if tn, ok := tags["gengo:deepcopy:interfaces"]; ok {
		if n := tn[0]; len(n) > 0 {
			interfaces = n
		}
	}

	defers := make([]*types.Named, 0)

	if interfaces != "" {
		c.Render(gengo.Snippet{gengo.T: `
func(in *@Type) DeepCopyObject() @ObjectInterface {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil 
}

`,
			"ObjectInterface": gengo.ID(interfaces),
			"Type":            gengo.ID(named.Obj()),
		})
	}

	switch x := named.Underlying().(type) {
	case *types.Interface:

	case *types.Map:
		c.Render(gengo.Snippet{gengo.T: `
func(in @Type) DeepCopy() @Type {
	if in == nil {
		return nil
	}
	out := make(@Type)
	in.DeepCopyInto(out)
	return out
}

func(in @Type) DeepCopyInto(out @Type) {
	for k := range in {
		out[k] = in[k]
	}
}

`,
			"Type": gengo.ID(named.Obj()),
		})

	case *types.Struct:
		c.Render(gengo.Snippet{gengo.T: `
func(in *@Type) DeepCopy() *@Type {
	if in == nil {
		return nil
	}
	out := new(@Type)
	in.DeepCopyInto(out)
	return out
}

func(in *@Type) DeepCopyInto(out *@Type) {
	@fieldsCopies
}
`,
			"Type": gengo.ID(named.Obj()),
			"fieldsCopies": (&helper.StructFieldsCopy{
				Struct:           x,
				DeepCopyIntoName: "DeepCopyInto",
				DeepCopyName:     "DeepCopy",
				OnLocalDep: func(named *types.Named) {
					defers = append(defers, named)
				},
			}).Snippet(c, named.Obj().Pkg()),
		})
	default:
		c.Render(gengo.Snippet{gengo.T: `
func(in *@Type) DeepCopy() *@Type {
	if in == nil {
		return nil
	}
	
	out := new(@Type)
	in.DeepCopyInto(out)
	return out
}

func (in *@Type) DeepCopyInto(out *@Type) {
	*out = *in
}
`,
			"Type": gengo.ID(named.Obj()),
		})
	}

	for i := range defers {
		if err := g.generateType(c, defers[i]); err != nil {
			return err
		}
	}

	return nil
}
