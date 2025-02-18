package defaultergen

import (
	"context"
	"fmt"
	"go/ast"
	"go/types"
	"iter"
	"strings"

	"github.com/octohelm/gengo/devpkg/deepcopygen/helper"
	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/gengo/pkg/gengo/snippet"
)

func init() {
	gengo.Register(&partialStructGen{})
}

type partialStructGen struct{}

func (*partialStructGen) Name() string {
	return "partialstruct"
}

func (g *partialStructGen) GenerateType(c gengo.Context, named *types.Named) error {
	tags, _ := c.Doc(named.Obj())
	if !gengo.IsGeneratorEnabled(g, tags) {
		return nil
	}

	name := named.Obj().Name()

	ps := PartialStruct{
		Name: fmt.Sprintf("%s%s", strings.ToUpper(name[0:1]), name[1:]),

		Replace: map[string][]string{},
		Omit:    map[string]bool{},
	}

	if exclude, ok := tags["gengo:partialstruct:omit"]; ok {
		for _, field := range exclude {
			ps.Omit[field] = true
		}
	}

	if replace, ok := tags["gengo:partialstruct:replace"]; ok {
		for _, field := range replace {
			parts := strings.SplitN(field, ":", 2)
			if len(parts) == 2 {
				ps.Replace[parts[0]] = strings.Split(parts[1], " ")
			}
		}
	}

	underlying := named.Underlying()

	ts, ok := underlying.(*types.Struct)
	if !ok {
		return fmt.Errorf("must be struct type, but got %s", underlying)
	}

	pkg := c.Package(named.Obj().Pkg().Path())
	decl := pkg.Decl(named.Obj().Pos())

	if d, ok := decl.(*ast.GenDecl); ok {
		for _, spec := range d.Specs {
			switch x := spec.(type) {
			case *ast.TypeSpec:
				switch x := x.Type.(type) {
				case *ast.Ident:
					switch x := pkg.ObjectOf(x).(type) {
					case *types.TypeName:
						ps.Origin = x
					}
				case *ast.SelectorExpr:
					switch x := pkg.ObjectOf(x.Sel).(type) {
					case *types.TypeName:
						ps.Origin = x
					}
				}
			}
		}
	}

	if ps.Origin == nil {
		return fmt.Errorf("need to define type like `type xxx sourcepkg.Type`")
	}

	return ps.generate(c, named, ts)
}

type PartialStruct struct {
	Name    string
	Origin  *types.TypeName
	Omit    map[string]bool
	Replace map[string][]string
}

func (ps *PartialStruct) generate(c gengo.Context, named *types.Named, x *types.Struct) error {
	c.RenderT(`
type @Type struct {
	@fields
}

func (in *@Type) DeepCopyAs() *@OriginType {
	if in == nil {
		return nil
	}
	out := new(@OriginType)
	in.DeepCopyIntoAs(out)
	return out
}

func (in *@Type) DeepCopyIntoAs(out *@OriginType)  {
	@fieldsCopies
}
`, snippet.Args{
		"Type":       snippet.ID(ps.Name),
		"OriginType": snippet.ID(ps.Origin),

		"fieldsCopies": &helper.StructFieldsCopy{
			Pkg:              named.Obj().Pkg(),
			DeepCopyIntoName: "DeepCopyIntoAs",
			DeepCopyName:     "DeepCopyAs",
			Struct:           x,
			Skip: func(f *types.Var) bool {
				return ps.Omit[f.Name()]
			},
			FieldContext: func(f *types.Var) *helper.FieldContext {
				if _, ok := ps.Replace[f.Name()]; ok {
					fc := &helper.FieldContext{
						PtrResultOrParam: true,
						HasDeepCopy:      true,
						HasDeepCopyInto:  true,
					}

					return fc
				}
				return nil
			},
		},

		"fields": snippet.Func(func(ctx context.Context) iter.Seq[string] {
			return func(yield func(string) bool) {
				for i := 0; i < x.NumFields(); i++ {
					f := x.Field(i)
					tag := x.Tag(i)
					fieldName := f.Name()

					if _, ok := ps.Omit[fieldName]; ok {
						continue
					}

					_, fieldDoc := c.Doc(f)

					if replaceTo, ok := ps.Replace[fieldName]; ok {
						if len(replaceTo) > 1 {
							tag = strings.Join(replaceTo[1:], " ")
						}

						s := snippet.T(`
@fieldDoc
@fieldName @fieldType `+"`"+`@fieldTag`+"`"+`
`, snippet.Args{
							"fieldDoc":  snippet.Comment(strings.Join(fieldDoc, "\n")),
							"fieldName": snippet.ID(fieldName),
							"fieldType": snippet.ID(replaceTo[0]),
							"fieldTag":  snippet.ID(tag),
						})

						for code := range snippet.Fragments(ctx, s) {
							if !yield(code) {
								return
							}
						}

						continue
					}

					s := snippet.T(`
@fieldDoc
@fieldName @fieldType `+"`"+`@fieldTag`+"`"+`
`, snippet.Args{
						"fieldDoc":  snippet.Comment(strings.Join(fieldDoc, "\n")),
						"fieldName": snippet.ID(fieldName),
						"fieldType": snippet.ID(f.Type()),
						"fieldTag":  snippet.ID(tag),
					})

					for code := range snippet.Fragments(ctx, s) {
						if !yield(code) {
							return
						}
					}
				}
			}
		}),
	})

	return nil
}
