package deepcopygen

import (
	"go/types"

	"github.com/octohelm/gengo/pkg/gengo"
	typesx "github.com/octohelm/x/types"
)

func init() {
	gengo.Register(&deepcopyGen{})
}

type deepcopyGen struct {
	gengo.SnippetWriter
	processed map[*types.Named]bool
}

func (deepcopyGen) Name() string {
	return "deepcopy"
}

func (*deepcopyGen) New() gengo.Generator {
	return &deepcopyGen{}
}

func (g *deepcopyGen) Init(c *gengo.Context, s gengo.GeneratorCreator) (gengo.Generator, error) {
	g.processed = map[*types.Named]bool{}

	return s.Init(c, g, func(g gengo.Generator, sw gengo.SnippetWriter) error {
		g.(*deepcopyGen).SnippetWriter = sw
		return nil
	})
}

func (g *deepcopyGen) GenerateType(c *gengo.Context, t *types.Named) error {
	return g.generateType(c, t)
}

func (g *deepcopyGen) generateType(c *gengo.Context, t *types.Named) error {
	if _, ok := g.processed[t]; ok {
		return nil
	}

	tags, _ := c.Universe.Package(t.Obj().Pkg().Path()).Doc(t.Obj().Pos())

	interfaces := ""
	if tn, ok := tags["gengo:deepcopy:interfaces"]; ok {
		if n := tn[0]; len(n) > 0 {
			interfaces = n
		}
	}

	defers := make([]*types.Named, 0)

	if interfaces != "" {
		g.Do(`
func(in *[[ .type | id ]]) DeepCopyObject() [[ .interfaces | id ]] {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil 
}

`, gengo.Args{
			"interfaces": interfaces,
			"type":       t.Obj(),
		})
	}

	switch x := t.Underlying().(type) {
	case *types.Map:
		g.Do(`
func(in [[ .type | id ]]) DeepCopy() [[ .type | id ]] {
	if in == nil {
		return nil
	}
	out := make([[ .type | id ]])
	in.DeepCopyInto(out)
	return out
}

func(in [[ .type | id ]]) DeepCopyInto(out [[ .type | id ]] ) {
	for k := range in {
		out[k] = in[k]
    }
}

`, gengo.Args{"type": t.Obj()})

	case *types.Struct:
		g.Do(`
func(in *[[ .type | id ]]) DeepCopy() *[[ .type | id ]] {
	if in == nil {
		return nil
	}
	out := new([[ .type | id ]])
	in.DeepCopyInto(out)
	return out
}

func(in *[[ .type | id ]]) DeepCopyInto(out *[[ .type | id ]]) {
	[[ .fieldCopies | render ]]
}
`, gengo.Args{
			"type": t.Obj(),
			"fieldCopies": func(sw gengo.SnippetWriter) {
				for i := 0; i < x.NumFields(); i++ {
					f := x.Field(i)

					ft := f.Type()

					switch x := ft.(type) {
					case *types.Named:
						inSamePkg := x.Obj().Pkg().Path() == t.Obj().Pkg().Path()
						hasDeepCopy := false
						hasDeepCopyInto := false
						ptr := true

						for i := 0; i < x.NumMethods(); i++ {
							m := x.Method(i)
							fn := m.Type().(*types.Signature)

							// FIXME better check
							hasDeepCopy = m.Name() == "DeepCopy" && fn.Results().Len() == 1 && fn.Params().Len() == 0
							hasDeepCopyInto = m.Name() == "DeepCopyInto" && fn.Params().Len() == 1 && fn.Results().Len() == 0

							if hasDeepCopy {
								if _, ok := fn.Results().At(0).Type().(*types.Pointer); !ok {
									ptr = false
								}
							}

							if hasDeepCopyInto {
								if _, ok := fn.Params().At(0).Type().(*types.Pointer); !ok {
									ptr = false
								}
							}

						}

						if inSamePkg {
							defers = append(defers, x)
						}

						g.Do(`[[ 
if .canDeepCopyIntoPtr ]] in.[[ .fieldName ]].DeepCopyInto(&out.[[ .fieldName ]]) [[ 
else if .canDeepCopyNonPtr ]] out.[[ .fieldName ]] = in.[[ .fieldName ]].DeepCopy() [[ 
else if .canDeepCopyPtr ]] out.[[ .fieldName ]] = *in.[[ .fieldName ]].DeepCopy() [[ 
else ]] out.[[ .fieldName ]] = in.[[ .fieldName ]] [[ end ]]
`, gengo.Args{
							"fieldName":          f.Name(),
							"canDeepCopyIntoPtr": ptr && (hasDeepCopyInto || inSamePkg),
							"canDeepCopyNonPtr":  !ptr && (hasDeepCopy || inSamePkg),
							"canDeepCopyPtr":     ptr && (hasDeepCopy || inSamePkg),
						})
					case *types.Map:
						g.Do(`
if in.[[ .fieldName ]] != nil {
	i, o := &in.[[ .fieldName ]], &out.[[ .fieldName ]] 
	*o = make([[ .mapType ]], len(*i))
	for key, val := range *i {
		(*o)[key] = val
	}
}
`, gengo.Args{
							"mapType":   sw.Dumper().TypeLit(typesx.FromTType(x)),
							"fieldName": f.Name(),
						})
					case *types.Slice:
						g.Do(`
if in.[[ .fieldName ]] != nil {
	i, o := &in.[[ .fieldName ]], &out.[[ .fieldName ]] 
	*o = make([[ .sliceType ]], len(*i))
	copy(*o, *i)
}
`, gengo.Args{
							"sliceType": sw.Dumper().TypeLit(typesx.FromTType(x)),
							"fieldName": f.Name(),
						})
					default:
						g.Do(`
out.[[ .fieldName ]] = in.[[ .fieldName ]]
`, gengo.Args{"fieldName": f.Name()})
					}
				}
			},
		})
	default:
		g.Do(`
func(in *[[ .type | id ]]) DeepCopy() *[[ .type | id ]] {
	if in == nil {
		return nil
	}
	
	out := new([[ .type | id ]])
	in.DeepCopyInto(out)
	return out
}

`, gengo.Args{"type": t.Obj()})
	}

	for i := range defers {
		if err := g.generateType(c, defers[i]); err != nil {
			return err
		}
	}

	return nil
}
