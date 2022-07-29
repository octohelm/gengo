package deepcopygen

import (
	"go/types"

	"github.com/octohelm/gengo/pkg/gengo"
)

func init() {
	gengo.Register(&deepcopyGen{})
}

type deepcopyGen struct {
	gengo.SnippetWriter
	processed map[*types.Named]bool
}

func (*deepcopyGen) Name() string {
	return "deepcopy"
}

func (*deepcopyGen) New(c gengo.Context) gengo.Generator {
	return &deepcopyGen{
		SnippetWriter: c.Writer(),
		processed:     map[*types.Named]bool{},
	}
}

func (g *deepcopyGen) GenerateType(c gengo.Context, named *types.Named) error {
	return g.generateType(c, named)
}

func (g *deepcopyGen) generateType(c gengo.Context, named *types.Named) error {
	if _, ok := g.processed[named]; ok {
		return nil
	}

	g.processed[named] = true

	tags, _ := c.Doc(named.Obj())

	interfaces := ""

	if tn, ok := tags["gengo:deepcopy:interfaces"]; ok {
		if n := tn[0]; len(n) > 0 {
			interfaces = n
		}
	}

	defers := make([]*types.Named, 0)

	if interfaces != "" {
		g.Do(`
func(in *@Type) DeepCopyObject() @ObjectInterface {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil 
}

`, gengo.Args{
			"ObjectInterface": gengo.ID(interfaces),
			"Type":            gengo.ID(named.Obj()),
		})
	}

	switch x := named.Underlying().(type) {
	case *types.Map:
		g.Do(`
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

`, gengo.Args{
			"Type": gengo.ID(named.Obj()),
		})

	case *types.Struct:
		g.Do(`
func(in *@Type) DeepCopy() *@Type {
	if in == nil {
		return nil
	}
	out := new(@Type)
	in.DeepCopyInto(out)
	return out
}

func(in *@Type) DeepCopyInto(out *@Type) {
	@fieldCopies
}
`, gengo.Args{
			"Type": gengo.ID(named.Obj()),
			"fieldCopies": func(sw gengo.SnippetWriter) {
				for i := 0; i < x.NumFields(); i++ {
					f := x.Field(i)

					ft := f.Type()

					switch x := ft.(type) {
					case *types.Named:
						inSamePkg := x.Obj().Pkg().Path() == named.Obj().Pkg().Path()
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

							// always gen
							hasDeepCopyInto = true
							hasDeepCopy = true
						}

						if ptr && hasDeepCopyInto {
							g.Do(`in.@fieldName.DeepCopyInto(&out.@fieldName)
`, gengo.Args{
								"fieldName": gengo.ID(f.Name()),
							})
						} else if !ptr && hasDeepCopy {
							g.Do(`out.@fieldName = in.@fieldName.DeepCopy()
`, gengo.Args{
								"fieldName": gengo.ID(f.Name()),
							})
						} else if ptr && hasDeepCopy {
							g.Do(`out.@fieldName = *in.@fieldName.DeepCopy()
`, gengo.Args{
								"fieldName": gengo.ID(f.Name()),
							})
						} else {
							g.Do(`out.@fieldName = in.@fieldName
`, gengo.Args{
								"fieldName": gengo.ID(f.Name()),
							})
						}
					case *types.Map:
						g.Do(`
if in.@fieldName != nil {
	i, o := &in.@fieldName, &out.@fieldName 
	*o = make(@MapType, len(*i))
	for key, val := range *i {
		(*o)[key] = val
	}
}
`, gengo.Args{
							"MapType":   gengo.ID(x),
							"fieldName": gengo.ID(f.Name()),
						})
					case *types.Slice:
						g.Do(`
if in.@fieldName != nil {
	i, o := &in.@fieldName, &out.@fieldName 
	*o = make(@SliceType, len(*i))
	copy(*o, *i)
}
`, gengo.Args{
							"SliceType": gengo.ID(x),
							"fieldName": gengo.ID(f.Name()),
						})
					default:
						g.Do(`
out.@fieldName = in.@fieldName
`, gengo.Args{
							"fieldName": gengo.ID(f.Name()),
						})
					}
				}
			},
		})
	default:
		g.Do(`
func(in *@Type) DeepCopy() *@Type {
	if in == nil {
		return nil
	}
	
	out := new(@Type)
	in.DeepCopyInto(out)
	return out
}

`, gengo.Args{
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
