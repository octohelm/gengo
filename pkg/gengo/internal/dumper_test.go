package internal

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/gengo/pkg/namer"
)

type Item struct {
	Name string `json:"name"`
}

type List[T any] struct {
	Items []T `json:"items,omitempty"`
}

func TestDumper_TypeLit(t *testing.T) {
	t.Run("GIVEN Dumper 实例", func(t *testing.T) {
		d := NewDumper(namer.NewRawNamer("", namer.NewDefaultImportTracker()))

		Then(t, "Dumper 应该创建成功",
			Expect(d, Be(cmp.NotNil[*Dumper]())),
		)

		t.Run("WHEN 处理类型字面量", func(t *testing.T) {
			t.Run("基础类型", func(t *testing.T) {
				Then(t, "指针类型",
					Expect(d.ReflectTypeLit(reflect.TypeFor[*bytes.Buffer]()),
						Equal("*bytes.Buffer"),
					),
				)

				Then(t, "切片类型",
					Expect(d.ReflectTypeLit(reflect.TypeFor[[]string]()),
						Equal("[]string"),
					),
				)

				Then(t, "Map 类型",
					Expect(d.ReflectTypeLit(reflect.TypeFor[map[string]string]()),
						Equal("map[string]string"),
					),
				)

				Then(t, "带标签的结构体指针",
					Expect(d.ReflectTypeLit(reflect.TypeFor[*struct {
						V string "json:\"v\" validate:\"@int[0,10]\""
					}]()),
						Equal(`*struct {V string `+"`json:\"v\" validate:\"@int[0,10]\"`"+`
}`),
					),
				)
			})

			t.Run("泛型类型", func(t *testing.T) {
				Then(t, "简单泛型实例",
					Expect(d.ReflectTypeLit(reflect.TypeFor[*List[Item]]()),
						Equal("*internal.List[internal.Item]"),
					),
				)

				Then(t, "嵌套泛型实例",
					Expect(d.ReflectTypeLit(reflect.TypeFor[*List[List[Item]]]()),
						Equal("*internal.List[internal.List[internal.Item]]"),
					),
				)
			})
		})

		t.Run("WHEN 处理值字面量", func(t *testing.T) {
			t.Run("指针值", func(t *testing.T) {
				value := &bytes.Buffer{}

				Then(t, "指针值表示",
					Expect(d.ValueLit(reflect.ValueOf(value)),
						Equal("&(bytes.Buffer{})"),
					),
				)
			})

			t.Run("切片值", func(t *testing.T) {
				value := []string{"1", "2"}

				Then(t, "切片值表示",
					Expect(d.ValueLit(reflect.ValueOf(value)),
						Equal(`[]string{
"1",
"2",
}`),
					),
				)
			})

			t.Run("空切片", func(t *testing.T) {
				value := []string{}

				Then(t, "空切片表示",
					Expect(d.ValueLit(reflect.ValueOf(value)),
						Equal("[]string{}"),
					),
				)
			})

			t.Run("结构体值", func(t *testing.T) {
				value := Item{Name: "test"}

				Then(t, "结构体值表示",
					Expect(d.ValueLit(reflect.ValueOf(value)),
						Equal(`internal.Item{
Name:"test",
}`),
					),
				)
			})
		})

		t.Run("边界情况", func(t *testing.T) {
			t.Run("nil值", func(t *testing.T) {
				var ptr *bytes.Buffer

				Then(t, "nil指针值",
					Expect(d.ValueLit(reflect.ValueOf(ptr)),
						Equal("nil"),
					),
				)
			})

			t.Run("零值", func(t *testing.T) {
				value := bytes.Buffer{}

				Then(t, "零值结构体",
					Expect(d.ValueLit(reflect.ValueOf(value)),
						Equal("bytes.Buffer{}"),
					),
				)
			})

			t.Run("Map值", func(t *testing.T) {
				value := map[string]int{"a": 1, "b": 2}

				Then(t, "Map 值表示",
					Expect(d.ValueLit(reflect.ValueOf(value)),
						Be(func(s string) error {
							// Map 输出顺序可能不确定，检查基本格式
							if len(s) == 0 {
								return errors.New("map 表示不应该为空")
							}
							if !strings.Contains(s, "map[string]int{") {
								return errors.New("应该以 map 类型开头")
							}
							return nil
						}),
					),
				)
			})
		})
	})
}
