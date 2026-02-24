package types

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
)

// TypeName 是生成器运行时使用的最小类型引用接口。
type TypeName interface {
	Pkg() *types.Package
	Name() string
	String() string
	Exported() bool
}

var _ TypeName = &types.TypeName{}

// Ref 根据 pkgPath 和 name 构造一个轻量的 TypeName。
func Ref(pkgPath string, name string) TypeName {
	return &ref{pkgPath: pkgPath, name: name}
}

// ParseRef 解析一个带包限定的类型引用字符串。
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

// MustParseRef 解析 ref；如果无效则 panic。
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

// ParseTypeRef 解析一个可带类型参数的类型引用。
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

		return nil, fmt.Errorf("invalid type ref: %s", s)
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

// TypeRef 表示一个解析后的类型引用，可能包含类型参数。
type TypeRef struct {
	Name     string
	PkgPath  string
	TypeList []*TypeRef
}

// Walk 会按深度优先顺序遍历 r 及其嵌套类型参数。
func (r *TypeRef) Walk(walk func(t *TypeRef) bool) {
	if !walk(r) {
		return
	}

	for _, t := range r.TypeList {
		t.Walk(walk)
	}
}

// String 会把类型引用重新序列化为文本形式。
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
