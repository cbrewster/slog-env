package testpackage

import (
	"context"
	"log/slog"
)

func LogSomething(logger *slog.Logger, level slog.Level, message string) {
	logger.Log(context.Background(), level, message)
}
