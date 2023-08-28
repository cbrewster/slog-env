package slogenv_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	slogenv "github.com/cbrewster/slog-env"
	"github.com/cbrewster/slog-env/internal/testpackage"
)

// testHandler provides a simple log handler which just records logs messages.
type testHandler struct {
	messages []string
}

// Enabled implements slog.Handler.
func (*testHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

// Handle implements slog.Handler.
func (h *testHandler) Handle(ctx context.Context, record slog.Record) error {
	h.messages = append(h.messages, record.Message)
	return nil
}

// WithAttrs implements slog.Handler.
func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

// WithGroup implements slog.Handler.
func (h *testHandler) WithGroup(name string) slog.Handler {
	return h
}

var _ slog.Handler = (*testHandler)(nil)

// TestDefaultLevel tests setting just the default level.
func TestDefaultLevel(t *testing.T) {
	for _, test := range []struct {
		name         string
		filter       string
		wantMessages []string
	}{
		{
			filter:       "debug",
			wantMessages: []string{"debug", "info", "warn", "error"},
		},
		{
			filter:       "info",
			wantMessages: []string{"info", "warn", "error"},
		},
		{
			filter:       "warn",
			wantMessages: []string{"warn", "error"},
		},
		{
			filter:       "error",
			wantMessages: []string{"error"},
		},
	} {
		t.Run(test.filter, func(t *testing.T) {
			os.Setenv("GO_LOG", test.filter)
			defer os.Unsetenv("GO_LOG")

			h := testHandler{}
			logger := slog.New(slogenv.NewHandler(&h))
			logger.Debug("debug")
			logger.Info("info")
			logger.Warn("warn")
			logger.Error("error")

			assert.Equal(t, test.wantMessages, h.messages)
		})
	}
}

// TestPackageFilter tests both the default level and package filter.
func TestPackageFilter(t *testing.T) {
	for _, test := range []struct {
		filter       string
		wantMessages []string
	}{
		{
			filter:       "testpackage=error",
			wantMessages: []string{"info", "warn", "error", "testpackage error"},
		},
		{
			filter:       "debug,testpackage=error",
			wantMessages: []string{"debug", "info", "warn", "error", "testpackage error"},
		},
		{
			filter:       "slog-env_test=debug,testpackage=error",
			wantMessages: []string{"debug", "info", "warn", "error", "testpackage error"},
		},
		{
			filter:       "error,testpackage=debug",
			wantMessages: []string{"error", "testpackage debug", "testpackage info", "testpackage warn", "testpackage error"},
		},
	} {
		t.Run(test.filter, func(t *testing.T) {
			os.Setenv("GO_LOG", test.filter)
			defer os.Unsetenv("GO_LOG")

			h := testHandler{}
			logger := slog.New(slogenv.NewHandler(&h))
			logger.Debug("debug")
			logger.Info("info")
			logger.Warn("warn")
			logger.Error("error")
			testpackage.LogSomething(logger, slog.LevelDebug, "testpackage debug")
			testpackage.LogSomething(logger, slog.LevelInfo, "testpackage info")
			testpackage.LogSomething(logger, slog.LevelWarn, "testpackage warn")
			testpackage.LogSomething(logger, slog.LevelError, "testpackage error")

			assert.Equal(t, test.wantMessages, h.messages)
		})
	}
}
