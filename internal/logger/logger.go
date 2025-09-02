package logger

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func init() {
	// Configure handler (console, JSON, custom, etc.)
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, // default log level
	})
	Logger = slog.New(handler)

	slog.SetDefault(Logger)
}
