package cache_test

import (
	"testing"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/gengo/pkg/gengo/internal/cache"
)

func TestCache(t *testing.T) {
	t.Run("GIVEN 一个缓存目录", func(t *testing.T) {
		dir := t.TempDir()

		t.Run("WHEN 查询不存在的 id", func(t *testing.T) {
			c := cache.NewWithDir(dir)

			Then(
				t, "THEN Exists 应返回 false",
				Expect(c.Exists("abc123"), Be(cmp.False())),
			)
		})

		t.Run("WHEN Mark 一个 id 后查询", func(t *testing.T) {
			c := cache.NewWithDir(dir)

			Must(t, func() error {
				return c.Mark("abc123")
			})

			Then(
				t, "THEN Exists 应返回 true",
				Expect(c.Exists("abc123"), Be(cmp.True())),
			)
		})
	})

	t.Run("GIVEN 不同缓存目录", func(t *testing.T) {
		dir1 := t.TempDir()
		dir2 := t.TempDir()

		t.Run("WHEN 在 dir1 中 Mark 一个 id", func(t *testing.T) {
			c1 := cache.NewWithDir(dir1)
			c2 := cache.NewWithDir(dir2)

			Must(t, func() error {
				return c1.Mark("abc123")
			})

			Then(
				t, "THEN dir2 中不应存在",
				Expect(c2.Exists("abc123"), Be(cmp.False())),
			)
		})
	})
}
