package gengo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/go-courier/logr"
)

func newLogger() *logger {
	return &logger{
		ctx:  context.Background(),
		slog: slog.New(&handler{h: slog.NewTextHandler(os.Stdout, nil)}),
	}
}

type handler struct {
	h   slog.Handler
	lvl slog.Level

	pkgs map[string][]*bytes.Buffer
}

func (h *handler) Handle(ctx context.Context, r slog.Record) error {
	if h.pkgs == nil {
		h.pkgs = make(map[string][]*bytes.Buffer)
	}

	line := bytes.NewBuffer(nil)

	scope := ""
	gengo := ""
	cost := ""
	tpe := ""
	cached := false

	for attr := range r.Attrs {
		if attr.Key == "cached" {
			cached = attr.Value.Bool()
		}
		if attr.Key == "type" {
			tpe = attr.Value.String()
		}
		if attr.Key == "scope" {
			scope = attr.Value.String()
		}
		if attr.Key == "gengo" {
			gengo = attr.Value.String()
		}
		if attr.Key == "cost" {
			cost = attr.Value.Duration().String()
		}
	}

	if gengo != "" {
		line.WriteString("\t --- ")
	} else {
		line.WriteString("--- ")
	}

	if r.Level == slog.LevelInfo {
		line.WriteString("DONE: ")
	} else if r.Level == slog.LevelError {
		line.WriteString("FAILED: ")
	}

	if gengo != "" {
		line.WriteString("+gengo:")
		line.WriteString(gengo)
		line.WriteString("\t")
	}

	if tpe != "" {
		line.WriteString(tpe)
	} else {
		line.WriteString(scope)
	}

	if cached {
		line.WriteString(" (cached)")
	} else if cost != "" {
		line.WriteString(" (")
		line.WriteString(cost)
		line.WriteString(")")
	}

	line.WriteByte('\n')

	if r.Level == slog.LevelError {
		line.WriteString(r.Message)
		line.WriteByte('\n')
	}

	if tpe == "" {
		_, _ = io.Copy(os.Stdout, line)

		if bufs, ok := h.pkgs[scope]; ok {
			for _, b := range bufs {
				_, _ = io.Copy(os.Stdout, b)
			}
		}
	} else {
		h.pkgs[scope] = append(h.pkgs[scope], line)
	}

	return nil
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &handler{h: h.h.WithAttrs(attrs)}
}

func (h *handler) WithGroup(name string) slog.Handler {
	return &handler{h: h.h.WithGroup(name)}
}

func (h *handler) Enabled(ctx context.Context, l slog.Level) bool {
	return l >= h.lvl
}

type logger struct {
	slog      *slog.Logger
	ctx       context.Context
	spans     []string
	attrs     []any
	startedAt time.Time
	debug     bool
}

func (d logger) WithValues(keyAndValues ...any) logr.Logger {
	d.attrs = append(d.attrs, keyAndValues...)
	return &d
}

func (d *logger) Start(ctx context.Context, name string, keyAndValues ...any) (context.Context, logr.Logger) {
	ll := &logger{
		slog: d.slog,
		ctx:  ctx,

		spans:     append(d.spans, name),
		attrs:     append(d.attrs, keyAndValues...),
		startedAt: time.Now(),
		debug:     strings.HasPrefix(name, "debug:"),
	}

	return logr.LoggerInjectContext(ctx, ll), ll
}

func (d *logger) End() {
	var dd logr.Logger = d
	if !d.startedAt.IsZero() {
		dd = dd.WithValues(slog.Duration("cost", time.Since(d.startedAt)))
	}

	if d.debug {
		dd.Debug("DONE")
	} else {
		dd.Info("DONE")
	}
}

func (d *logger) toAttrs() []any {
	if len(d.spans) == 0 {
		return d.attrs
	}
	return append(d.attrs, slog.String("span", strings.Join(d.spans, " ")))
}

func (d *logger) Debug(format string, args ...any) {
	if !d.slog.Enabled(d.ctx, slog.LevelDebug) {
		return
	}
	d.slog.Log(d.ctx, slog.LevelDebug, fmt.Sprintf(format, args...), d.toAttrs()...)
}

func (d *logger) Info(format string, args ...any) {
	if !d.slog.Enabled(d.ctx, slog.LevelInfo) {
		return
	}
	d.slog.Log(d.ctx, slog.LevelInfo, fmt.Sprintf(format, args...), d.toAttrs()...)
}

func (d *logger) Warn(err error) {
	if !d.slog.Enabled(d.ctx, slog.LevelWarn) {
		return
	}
	d.slog.Log(d.ctx, slog.LevelWarn, err.Error(), d.toAttrs()...)
}

func (d *logger) Error(err error) {
	d.slog.Error(err.Error(), d.toAttrs()...)
}
