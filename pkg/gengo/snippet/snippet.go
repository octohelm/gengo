package snippet

import (
	"context"
	"iter"
)

// Snippet 表示一个可延迟渲染的 Go 源码片段。
type Snippet interface {
	IsNil() bool
	Frag(ctx context.Context) iter.Seq[string]
}

// Fragments 会把 s 展开成字符串序列，并跳过 nil 的 snippet。
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

// Snippets 把多个 snippet 组合成一个片段序列。
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

// Func 把一个返回片段序列的函数适配成 Snippet。
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
