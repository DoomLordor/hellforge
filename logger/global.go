package logger

func Trace(message string) {
	baseLogger.Trace().Msg(message)
}

func Tracef(message string, args ...any) {
	baseLogger.Trace().Msgf(message, args...)
}

func Debug(message string) {
	baseLogger.Debug().Msg(message)
}

func Debugf(message string, args ...any) {
	baseLogger.Debug().Msgf(message, args...)
}

func Info(message string) {
	baseLogger.Info().Msg(message)
}

func Infof(message string, args ...any) {
	baseLogger.Info().Msgf(message, args...)
}

func Warn(message string) {
	baseLogger.Warn().Msg(message)
}

func Warnf(message string, args ...any) {
	baseLogger.Warn().Msgf(message, args...)
}

func Error(err error, message string) {
	baseLogger.Err(err).Msg(message)
}

func Errorf(err error, message string, args ...any) {
	baseLogger.Err(err).Msgf(message, args...)
}

func Fatal(message string) {
	baseLogger.Fatal().Msg(message)
}

func Fatalf(message string, args ...any) {
	baseLogger.Fatal().Msgf(message, args...)
}
