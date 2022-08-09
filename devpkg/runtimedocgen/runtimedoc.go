package runtimedocgen

import (
	"github.com/octohelm/gengo/pkg/gengo"
	"go/ast"
	"go/types"
)

func init() {
	gengo.Register(&runtimedocGen{})
}

type runtimedocGen struct {
	processed     map[*types.Named]bool
	helperWritten bool
}

func (*runtimedocGen) Name() string {
	return "runtimedoc"
}

func (g *runtimedocGen) GenerateType(c gengo.Context, named *types.Named) error {
	if !ast.IsExported(named.Obj().Name()) {
		return gengo.ErrSkip
	}

	if g.processed == nil {
		g.processed = map[*types.Named]bool{}
	}

	return g.generateType(c, named)
}

func (g *runtimedocGen) createHelperOnce(c gengo.Context) {
	if g.helperWritten {
		return
	}
	g.helperWritten = true

	c.Render(gengo.Snippet{gengo.T: `
// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

`})
}

func hasExposeField(t *types.Struct) bool {
	for i := 0; i < t.NumFields(); i++ {
		if t.Field(i).Exported() {
			return true
		}
	}
	return false
}

func (g *runtimedocGen) generateType(c gengo.Context, named *types.Named) error {
	if _, ok := g.processed[named]; ok {
		return nil
	}
	g.processed[named] = true

	g.createHelperOnce(c)

	defers := make([]*types.Named, 0)

	switch x := named.Underlying().(type) {
	case *types.Struct:
		if !hasExposeField(x) {
			return nil
		}

		_, doc := c.Doc(named.Obj())

		c.Render(gengo.Snippet{gengo.T: `
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

`,
			"Type": gengo.ID(named.Obj()),
			"doc":  doc,
			"cases": func(sw gengo.SnippetWriter) {
				for i := 0; i < x.NumFields(); i++ {
					f := x.Field(i)

					if !ast.IsExported(f.Name()) {
						continue
					}

					_, fieldDoc := c.Doc(f)

					if _, ok := f.Type().(*types.Struct); ok {
						panic("not support inline struct")
					}

					// skip empty struct
					if s, ok := f.Type().Underlying().(*types.Struct); ok {
						if !hasExposeField(s) {
							continue
						}
					}

					if sub, ok := f.Type().(*types.Named); ok {
						if isCustomDefinedNamed(sub) && sub.Obj().Pkg().Path() == named.Obj().Pkg().Path() {
							defers = append(defers, named)
						}
					}

					sw.Render(gengo.Snippet{gengo.T: `
case @fieldName:
	return @fieldDoc, true
`,
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
							if !hasExposeField(s) {
								continue
							}
						}

						sw.Render(gengo.Snippet{gengo.T: `
if doc, ok := runtimeDoc(v.@fieldName, names...); ok  {
	return doc, ok
}
`,
							"fieldName": gengo.ID(f.Name()),
						})
					}
				}
			},
		})

	default:
		_, doc := c.Doc(named.Obj())

		c.Render(gengo.Snippet{gengo.T: `
func(@Type) RuntimeDoc(names ...string) ([]string, bool) {
	return @doc, true
}
`,
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

func isCustomDefinedNamed(sub *types.Named) bool {
	return sub.Obj() != nil && sub.Obj().Pkg() != nil
}
