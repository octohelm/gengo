package namer_test

import (
	"testing"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/gengo/pkg/namer"
	gengotypes "github.com/octohelm/gengo/pkg/types"
)

func TestDefaultImportTracker(t *testing.T) {
	t.Run("WHEN 添加标准库类型", func(t *testing.T) {
		tracker := namer.NewDefaultImportTracker()
		tracker.AddType(gengotypes.Ref("fmt", "Stringer"))

		Then(t, "THEN 应使用标准库本地名",
			Expect(tracker.LocalNameOf("fmt"), Equal("fmt")),
		)

		Then(t, "THEN 应可通过本地名反查路径",
			ExpectMustValue(func() (string, error) {
				path, ok := tracker.PathOf("fmt")
				if !ok {
					return "", cmp.NotNil[*string]()(nil)
				}
				return path, nil
			}, Equal("fmt")),
		)
	})

	t.Run("WHEN 添加与标准库重名的外部包", func(t *testing.T) {
		tracker := namer.NewDefaultImportTracker()
		tracker.AddType(gengotypes.Ref("github.com/example/context", "Value"))

		Then(t, "THEN 应避开标准库 context 名称",
			Expect(tracker.LocalNameOf("github.com/example/context"), Equal("examplecontext")),
		)
	})

	t.Run("WHEN 添加 domain 或 apis 风格路径", func(t *testing.T) {
		tracker := namer.NewDefaultImportTracker()
		tracker.AddType(gengotypes.Ref("github.com/example/domain/user", "User"))
		tracker.AddType(gengotypes.Ref("github.com/example/apis/order", "Order"))

		Then(t, "THEN domain 后缀命名遇到标准库冲突时应继续退避",
			Expect(tracker.LocalNameOf("github.com/example/domain/user"), Equal("domainuser")),
		)

		Then(t, "THEN apis 后缀应优先使用资源名",
			Expect(tracker.LocalNameOf("github.com/example/apis/order"), Equal("order")),
		)
	})

	t.Run("WHEN 查询 Imports", func(t *testing.T) {
		tracker := namer.NewDefaultImportTracker()
		tracker.AddType(gengotypes.Ref("fmt", "Stringer"))
		tracker.AddType(gengotypes.Ref("github.com/example/context", "Value"))

		Then(t, "THEN Imports 应返回路径到本地名映射",
			Expect(tracker.Imports(), Equal(map[string]string{
				"fmt":                        "fmt",
				"github.com/example/context": "examplecontext",
			})),
		)
	})
}

func TestRawNamer(t *testing.T) {
	t.Run("WHEN 渲染同包类型", func(t *testing.T) {
		tracker := namer.NewDefaultImportTracker()
		n := namer.NewRawNamer("github.com/octohelm/gengo/pkg/namer", tracker)

		Then(t, "THEN 应只输出类型名",
			Expect(n.Name(gengotypes.Ref("github.com/octohelm/gengo/pkg/namer", "Namer")), Equal("Namer")),
		)

		Then(t, "THEN 不应产生额外 import",
			Expect(len(tracker.Imports()), Equal(0)),
		)
	})

	t.Run("WHEN 渲染跨包类型", func(t *testing.T) {
		tracker := namer.NewDefaultImportTracker()
		n := namer.NewRawNamer("github.com/octohelm/gengo/pkg/namer", tracker)

		Then(t, "THEN 应输出带本地包名前缀的类型",
			Expect(n.Name(gengotypes.Ref("fmt", "Stringer")), Equal("fmt.Stringer")),
		)

		Then(t, "THEN import 应被记录",
			Expect(tracker.Imports(), Equal(map[string]string{
				"fmt": "fmt",
			})),
		)
	})

	t.Run("WHEN 渲染带类型参数的引用", func(t *testing.T) {
		tracker := namer.NewDefaultImportTracker()
		n := namer.NewRawNamer("github.com/octohelm/gengo/pkg/namer", tracker)

		typeName := gengotypes.Ref(
			"github.com/octohelm/gengo/pkg/namer",
			"Wrapper[github.com/octohelm/gengo/pkg/namer.Item,github.com/example/context.Value]",
		)

		Then(t, "THEN 同包参数应去掉包前缀，跨包参数应使用本地名",
			Expect(n.Name(typeName), Equal("Wrapper[Item,examplecontext.Value]")),
		)

		Then(t, "THEN 仅跨包参数会进入 import",
			Expect(tracker.Imports(), Equal(map[string]string{
				"github.com/example/context": "examplecontext",
			})),
		)
	})
}
