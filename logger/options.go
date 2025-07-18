package logger

import (
	"io"

	"github.com/rs/zerolog"
)

type Option func(c *config)

type WriterOption func(w io.Writer) io.Writer
type LoggerOption func(logger zerolog.Logger) zerolog.Logger

// WithWriterOptions add writer options to logger config
func WithWriterOptions(options ...WriterOption) Option {
	return func(c *config) {
		c.writerOptions = append(c.writerOptions, options...)
	}
}

// WithLoggerOptions add logger options to logger config
func WithLoggerOptions(options ...LoggerOption) Option {
	return func(c *config) {
		c.loggerOptions = append(c.loggerOptions, options...)
	}
}

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
func WithCallerFrameCount(frameCount int) LoggerOption {
	return func(logger zerolog.Logger) zerolog.Logger {
		return logger.With().CallerWithSkipFrameCount(frameCount).Logger()
	}
}

// WithLevel set log level
func WithLevel(level string) LoggerOption {
	return func(logger zerolog.Logger) zerolog.Logger {
		return logger.Level(ParseLogLevel(level))
	}
}

// WithTraceID add trace id to logger output
func WithTraceID() LoggerOption {
	return func(logger zerolog.Logger) zerolog.Logger {
		return logger.Hook(&tracingHook{})
	}
}
