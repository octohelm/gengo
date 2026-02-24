package gengo

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"
)

func TestHandlerAPI(t *testing.T) {
	h := &handler{
		h: slog.NewTextHandler(io.Discard, nil),
	}

	Then(t, "WithAttrs 与 WithGroup 应返回可用 handler",
		Expect(h.WithAttrs([]slog.Attr{slog.String("k", "v")}), Be(cmp.NotNil[slog.Handler]())),
		Expect(h.WithGroup("g"), Be(cmp.NotNil[slog.Handler]())),
	)

	t.Run("Handle", func(t *testing.T) {
		rType := slog.NewRecord(time.Now(), slog.LevelInfo, "done", 0)
		rType.AddAttrs(
			slog.String("scope", "github.com/octohelm/gengo/testdata/a/c"),
			slog.String("type", "Obj"),
			slog.String("gengo", "deepcopy"),
			slog.Bool("cached", true),
		)
		Must(t, func() error {
			return h.Handle(context.Background(), rType)
		})

		Then(t, "带 type 的记录应先缓存到对应 scope",
			Expect(len(h.pkgs["github.com/octohelm/gengo/testdata/a/c"]), Equal(1)),
		)

		rPkg := slog.NewRecord(time.Now(), slog.LevelError, "failed", 0)
		rPkg.AddAttrs(
			slog.String("scope", "github.com/octohelm/gengo/testdata/a/c"),
			slog.Duration("cost", 10*time.Millisecond),
		)
		Must(t, func() error {
			return h.Handle(context.Background(), rPkg)
		})
	})
}

func TestLoggerAPI(t *testing.T) {
	l := newLogger()

	t.Run("toAttrs", func(t *testing.T) {
		ll := (&logger{}).WithValues("k", "v").(*logger)
		Then(t, "无 span 时应仅返回原始 attrs",
			Expect(len(ll.toAttrs()), Equal(2)),
		)

		ll.spans = []string{"g1", "g2"}
		Then(t, "有 span 时应追加 span 属性",
			Expect(len(ll.toAttrs()), Equal(3)),
			ExpectMust(func() error {
				spanAttr, ok := ll.toAttrs()[2].(slog.Attr)
				if !ok {
					return errors.New("expected slog.Attr")
				}
				if spanAttr.Key != "span" || spanAttr.Value.String() != "g1 g2" {
					return errors.New("unexpected span attr")
				}
				return nil
			}),
		)
	})

	t.Run("Error", func(t *testing.T) {
		Then(t, "Error 调用应可执行",
			ExpectMust(func() error {
				l.Error(errors.New("boom"))
				return nil
			}),
		)
	})
}
