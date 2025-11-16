package logger

import (
	"log/slog"
	"os"
	"pr-reviewer-service/config"
	"strings"
)

func New(cfg *config.ServerConfig) *slog.Logger {
	level := parseLevel(cfg.LogLevel)

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

func parseLevel(levelStr string) slog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
