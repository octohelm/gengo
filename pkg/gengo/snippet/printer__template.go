package snippet

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"strings"
	"text/scanner"
)

func IDArg(name string, id any) TArg {
	return &arg{name: name, snippet: ID(id)}
}

func ValueArg(name string, v any) TArg {
	return &arg{name: name, snippet: Value(v)}
}

func Arg(name string, snippet Snippet) TArg {
	return &arg{name: name, snippet: snippet}
}

type Args map[string]Snippet

func (args Args) Args() iter.Seq2[string, Snippet] {
	return func(yield func(string, Snippet) bool) {
		for k, s := range args {
			if !yield(k, s) {
				return
			}
		}
	}
}

type arg struct {
	name    string
	snippet Snippet
}

func (a *arg) Args() iter.Seq2[string, Snippet] {
	return func(yield func(string, Snippet) bool) {
		if !yield(a.name, a.snippet) {
			return
		}
	}
}

type TArg interface {
	Args() iter.Seq2[string, Snippet]
}

func T(fmt string, args ...TArg) Snippet {
	t := &template{
		format: fmt,
		args:   map[string]Snippet{},
	}

	for _, a := range args {
		if a != nil {
			for name, s := range a.Args() {
				t.args[name] = s
			}
		}
	}

	return t
}

type template struct {
	format string
	args   map[string]Snippet
}

func (t *template) IsNil() bool {
	return len(t.format) == 0
}

func (t *template) Frag(ctx context.Context) iter.Seq[string] {
	return func(yield func(string) bool) {
		argSet := t.args
		if argSet == nil {
			argSet = map[string]Snippet{}
		}

		s := &scanner.Scanner{}
		s.Init(bytes.NewBuffer([]byte(strings.TrimLeft(t.format, "\n"))))
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
						if v.IsNil() {
							continue
						}

						for code := range v.Frag(ctx) {
							if !yield(code) {
								return
							}
						}
					} else {
						panic(fmt.Sprintf("missing named arg `%s` in %s", name, t.format))
					}
				}

				if c == '@' {
					continue
				}

				if !(c == scanner.EOF || c == '\'') {
					if !yield(string(c)) {
						return
					}
				}
			default:
				if !yield(string(c)) {
					return
				}
			}

			c = s.Next()
		}
	}
}
