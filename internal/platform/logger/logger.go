package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog.Logger for structured application logging.
type Logger struct {
	zerolog.Logger
}

// New builds a Logger with sane defaults.
func New(appName, environment string) Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	base := zerolog.New(output).With().Timestamp().Str("app", appName).Str("env", environment).Logger()
	return Logger{Logger: base}
}

// WithOutput allows overriding the logger output, useful for tests.
func (l Logger) WithOutput(w io.Writer) Logger {
	output := zerolog.ConsoleWriter{Out: w, TimeFormat: time.RFC3339}
	logger := zerolog.New(output).With().Timestamp().Logger()
	return Logger{Logger: logger}
}
