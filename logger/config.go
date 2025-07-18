package logger

type config struct {
	writerOptions []WriterOption
	loggerOptions []LoggerOption
}

type TelegramConfig struct {
	Token     string
	ChatID    int64
	Namespace string
	Subsystem string
}
