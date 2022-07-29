package runtimedocgen

import (
	"github.com/octohelm/gengo/pkg/gengo"
	"go/types"
)

func init() {
	gengo.Register(&runtimedocGen{})
}

type runtimedocGen struct {
	gengo.SnippetWriter
	processed     map[*types.Named]bool
	helperWritten bool
}

func (*runtimedocGen) Name() string {
	return "runtimedoc"
}

func (*runtimedocGen) New(c gengo.Context) gengo.Generator {
	g := &runtimedocGen{
		SnippetWriter: c.Writer(),
		processed:     map[*types.Named]bool{},
	}

	return g
}

func (g *runtimedocGen) GenerateType(c gengo.Context, named *types.Named) error {
	return g.generateType(c, named)
}

func (g *runtimedocGen) createHelperOnce() {
	if g.helperWritten {
		return
	}
	g.helperWritten = true

	g.Do(`
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface { RuntimeDoc(names ...string) ([]string, bool) }); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

`)
}

func (g *runtimedocGen) generateType(c gengo.Context, named *types.Named) error {
	if _, ok := g.processed[named]; ok {
		return nil
	}
	g.processed[named] = true

	g.createHelperOnce()

	defers := make([]*types.Named, 0)

	switch x := named.Underlying().(type) {

	case *types.Struct:
		_, doc := c.Doc(named.Obj())

		g.Do(`
func(v @Type) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		@cases
		}
		@embeds
		return nil, false
	}
	return @doc, true
}

`, gengo.Args{
			"Type": gengo.ID(named.Obj()),
			"doc":  doc,
			"cases": func(sw gengo.SnippetWriter) {
				for i := 0; i < x.NumFields(); i++ {
					f := x.Field(i)
					_, fieldDoc := c.Doc(f)

					if _, ok := f.Type().(*types.Struct); ok {
						panic("not support inline struct")
					}

					// skip empty struct
					if s, ok := f.Type().Underlying().(*types.Struct); ok {
						if s.NumFields() == 0 {
							continue
						}
					}

					if sub, ok := f.Type().(*types.Named); ok {
						if sub.Obj().Pkg().Path() == named.Obj().Pkg().Path() {
							defers = append(defers, named)
						}
					}

					sw.Do(`
case @fieldName:
	return @fieldDoc, true
`, gengo.Args{
						"fieldName": f.Name(),
						"fieldDoc":  fieldDoc,
					})
				}
			},
			"embeds": func(sw gengo.SnippetWriter) {
				for i := 0; i < x.NumFields(); i++ {
					f := x.Field(i)
					if f.Embedded() {
						if s, ok := f.Type().Underlying().(*types.Struct); ok {
							if s.NumFields() == 0 {
								continue
							}
						}

						sw.Do(`
if doc, ok := runtimeDoc(v.@fieldName, names...); ok  {
	return doc, ok
}
`, gengo.Args{
							"fieldName": gengo.ID(f.Name()),
						})
					}
				}
			},
		})

	default:
		_, doc := c.Doc(named.Obj())

		g.Do(`
func(@Type) RuntimeDoc(names ...string) ([]string, bool) {
	return @doc, true
}
`, gengo.Args{
			"Type": gengo.ID(named.Obj()),
			"doc":  doc,
		})
	}

	for i := range defers {
		if err := g.generateType(c, defers[i]); err != nil {
			return err
		}
	}

	return nil
}
