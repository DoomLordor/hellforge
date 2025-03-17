package logger

import (
	"fmt"
	"io"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
)

type telegramWriter struct {
	baseWriter io.Writer
	bot        *tgbotapi.BotAPI
	chatID     int64
	namespace  string
	subsystem  string
}

func newTelegramWriter(w io.Writer, cfg TelegramConfig) (io.Writer, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}

	writer := &telegramWriter{
		baseWriter: w,
		bot:        bot,
		chatID:     cfg.ChatID,
		namespace:  cfg.Namespace,
		subsystem:  cfg.Subsystem,
	}

	return writer, nil
}

func (t *telegramWriter) Write(p []byte) (n int, err error) {
	return t.baseWriter.Write(p)
}

func (t *telegramWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	switch level {
	case zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel:
		text := fmt.Sprintf(`Namespace[%s] Subsystem[%s] message: %s`, t.namespace, t.subsystem, p)
		message := tgbotapi.NewMessage(t.chatID, text)
		_, err := t.bot.Send(message)
		if err != nil {
			return 0, err
		}
	default:
	}

	return t.baseWriter.Write(p)
}
