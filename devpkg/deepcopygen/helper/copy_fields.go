package helper

import (
	"go/types"

	"github.com/octohelm/gengo/pkg/gengo"
)

type FieldContext struct {
	InSamePkg        bool
	HasDeepCopy      bool
	HasDeepCopyInto  bool
	PtrResultOrParam bool
}

type StructFieldsCopy struct {
	DeepCopyName     string
	DeepCopyIntoName string
	Struct           *types.Struct
	Skip             func(f *types.Var) bool
	OnLocalDep       func(named *types.Named)
	FieldContext     func(f *types.Var) *FieldContext
}

func (sfc *StructFieldsCopy) Snippet(c gengo.Context, pkg *types.Package) func(sw gengo.SnippetWriter) {
	if sfc.DeepCopyName == "" {
		sfc.DeepCopyName = "DeepCopy"
	}

	if sfc.DeepCopyIntoName == "" {
		sfc.DeepCopyIntoName = "DeepCopyInto"
	}

	return func(sw gengo.SnippetWriter) {
		for i := 0; i < sfc.Struct.NumFields(); i++ {
			f := sfc.Struct.Field(i)

			if sfc.Skip != nil && sfc.Skip(f) {
				continue
			}

			fieldType := f.Type()

			switch x := fieldType.(type) {
			case *types.Named:
				var fc *FieldContext

				if sfc.FieldContext != nil {
					if fcc := sfc.FieldContext(f); fcc != nil {
						fc = fcc
					}
				}

				if fc == nil {
					fc = &FieldContext{}

					fc.InSamePkg = x.Obj().Pkg().Path() == pkg.Path()
					fc.PtrResultOrParam = true

					for i := 0; i < x.NumMethods(); i++ {
						m := x.Method(i)
						fn := m.Type().(*types.Signature)

						// FIXME better check
						fc.HasDeepCopy = m.Name() == sfc.DeepCopyName && fn.Results().Len() == 1 && fn.Params().Len() == 0
						fc.HasDeepCopyInto = m.Name() == sfc.DeepCopyIntoName && fn.Params().Len() == 1 && fn.Results().Len() == 0

						if fc.HasDeepCopy {
							if _, ok := fn.Results().At(0).Type().(*types.Pointer); !ok {
								fc.PtrResultOrParam = false
							}
						}

						if fc.HasDeepCopyInto {
							if _, ok := fn.Params().At(0).Type().(*types.Pointer); !ok {
								fc.PtrResultOrParam = false
							}
						}
					}
				}

				if fc.InSamePkg {
					if sfc.OnLocalDep != nil {
						sfc.OnLocalDep(x)
					}

					// always gen
					fc.HasDeepCopyInto = true
					fc.HasDeepCopy = true
				}

				if fc.PtrResultOrParam && fc.HasDeepCopyInto {
					c.Render(gengo.Snippet{gengo.T: `in.@fieldName.@DeepCopyInto(&out.@fieldName)
`,
						"fieldName":    gengo.ID(f.Name()),
						"DeepCopyInto": gengo.ID(sfc.DeepCopyIntoName),
					})
				} else if !fc.PtrResultOrParam && fc.HasDeepCopy {
					c.Render(gengo.Snippet{gengo.T: `out.@fieldName = in.@fieldName.@DeepCopy()
`,
						"fieldName": gengo.ID(f.Name()),
						"DeepCopy":  gengo.ID(sfc.DeepCopyName),
					})
				} else if fc.PtrResultOrParam && fc.HasDeepCopy {
					c.Render(gengo.Snippet{gengo.T: `out.@fieldName = *in.@fieldName.@DeepCopy()
`,
						"fieldName": gengo.ID(f.Name()),
						"DeepCopy":  gengo.ID(sfc.DeepCopyName),
					})
				} else {
					c.Render(gengo.Snippet{gengo.T: `out.@fieldName = in.@fieldName
`,
						"fieldName": gengo.ID(f.Name()),
					})
				}
			case *types.Map:
				c.Render(gengo.Snippet{gengo.T: `
if in.@fieldName != nil {
	i, o := &in.@fieldName, &out.@fieldName 
	*o = make(@MapType, len(*i))
	for key, val := range *i {
		(*o)[key] = val
	}
}
`,
					"MapType":   gengo.ID(x),
					"fieldName": gengo.ID(f.Name()),
				})
			case *types.Slice:
				c.Render(gengo.Snippet{gengo.T: `
if in.@fieldName != nil {
	i, o := &in.@fieldName, &out.@fieldName 
	*o = make(@SliceType, len(*i))
	copy(*o, *i)
}
`,
					"SliceType": gengo.ID(x),
					"fieldName": gengo.ID(f.Name()),
				})
			default:
				c.Render(gengo.Snippet{gengo.T: `
out.@fieldName = in.@fieldName
`,
					"fieldName": gengo.ID(f.Name()),
				})
			}
		}
	}
}
