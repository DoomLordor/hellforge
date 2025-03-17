package logger

import (
	"strings"

	"github.com/rs/zerolog"
)

func ParseLogLevel(level string) zerolog.Level {
	level = strings.ToLower(level)
	switch level {
	default:
		return zerolog.InfoLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "no":
		return zerolog.NoLevel
	case "disable":
		return zerolog.Disabled
	case "trace":
		return zerolog.TraceLevel
	}
}
