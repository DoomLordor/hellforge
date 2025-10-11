package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

const BaseLoggerName = "base"

var baseLogger *zerolog.Logger

func init() {
	InitLogger(os.Stdout)
}

func InitLogger(w io.Writer, options ...Option) {
	c := &config{}
	for _, opt := range options {
		opt(c)
	}

	for _, option := range c.writerOptions {
		w = option(w)
	}

	logger := zerolog.New(w).With().Timestamp().Logger()
	for _, option := range c.loggerOptions {
		logger = option(logger)
	}

	baseLogger = &logger
}

func NewLogger(name string) zerolog.Logger {
	return baseLogger.With().Str("module", name).Logger()
}
