// Package logger provides a configured slog.Logger instance with JSON output.
package logger

import (
	"log/slog"
	"os"
)

// New creates a new *slog.Logger configured with the given service name.
//
// The logger uses a JSON handler writing to os.Stdout. Every log entry
// includes the service name via slog.String("service", serviceName).
//
// The log level is determined by the APP_ENV environment variable:
//   - If APP_ENV equals "production", the level is set to slog.LevelInfo.
//   - Otherwise, the level is set to slog.LevelDebug.
func New(serviceName string) *slog.Logger {
	level := slog.LevelDebug
	if os.Getenv("APP_ENV") == "production" {
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	logger := slog.New(handler).With(slog.String("service", serviceName))

	return logger
}
