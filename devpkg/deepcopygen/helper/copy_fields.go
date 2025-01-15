package helper

import (
	"context"
	"go/types"
	"iter"

	"github.com/octohelm/gengo/pkg/gengo/snippet"
)

type FieldContext struct {
	InSamePkg        bool
	HasDeepCopy      bool
	HasDeepCopyInto  bool
	PtrResultOrParam bool
}

type StructFieldsCopy struct {
	Pkg              *types.Package
	Struct           *types.Struct
	DeepCopyName     string
	DeepCopyIntoName string
	Skip             func(f *types.Var) bool
	OnLocalDep       func(named *types.Named)
	FieldContext     func(f *types.Var) *FieldContext
}

func (sfc *StructFieldsCopy) IsNil() bool {
	return false
}

func (sfc *StructFieldsCopy) Frag(ctx context.Context) iter.Seq[string] {
	if sfc.DeepCopyName == "" {
		sfc.DeepCopyName = "DeepCopy"
	}

	if sfc.DeepCopyIntoName == "" {
		sfc.DeepCopyIntoName = "DeepCopyInto"
	}

	return func(yield func(string) bool) {
		for i := 0; i < sfc.Struct.NumFields(); i++ {
			f := sfc.Struct.Field(i)

			if sfc.Skip != nil && sfc.Skip(f) {
				continue
			}

			for code := range snippet.Fragments(ctx, sfc.createFieldSnippet(f)) {
				if !yield(code) {
					return
				}
			}
		}
	}
}

func (sfc *StructFieldsCopy) createFieldSnippet(f *types.Var) snippet.Snippet {
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

			fc.InSamePkg = x.Obj().Pkg().Path() == sfc.Pkg.Path()
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
			return snippet.T(`
in.@fieldName.@DeepCopyInto(&out.@fieldName)
`, snippet.Args{
				"fieldName":    snippet.ID(f.Name()),
				"DeepCopyInto": snippet.ID(sfc.DeepCopyIntoName),
			})
		} else if !fc.PtrResultOrParam && fc.HasDeepCopy {
			return snippet.T(`out.@fieldName = in.@fieldName.@DeepCopy()
`, snippet.Args{
				"fieldName": snippet.ID(f.Name()),
				"DeepCopy":  snippet.ID(sfc.DeepCopyName),
			})
		} else if fc.PtrResultOrParam && fc.HasDeepCopy {
			return snippet.T(`out.@fieldName = *in.@fieldName.@DeepCopy()
`, snippet.Args{
				"fieldName": snippet.ID(f.Name()),
				"DeepCopy":  snippet.ID(sfc.DeepCopyName),
			})
		} else {
			return snippet.T(`out.@fieldName = in.@fieldName
`, snippet.Args{
				"fieldName": snippet.ID(f.Name()),
			})
		}
	case *types.Map:
		return snippet.T(`
if in.@fieldName != nil {
	i, o := &in.@fieldName, &out.@fieldName 
	*o = make(@MapType, len(*i))
	for key, val := range *i {
		(*o)[key] = val
	}
}
`, snippet.Args{
			"MapType":   snippet.ID(x),
			"fieldName": snippet.ID(f.Name()),
		})
	case *types.Slice:
		return snippet.T(`
if in.@fieldName != nil {
	i, o := &in.@fieldName, &out.@fieldName 
	*o = make(@SliceType, len(*i))
	copy(*o, *i)
}
`, snippet.Args{
			"SliceType": snippet.ID(x),
			"fieldName": snippet.ID(f.Name()),
		})
	default:
		return snippet.T(`
out.@fieldName = in.@fieldName
`, snippet.Args{
			"fieldName": snippet.ID(f.Name()),
		})
	}
}
