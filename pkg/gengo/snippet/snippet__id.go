package snippet

import (
	"context"
	"fmt"
	"iter"
	"reflect"
	"strings"

	"github.com/octohelm/gengo/pkg/gengo/internal"
	gengotypes "github.com/octohelm/gengo/pkg/types"
	typesx "github.com/octohelm/x/types"
	"go/types"
)

func PkgExpose(pkgPath string, expose string) Snippet {
	return &pkgExposer{
		typeName: gengotypes.Ref(pkgPath, expose),
	}
}

func PkgExposeOf(x any) Snippet {
	return pkgExpose(reflect.TypeOf(x))
}

func PkgExposeFor[T any](exposes ...string) Snippet {
	if len(exposes) > 0 {
		return &pkgExposer{
			typeName: gengotypes.Ref(reflect.TypeFor[T]().PkgPath(), exposes[0]),
		}
	}

	return pkgExpose(reflect.TypeFor[T]())
}

func pkgExpose(tp reflect.Type) Snippet {
	for tp.Kind() == reflect.Pointer {
		tp = tp.Elem()
	}

	if tp.Kind() == reflect.Func {
		panic(fmt.Errorf("unsupported %s, which cannot get pkgPath and name", tp))
	}

	return &pkgExposer{
		typeName: gengotypes.Ref(tp.PkgPath(), strings.SplitN(tp.Name(), "[", 2)[0]),
	}
}

type pkgExposer struct {
	typeName gengotypes.TypeName
}

func (i *pkgExposer) IsNil() bool {
	return i.typeName == nil
}

func (i *pkgExposer) Frag(ctx context.Context) iter.Seq[string] {
	d := internal.DumperContext.From(ctx)
	return func(yield func(string) bool) {
		if !yield(d.Name(i.typeName)) {
			return
		}

	}
}

func ID(v any) Snippet {
	return &ident{v: v}
}

type ident struct {
	v any
}

func (i *ident) IsNil() bool {
	return i.v == nil
}

func (i *ident) Frag(ctx context.Context) iter.Seq[string] {
	d := internal.DumperContext.From(ctx)

	return func(yield func(string) bool) {
		switch x := i.v.(type) {
		case *types.Alias:
			ref, err := gengotypes.ParseRef(x.String())
			if err != nil {
				if !yield(x.String()) {
					return
				}
				return
			}
			if !yield(d.Name(ref)) {
				return
			}
			return
		case string:
			ref, err := gengotypes.ParseRef(x)
			if err != nil {
				if !yield(x) {
					return
				}
				return
			}
			if !yield(d.Name(ref)) {
				return
			}
			return
		case gengotypes.TypeName:
			if !yield(d.Name(x)) {
				return
			}
			return
		case reflect.Type:
			if !yield(d.TypeLit(typesx.FromRType(x))) {
				return
			}
			return
		case types.Type:
			if !yield(d.TypeLit(typesx.FromTType(x))) {
				return
			}
			return
		default:
			panic(fmt.Sprintf("unspported %T", x))
		}
	}
}
