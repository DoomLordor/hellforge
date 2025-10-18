package defaults

import (
	"os"

	"github.com/DoomLordor/hellforge/logger"
)

type Logger struct {
	WithConsoleWriter bool   `env:"WITH_CONSOLE_WRITER" envDefault:"false"`
	Level             string `env:"LEVEL" envDefault:"info"`
	FrameCount        int    `env:"FRAME_COUNT" envDefault:"0"`

	TelegramEnabled   bool   `env:"TELEGRAM_ENABLED" envDefault:"false"`
	TelegramToken     string `env:"TELEGRAM_TOKEN" envDefault:""`
	TelegramChatID    int64  `env:"TELEGRAM_CHAT_ID" envDefault:"0"`
	TelegramNamespace string `env:"TELEGRAM_NAMESPACE" envDefault:""`
	TelegramSubsystem string `env:"TELEGRAM_SUBSYSTEM" envDefault:""`
}

func (c *Logger) ToLoggerConfig() logger.TelegramConfig {
	return logger.TelegramConfig{
		Token:     c.TelegramToken,
		ChatID:    c.TelegramChatID,
		Namespace: c.TelegramNamespace,
		Subsystem: c.TelegramSubsystem,
	}
}

func (c *Logger) Init() {
	writerOptions := make([]logger.WriterOption, 0, 2)
	if c.WithConsoleWriter {
		writerOptions = append(writerOptions, logger.WithConsoleWriter())
	}

	if c.TelegramEnabled {
		writerOptions = append(writerOptions, logger.WithTelegramWriter(c.ToLoggerConfig()))
	}

	loggerOptions := make([]logger.LoggerOption, 0, 3)
	loggerOptions = append(loggerOptions, logger.WithLevel(c.Level), logger.WithTraceID())
	if c.FrameCount != 0 {
		loggerOptions = append(loggerOptions, logger.WithCallerFrameCount(c.FrameCount))
	}

	logger.InitLogger(
		os.Stderr,
		logger.WithWriterOptions(writerOptions...),
		logger.WithLoggerOptions(loggerOptions...),
	)
}
