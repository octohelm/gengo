package snippet

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"maps"
	"strings"
	"text/scanner"
)

// IDArg 为 name 绑定一个按标识符方式渲染的模板参数。
func IDArg(name string, id any) TArg {
	return &arg{name: name, snippet: ID(id)}
}

// ValueArg 为 name 绑定一个按 Go 值字面量渲染的模板参数。
func ValueArg(name string, v any) TArg {
	return &arg{name: name, snippet: Value(v)}
}

// Arg 为 name 绑定一个 snippet 模板参数。
func Arg(name string, snippet Snippet) TArg {
	return &arg{name: name, snippet: snippet}
}

// Args 表示一组可复用的命名模板参数。
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

// TArg 为 T 提供一个或多个命名参数。
type TArg interface {
	Args() iter.Seq2[string, Snippet]
}

// T 渲染一个带命名参数的模板 snippet。
func T(fmt string, args ...TArg) Snippet {
	t := &template{
		format: fmt,
		args:   map[string]Snippet{},
	}

	for _, a := range args {
		if a != nil {
			maps.Insert(t.args, a.Args())
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
