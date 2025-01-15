package snippet

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"text/scanner"
)

// Sprintf
// same as fmt.Sprintf, but only support:
//
//	%v print as go value
//	%T print as ident
func Sprintf(fmt string, args ...any) Snippet {
	return &printer{
		fmt:  fmt,
		args: args,
	}
}

type printer struct {
	fmt  string
	args []any
}

func (p *printer) IsNil() bool {
	return p.fmt == ""
}

func (p *printer) Frag(ctx context.Context) iter.Seq[string] {
	return func(yield func(string) bool) {
		s := &scanner.Scanner{}
		s.Init(bytes.NewBuffer([]byte(p.fmt)))
		s.Error = func(s *scanner.Scanner, msg string) {}

		argIdx := 0

		getArg := func() any {
			if argIdx < len(p.args) {
				a := p.args[argIdx]
				argIdx++
				return a
			}
			panic(fmt.Errorf("missing arg %d", argIdx))
		}

		c := s.Next()
		for {
			if c == scanner.EOF {
				break
			}

			switch c {
			case '%':
				c = s.Next()

				switch c {
				case 'T':
					a := getArg()

					switch x := a.(type) {
					case Snippet:
						for c := range x.Frag(ctx) {
							if !yield(c) {
								return
							}
						}
					default:
						for c := range ID(x).Frag(ctx) {
							if !yield(c) {
								return
							}
						}
					}
				case 'v':
					a := getArg()

					switch x := a.(type) {
					case Snippet:
						for c := range x.Frag(ctx) {
							if !yield(c) {
								return
							}
						}
					default:
						for c := range Value(x).Frag(ctx) {
							if !yield(c) {
								return
							}
						}
					}
				case '%':
					if !yield(string(c)) {
						return
					}
					continue
				default:
					panic(fmt.Errorf("unsupported %%%v", c))
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
