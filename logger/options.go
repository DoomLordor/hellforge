package logger

import (
	"io"

	"github.com/rs/zerolog"
)

type WriterOption func(w io.Writer) io.Writer
type Option func(logger zerolog.Logger) zerolog.Logger

// WithConsoleWriter wrap writer with human-friendly output format
func WithConsoleWriter() WriterOption {
	return func(w io.Writer) io.Writer {
		return zerolog.ConsoleWriter{Out: w, TimeFormat: zerolog.TimeFieldFormat}
	}
}

// WithTelegramWriter wrap writer with telegram sender
func WithTelegramWriter(config TelegramConfig) WriterOption {
	return func(w io.Writer) io.Writer {
		w, err := newTelegramWriter(w, config)
		if err != nil {
			panic(err)
		}

		return w
	}
}

// WithCallerFrameCount adds the file:line of the caller
func WithCallerFrameCount(frameCount int) Option {
	return func(logger zerolog.Logger) zerolog.Logger {
		return logger.With().CallerWithSkipFrameCount(frameCount).Logger()
	}
}

// WithLevel set log level
func WithLevel(level string) Option {
	return func(logger zerolog.Logger) zerolog.Logger {
		return logger.Level(ParseLogLevel(level))
	}
}
