package snippet

import (
	"context"
	"iter"
	"strings"
)

func Comment(v string) Snippet {
	return Func(func(ctx context.Context) iter.Seq[string] {
		return func(yield func(code string) bool) {
			if v == "" {
				return
			}

			for i, l := range strings.Split(v, "\n") {
				if i > 0 {
					if !yield("\n") {
						return
					}
				}

				if !yield("// " + l) {
					return
				}
			}
		}
	})
}
