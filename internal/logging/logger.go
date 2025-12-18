package logging

import (
	"log/slog"
	"os"
)

// NewLogger builds a JSON structured logger with timestamps and source info.
func NewLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})
	return slog.New(handler)
}
