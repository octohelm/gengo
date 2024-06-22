package types

import (
	"fmt"
	"github.com/pkg/errors"
	"go/ast"
	"go/types"
	"strings"
)

type TypeName interface {
	Pkg() *types.Package
	Name() string
	String() string
	Exported() bool
}

var _ TypeName = &types.TypeName{}

func Ref(pkgPath string, name string) TypeName {
	return &ref{pkgPath: pkgPath, name: name}
}

func ParseRef(ref string) (TypeName, error) {
	base := ref
	if i := strings.Index(ref, "["); i > 0 {
		base = base[0:i]
	}
	if i := strings.LastIndex(base, "."); i > 0 {
		return Ref(ref[0:i], ref[i+1:]), nil
	}
	return nil, fmt.Errorf("unsupported ref: %s", ref)
}

func MustParseRef(ref string) TypeName {
	r, err := ParseRef(ref)
	if err != nil {
		panic(nil)
	}
	return r
}

type ref struct {
	pkgPath string
	name    string
}

func (ref) Underlying() types.Type {
	return nil
}

func (r *ref) String() string {
	return r.pkgPath + "." + r.name
}

func (r *ref) Pkg() *types.Package {
	return types.NewPackage(r.pkgPath, "")
}

func (r *ref) Name() string {
	return r.name
}

func (r *ref) Exported() bool {
	return ast.IsExported(r.name)
}

func ParseTypeRef(s string) (*TypeRef, error) {
	if i := strings.Index(s, "["); i > 0 {
		if strings.LastIndex(s, "]") == len(s)-1 {
			t, err := ParseTypeRef(s[0:i])
			if err != nil {
				return nil, err
			}

			typeListStr := s[i+1 : len(s)-1]
			inTypeParam := false
			started := 0

			commit := func(i int) error {
				sub, err := ParseTypeRef(typeListStr[started:i])
				if err != nil {
					return err
				}
				t.TypeList = append(t.TypeList, sub)
				started = i + 1
				return nil
			}

			for i, c := range typeListStr {
				switch c {
				case '[':
					inTypeParam = true
				case ']':
					inTypeParam = false
				case ',':
					if !inTypeParam {
						if err := commit(i); err != nil {
							return nil, err
						}
					}
				}
			}

			if err := commit(len(typeListStr)); err != nil {
				return nil, err
			}

			return t, nil
		}

		return nil, errors.Errorf("invalid type ref: %s", s)
	}

	if i := strings.LastIndex(s, "."); i > 0 {
		return &TypeRef{
			PkgPath: s[0:i],
			Name:    s[i+1:],
		}, nil
	}

	return &TypeRef{
		Name: s,
	}, nil
}

type TypeRef struct {
	Name     string
	PkgPath  string
	TypeList []*TypeRef
}

func (r *TypeRef) Walk(walk func(t *TypeRef) bool) {
	if !walk(r) {
		return
	}

	for _, t := range r.TypeList {
		t.Walk(walk)
	}
}

func (r *TypeRef) String() string {
	b := &strings.Builder{}

	if len(r.PkgPath) > 0 {
		b.WriteString(r.PkgPath)
		b.WriteByte('.')
	}

	b.WriteString(r.Name)

	if len(r.TypeList) > 0 {
		b.WriteByte('[')
		for i, x := range r.TypeList {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(x.String())
		}
		b.WriteByte(']')
	}

	return b.String()
}
