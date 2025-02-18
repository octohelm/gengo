package snippet

import (
	"context"
	"iter"
)

func GoDirective(directive string, args ...string) Snippet {
	return Func(func(ctx context.Context) iter.Seq[string] {
		return func(yield func(code string) bool) {
			if directive == "" {
				return
			}

			if !yield("//go:") {
				return
			}

			if !yield(directive) {
				return
			}

			for _, arg := range args {
				if len(arg) > 0 {
					if !yield(" ") {
						return
					}
					if !yield(arg) {
						return
					}
				}
			}
		}
	})
}
