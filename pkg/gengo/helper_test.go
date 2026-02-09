package gengo_test

import (
	"reflect"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/gengo/pkg/gengo"
	gengotypes "github.com/octohelm/gengo/pkg/types"
)

type List[T any] struct {
	Items []T `json:"items"`
}

func TestPkgImportPathAndExpose(t *testing.T) {
	t.Run("GIVEN List[string] 类型", func(t *testing.T) {
		tpe := reflect.TypeOf(List[string]{})

		t.Run("WHEN 生成类型引用", func(t *testing.T) {
			refStr := gengotypes.Ref(tpe.PkgPath(), tpe.Name()).String()

			t.Run("THEN 解析包导入路径和暴露名称", func(t *testing.T) {
				pkg, expose := gengo.PkgImportPathAndExpose(refStr)

				Then(t, "包导入路径应该正确",
					Expect(pkg, Equal("github.com/octohelm/gengo/pkg/gengo_test")),
				)

				Then(t, "暴露名称应该正确",
					Expect(expose, Equal("List")),
				)
			})
		})
	})

	t.Run("更多类型示例", func(t *testing.T) {
		testCases := []struct {
			name           string
			typeVal        any
			expectedPkg    string
			expectedExpose string
		}{
			{
				name:           "泛型类型实例",
				typeVal:        List[int]{},
				expectedPkg:    "github.com/octohelm/gengo/pkg/gengo_test",
				expectedExpose: "List",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				tpe := reflect.TypeOf(tc.typeVal)
				refStr := gengotypes.Ref(tpe.PkgPath(), tpe.Name()).String()

				pkg, expose := gengo.PkgImportPathAndExpose(refStr)

				Then(t, "包导入路径匹配",
					Expect(pkg, Equal(tc.expectedPkg)),
				)

				Then(t, "暴露名称匹配",
					Expect(expose, Equal(tc.expectedExpose)),
				)
			})
		}
	})

	// 测试边界情况
	t.Run("边界情况", func(t *testing.T) {
		t.Run("WHEN 处理内置类型", func(t *testing.T) {
			// 假设处理内置类型
			t.Run("内置string类型", func(t *testing.T) {
				pkg, expose := gengo.PkgImportPathAndExpose("string")

				Then(t, "内置类型应该返回空包路径",
					Expect(pkg, Equal("")),
				)

				Then(t, "内置类型应该返回类型名",
					Expect(expose, Equal("string")),
				)
			})
		})
	})
}
