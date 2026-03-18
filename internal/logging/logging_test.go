package logging

import (
	"context"
	"log/slog"
	"sync"
	"testing"
)

type captureHandler struct {
	mu       sync.Mutex
	lastMsg  string
	lastLvl  slog.Level
	hasEntry bool
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastMsg = r.Message
	h.lastLvl = r.Level
	h.hasEntry = true
	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler      { return h }

func TestSetDefault_NilUsesSlogDefault(t *testing.T) {
	old := L()
	t.Cleanup(func() { SetDefault(old) })

	SetDefault(nil)
	if L() != slog.Default() {
		t.Fatalf("expected slog.Default()")
	}
}

func TestSetDefault_CustomLoggerAndWrappers(t *testing.T) {
	old := L()
	t.Cleanup(func() { SetDefault(old) })

	h := &captureHandler{}
	SetDefault(slog.New(h))

	Debug("d")
	if !h.hasEntry || h.lastMsg != "d" || h.lastLvl != slog.LevelDebug {
		t.Fatalf("expected debug record, got msg=%q level=%v", h.lastMsg, h.lastLvl)
	}

	Info("i")
	if h.lastMsg != "i" || h.lastLvl != slog.LevelInfo {
		t.Fatalf("expected info record, got msg=%q level=%v", h.lastMsg, h.lastLvl)
	}

	Warn("w")
	if h.lastMsg != "w" || h.lastLvl != slog.LevelWarn {
		t.Fatalf("expected warn record, got msg=%q level=%v", h.lastMsg, h.lastLvl)
	}

	Error("e")
	if h.lastMsg != "e" || h.lastLvl != slog.LevelError {
		t.Fatalf("expected error record, got msg=%q level=%v", h.lastMsg, h.lastLvl)
	}
}
