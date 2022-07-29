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

func RangeSnippet[T any](snippetTemplates map[string]string, list []T, buildArgs func(item T) Args) Args {
	args := Args{}

	for k := range snippetTemplates {
		snippetTemplate := snippetTemplates[k]

		args[k] = func(s SnippetWriter) {
			for i := range list {
				s.Do(snippetTemplate, buildArgs(list[i]))
			}
		}
	}

	return args
}

func Snippet[T any](snippetTemplates map[string]string, input T, buildArgs func(input T) Args) Args {
	args := Args{}

	for k := range snippetTemplates {
		snippetTemplate := snippetTemplates[k]

		args[k] = func(s SnippetWriter) {
			s.Do(snippetTemplate, buildArgs(input))
		}
	}

	return args
}

type SnippetWriter interface {
	io.Writer
	Dumper() *Dumper
	Do(snippetTemplate string, args ...Args)
}

type Args = map[string]any

func NewSnippetWriter(w io.Writer, ns namer.NameSystems) SnippetWriter {
	sw := &snippetWriter{
		Writer: w,
		ns:     ns,
	}
	return sw
}

type snippetWriter struct {
	io.Writer
	ns namer.NameSystems
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

func (sw *snippetWriter) Do(format string, args ...Args) {
	argSet := Args{}

	for i := range args {
		a := args[i]
		for k := range a {
			argSet[k] = a[k]
		}
	}

	s := &scanner.Scanner{}
	s.Init(bytes.NewBuffer([]byte(strings.TrimLeft(format, "\n"))))
	s.Error = func(s *scanner.Scanner, msg string) {}

	for c := s.Next(); c != scanner.EOF; c = s.Next() {
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
					switch arg := v.(type) {
					case Name:
						sw.writeString(arg(sw.Dumper()))
					case Render:
						arg(NewSnippetWriter(sw.Writer, sw.ns))
					default:
						sw.writeString(sw.Dumper().ValueLit(arg))
					}
				} else {
					panic(fmt.Sprintf("missing named arg `%s` in %s", name, format))
				}
			}

			if !(c == scanner.EOF || c == '\'') {
				sw.writeString(string(c))
			}
		default:
			sw.writeString(string(c))
		}
	}
}

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
