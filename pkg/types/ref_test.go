package types

import (
	"reflect"
	"testing"

	"github.com/octohelm/x/cmp"
	"github.com/octohelm/x/testing/bdd"
	. "github.com/octohelm/x/testing/v2"
)

type List[T any] struct {
	Items []T `json:"items"`
}

func TestRef(t *testing.T) {
	t.Run("GIVEN List[string]类型", func(t *testing.T) {
		tpe := reflect.TypeFor[List[string]]()

		t.Run("WHEN 生成并解析类型引用", func(t *testing.T) {
			refStr := Ref(tpe.PkgPath(), tpe.Name()).String()
			ref := bdd.MustDo(func() (TypeName, error) {
				return ParseRef(refStr)
			})

			Then(t, "包路径应该正确",
				Expect(ref.Pkg().Path(),
					Equal("github.com/octohelm/gengo/pkg/types"),
				),
			)

			Then(t, "类型名称应该正确",
				Expect(ref.Name(),
					Equal("List[string]"),
				),
			)

			Then(t, "重新序列化应该一致",
				Expect(ref.String(),
					Equal(refStr),
				),
			)
		})
	})
}

func TestTypeRef(t *testing.T) {
	t.Run("GIVEN List[List[string]]类型", func(t *testing.T) {
		tpe := reflect.TypeFor[List[List[string]]]()

		t.Run("WHEN 生成类型引用", func(t *testing.T) {
			x := Ref(tpe.PkgPath(), tpe.Name()).String()

			Then(t, "解析类型引用应该成功",
				ExpectMustValue(
					func() (*TypeRef, error) {
						return ParseTypeRef(x)
					},
					Be(cmp.NotNil[*TypeRef]()),
				),
			)

			t.Run("THEN 验证解析结果", func(t *testing.T) {
				ref := MustValue(t, func() (*TypeRef, error) {
					return ParseTypeRef(x)
				})

				Then(t, "字符串表示应该正确",
					Expect(ref.String(),
						Equal("github.com/octohelm/gengo/pkg/types.List[github.com/octohelm/gengo/pkg/types.List[string]]"),
					),
				)

				Then(t, "包路径应该正确",
					Expect(ref.PkgPath,
						Equal("github.com/octohelm/gengo/pkg/types"),
					),
				)

				Then(t, "解析后的类型引用可以再次解析",
					ExpectMustValue(
						func() (*TypeRef, error) {
							return ParseTypeRef(ref.String())
						},
						Equal(ref),
					),
				)
			})
		})
	})
}
