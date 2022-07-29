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
	processed map[*types.Named]bool
}

func (*runtimedocGen) Name() string {
	return "runtimedoc"
}

func (*runtimedocGen) New(c gengo.Context) gengo.Generator {
	return &runtimedocGen{
		SnippetWriter: c.Writer(),
		processed:     map[*types.Named]bool{},
	}
}

func (g *runtimedocGen) GenerateType(c gengo.Context, named *types.Named) error {
	return g.generateType(c, named)
}

func (g *runtimedocGen) generateType(c gengo.Context, named *types.Named) error {
	if _, ok := g.processed[named]; ok {
		return nil
	}
	g.processed[named] = true

	defers := make([]*types.Named, 0)

	switch x := named.Underlying().(type) {

	case *types.Struct:
		_, doc := c.Doc(named.Obj())

		g.Do(`
func(v [[ .type | id ]]) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		[[ .cases | render ]]
		}
		[[ .embeds | render ]]	
		return nil, false
	}
	return [[ .doc ]], true
}

`, gengo.Args{
			"type": named.Obj(),
			"doc":  g.Dumper().ValueLit(doc),
			"cases": func(sw gengo.SnippetWriter) {
				for i := 0; i < x.NumFields(); i++ {
					f := x.Field(i)
					_, fieldDoc := c.Doc(f)

					if _, ok := f.Type().(*types.Struct); ok {
						panic("not support inline struct")
					}

					if sub, ok := f.Type().(*types.Named); ok {
						if sub.Obj().Pkg().Path() == named.Obj().Pkg().Path() {
							defers = append(defers, named)
						}
					}

					sw.Do(`
case [[ .fieldName | quote ]]:
	return [[ .fieldDoc ]], true
`, gengo.Args{
						"fieldName": f.Name(),
						"fieldDoc":  g.Dumper().ValueLit(fieldDoc),
					})
				}
			},
			"embeds": func(sw gengo.SnippetWriter) {
				for i := 0; i < x.NumFields(); i++ {
					f := x.Field(i)
					if f.Embedded() {
						sw.Do(`
if doc, ok := v.[[ .fieldName ]].RuntimeDoc(names...); ok  {
	return doc, ok
}
`, gengo.Args{
							"fieldName": f.Name(),
						})
					}
				}
			},
		})

	default:
		_, doc := c.Doc(named.Obj())

		g.Do(`
func([[ .type | id ]]) RuntimeDoc(names ...string) ([]string, bool) {
	return [[ .doc ]], true
}
`, gengo.Args{
			"type": named.Obj(),
			"doc":  g.Dumper().ValueLit(doc),
		})
	}

	for i := range defers {
		if err := g.generateType(c, defers[i]); err != nil {
			return err
		}
	}

	return nil
}
