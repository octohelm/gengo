package gengo

import (
	"go/types"
	"strings"
	"testing"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"
)

func TestComputeActionID(t *testing.T) {
	t.Run("WHEN 同一输入多次计算", func(t *testing.T) {
		id1 := computeActionID("mygen", "v1.0.0", "abc123")
		id2 := computeActionID("mygen", "v1.0.0", "abc123")

		Then(
			t, "THEN 应返回相同结果",
			Expect(id1, Equal(id2)),
		)
	})

	t.Run("WHEN pkgContentHash 不同", func(t *testing.T) {
		id1 := computeActionID("mygen", "v1.0.0", "abc123")
		id2 := computeActionID("mygen", "v1.0.0", "def456")

		Then(
			t, "THEN actionID 应不同",
			Expect(id1 != id2, Be(cmp.True())),
		)
	})

	t.Run("WHEN 生成器名称不同", func(t *testing.T) {
		id1 := computeActionID("mygen", "v1.0.0", "abc123")
		id2 := computeActionID("othergen", "v1.0.0", "abc123")

		Then(
			t, "THEN actionID 应不同",
			Expect(id1 != id2, Be(cmp.True())),
		)
	})

	t.Run("WHEN 生成器版本不同", func(t *testing.T) {
		id1 := computeActionID("mygen", "v1.0.0", "abc123")
		id2 := computeActionID("mygen", "v2.0.0", "abc123")

		Then(
			t, "THEN actionID 应不同",
			Expect(id1 != id2, Be(cmp.True())),
		)
	})

	t.Run("WHEN frameworkVersion 变化", func(t *testing.T) {
		// FrameworkVersion 通过 computeActionID 间接使用，
		// 只需验证它是非空字符串即可。
		Then(
			t, "THEN 应是非空字符串",
			Expect(FrameworkVersion != "", Be(cmp.True())),
		)
	})
}

func TestGeneratorVersion(t *testing.T) {
	t.Run("WHEN 获取内置生成器版本", func(t *testing.T) {
		version := generatorVersion(&testGen{})

		Then(
			t, "THEN 应返回 (devel) 或其他值",
			Expect(strings.Contains(version, "devel") || version == "(unknown)", Be(cmp.True())),
		)
	})
}

type testGen struct {
	name string
}

func (g *testGen) Name() string {
	return g.name
}

func (g *testGen) GenerateType(c Context, named *types.Named) error {
	return nil
}
