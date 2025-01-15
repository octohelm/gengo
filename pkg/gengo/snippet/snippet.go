package snippet

import (
	"context"
	"iter"
)

type Snippet interface {
	IsNil() bool
	Frag(ctx context.Context) iter.Seq[string]
}

func Fragments(ctx context.Context, s Snippet) iter.Seq[string] {
	return func(yield func(string) bool) {
		if s.IsNil() {
			return
		}

		for c := range s.Frag(ctx) {
			if !yield(c) {
				return
			}
		}
	}
}

type Snippets iter.Seq[Snippet]

func (Snippets) IsNil() bool {
	return false
}

func (Snippets) String() string {
	return ""
}

func (f Snippets) Frag(ctx context.Context) iter.Seq[string] {
	return func(yield func(string) bool) {
		for c := range f {
			if c.IsNil() {
				continue
			}

			for v := range c.Frag(ctx) {
				if !yield(v) {
					return
				}
			}
		}
	}
}

func Func(f func(ctx context.Context) iter.Seq[string]) Snippet {
	return fn(f)
}

type fn func(ctx context.Context) iter.Seq[string]

func (fn) IsNil() bool {
	return false
}

func (f fn) Frag(ctx context.Context) iter.Seq[string] {
	return func(yield func(string) bool) {
		for c := range f(ctx) {
			if !yield(c) {
				return
			}
		}
	}
}
