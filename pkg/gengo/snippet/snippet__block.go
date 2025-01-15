package snippet

import (
	"context"
	"iter"
)

type Block string

func (v Block) IsNil() bool {
	return len(v) == 0
}

func (v Block) Frag(ctx context.Context) iter.Seq[string] {
	return func(yield func(string) bool) {
		if !yield(string(v)) {
			return
		}
	}
}
