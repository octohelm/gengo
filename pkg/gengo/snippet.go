package gengo

import (
	"bytes"
	"fmt"
	"go/types"
	"io"
	"reflect"
	"strings"
	"text/scanner"

	"github.com/octohelm/gengo/pkg/namer"
	gengotypes "github.com/octohelm/gengo/pkg/types"
	typesx "github.com/octohelm/x/types"
)

type SnippetWriter interface {
	io.Writer
	Render(snippet Snippet)
}

func NewSnippetWriter(w io.Writer, ns namer.NameSystems) SnippetWriter {
	return &snippetWriter{
		Writer: w,
		ns:     ns,
	}
}

type snippetWriter struct {
	io.Writer
	ns namer.NameSystems
}

func (sw *snippetWriter) Render(snippet Snippet) {
	if snippet == nil {
		return
	}

	switch snippet[T].(type) {
	case string:
		if s, ok := snippet[T].(string); ok {
			sw.render(s, snippet)
		}
	}
}

func (sw *snippetWriter) writeString(s string) {
	_, _ = io.WriteString(sw.Writer, s)
}

func (sw *snippetWriter) Dumper() *Dumper {
	if rawNamer, ok := sw.ns["raw"]; ok {
		return NewDumper(rawNamer)
	}
	return nil
}

func (sw *snippetWriter) render(format string, args ...map[string]any) {
	argSet := map[string]any{}

	for i := range args {
		a := args[i]
		for k := range a {
			argSet[k] = a[k]
		}
	}

	s := &scanner.Scanner{}
	s.Init(bytes.NewBuffer([]byte(strings.TrimLeft(format, "\n"))))
	s.Error = func(s *scanner.Scanner, msg string) {}

	c := s.Next()
	for {
		if c == scanner.EOF {
			break
		}

		switch c {
		case '@':
			named := bytes.NewBuffer(nil)

			for {
				c = s.Next()

				if c == scanner.EOF || c == '\'' {
					break
				}

				if (c >= 'A' && c <= 'Z') ||
					(c >= 'a' && c <= 'z') ||
					(c >= '0' && c <= '9') ||
					c == '_' {

					named.WriteRune(c)
					continue
				}
				break
			}

			if named.Len() > 0 {
				name := named.String()

				if v, ok := argSet[name]; ok {
					switch x := v.(type) {
					case Render:
						x(NewSnippetWriter(sw.Writer, sw.ns))
					case Snippet:
						NewSnippetWriter(sw.Writer, sw.ns).Render(x)
					case SnippetBuild:
						NewSnippetWriter(sw.Writer, sw.ns).Render(x())
					case []Snippet:
						for i := range x {
							NewSnippetWriter(sw.Writer, sw.ns).Render(x[i])
						}
					case Name:
						sw.writeString(x(sw.Dumper()))
					default:
						sw.writeString(sw.Dumper().ValueLit(x))
					}
				} else {
					panic(fmt.Sprintf("missing named arg `%s` in %s", name, format))
				}
			}

			if c == '@' {
				continue
			}

			if !(c == scanner.EOF || c == '\'') {
				sw.writeString(string(c))
			}
		default:
			sw.writeString(string(c))
		}

		c = s.Next()
	}
}

const T = "__T__"

func SnippetT(t string) Snippet {
	return Snippet{T: t}
}

type Snippet map[string]any

type SnippetBuild = func() Snippet

type Render = func(sw SnippetWriter)

type Name = func(d *Dumper) string

func ID(v any) Name {
	return func(d *Dumper) string {
		switch x := v.(type) {
		case string:
			ref, err := gengotypes.ParseRef(x)
			if err != nil {
				return x
			}
			return d.Name(ref)
		case gengotypes.TypeName:
			return d.Name(x)
		case reflect.Type:
			return d.TypeLit(typesx.FromRType(x))
		case types.Type:
			return d.TypeLit(typesx.FromTType(x))
		default:
			panic(fmt.Sprintf("unspported %T", v))
		}
		return ""
	}
}

func Comment(v string) Render {
	return func(sw SnippetWriter) {
		if v == "" {
			return
		}

		for i, l := range strings.Split(v, "\n") {
			if i > 0 {
				_, _ = io.WriteString(sw, "\n")
			}
			_, _ = io.WriteString(sw, "// ")
			_, _ = io.WriteString(sw, l)
		}
	}
}

func EachSnippet(n int, build func(i int) Snippet) []Snippet {
	snippets := make([]Snippet, n)

	for i := 0; i < n; i++ {
		snippets[i] = build(i)
	}

	return snippets
}

func MapSnippet[T any](list []T, build func(item T) Snippet) []Snippet {
	snippets := make([]Snippet, len(list))

	for i := range list {
		snippets[i] = build(list[i])
	}

	return snippets
}
