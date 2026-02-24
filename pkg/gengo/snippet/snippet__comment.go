package snippet

import (
	"context"
	"iter"
	"strings"
)

// Comment 将 v 渲染为一行或多行注释。
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
