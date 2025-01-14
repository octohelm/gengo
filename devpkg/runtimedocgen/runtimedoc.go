package runtimedocgen

import (
	"fmt"
	"go/ast"
	"go/types"
	"regexp"

	"github.com/octohelm/gengo/pkg/gengo"
)

func init() {
	gengo.Register(&runtimedocGen{})
}

type runtimedocGen struct {
	processed map[*types.Named]bool

	helperWritten bool
}

func (*runtimedocGen) Name() string {
	return "runtimedoc"
}

func (g *runtimedocGen) GenerateType(c gengo.Context, named *types.Named) error {
	if _, ok := named.Obj().Type().Underlying().(*types.Interface); ok {
		return gengo.ErrSkip
	}

	if _, ok := named.Obj().Type().Underlying().(*types.Alias); ok {
		return gengo.ErrSkip
	}

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
import _ "embed"

// nolint:deadcode,unused
func runtimeDoc(v any, prefix string, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		doc, ok := c.RuntimeDoc(names...)
		if ok {	
			if prefix != "" && len(doc) > 0 {
				doc[0] = prefix + doc[0]
				return doc, true
			}

			return doc, true			
		}
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

var reEmbedDoc = regexp.MustCompile(`\[\[(?P<path>[^]]+)]]`)

func parseEmbed(prefix string, doc []string) (final []any, embeds []gengo.Snippet) {
	for i, line := range doc {
		if reEmbedDoc.MatchString(line) {
			matches := reEmbedDoc.FindStringSubmatch(line)
			path := matches[reEmbedDoc.SubexpIndex("path")]

			varName := gengo.LowerCamelCase(fmt.Sprintf("embed_doc_of_%s_%d", prefix, i))

			embeds = append(embeds, gengo.SnippetT(fmt.Sprintf(`
//go:embed %s
var %s string
`, path, varName)))

			final = append(final, gengo.ID(varName))
			continue
		}

		final = append(final, line)
	}

	return
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

		finalDoc, embeds := parseEmbed(named.Obj().Name(), doc)

		c.Render(gengo.Snippet{gengo.T: `
@externalEmbeds
func(v *@Type) RuntimeDoc(names ...string) ([]string, bool) {
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
			"externalEmbeds": func(sw gengo.SnippetWriter) {
				for _, e := range embeds {
					sw.Render(e)
				}
			},
			"doc": func(sw gengo.SnippetWriter) {
				sw.Render(gengo.SnippetT("[]string{"))
				for _, line := range finalDoc {
					sw.Render(gengo.Snippet{
						gengo.T: `@line,`,
						"line":  line,
					})
				}
				sw.Render(gengo.SnippetT(`
}`))
			},
			"cases": func(sw gengo.SnippetWriter) {
				for i := 0; i < x.NumFields(); i++ {
					f := x.Field(i)

					if !ast.IsExported(f.Name()) {
						continue
					}

					if f.Embedded() {
						continue
					}

					_, fieldDoc := c.Doc(f)

					if _, ok := f.Type().(*types.Struct); ok {
						c.Logger().Warn(fmt.Errorf("skip inline struct in %s", named))
						continue
					}

					// skip empty struct
					if s, ok := f.Type().Underlying().(*types.Struct); ok {
						if s.NumFields() == 0 {
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

						_, fieldDoc := c.Doc(f)

						prefix := ""

						if len(fieldDoc) > 0 {
							prefix = fieldDoc[0]
						}

						if _, ok := f.Type().(*types.Pointer); ok {
							sw.Render(gengo.Snippet{gengo.T: `
if doc, ok := runtimeDoc(v.@fieldName, @prefix, names...); ok  {
	return doc, ok
}
`,
								"fieldName": gengo.ID(f.Name()),
								"prefix":    prefix,
							})

							continue
						}

						sw.Render(gengo.Snippet{gengo.T: `
if doc, ok := runtimeDoc(&v.@fieldName, @prefix, names...); ok  {
	return doc, ok
}
`,
							"fieldName": gengo.ID(f.Name()),
							"prefix":    prefix,
						})
					}
				}
			},
		})

	default:
		_, doc := c.Doc(named.Obj())

		c.Render(gengo.Snippet{gengo.T: `
func(*@Type) RuntimeDoc(names ...string) ([]string, bool) {
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
