package snippet

import (
	"context"
	"iter"

	"github.com/octohelm/gengo/pkg/gengo/internal"
)

func Value(v any) Snippet {
	return &value{v: v}
}

type value struct {
	v any
}

func (v *value) IsNil() bool {
	return v.v == nil
}

func (v *value) Frag(ctx context.Context) iter.Seq[string] {
	d := internal.DumperContext.From(ctx)

	return func(yield func(string) bool) {
		if !yield(d.ValueLit(v.v)) {
			return
		}
	}
}
