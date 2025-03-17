package logger

import (
	"io"
	"os"
	"sync"

	"github.com/rs/zerolog"
)

const BaseLoggerName = "base"

type Logger struct {
	name   string
	mu     *sync.RWMutex
	logger *zerolog.Logger
}

var baseLogger *Logger
var loggerMap map[string]*Logger

func init() {
	_ = InitLogger(os.Stdout)
}

func WrapWriter(w io.Writer, options ...WriterOption) io.Writer {
	for _, option := range options {
		w = option(w)
	}

	return w
}

func InitLogger(w io.Writer, options ...Option) error {
	loggerMap = make(map[string]*Logger, 10)

	logger := zerolog.New(w).With().Timestamp().Logger()
	for _, option := range options {
		logger = option(logger)
	}

	baseLogger = &Logger{
		name:   BaseLoggerName,
		mu:     &sync.RWMutex{},
		logger: &logger,
	}
	loggerMap[BaseLoggerName] = baseLogger
	return nil
}

func NewLogger(name string) *Logger {
	zeroLogger := baseLogger.logger.With().Str("module", name).Logger()
	logger := &Logger{
		name:   name,
		mu:     &sync.RWMutex{},
		logger: &zeroLogger,
	}
	loggerMap[name] = logger

	return logger
}

func SetLevel(name, levelName string) {
	logger, ok := loggerMap[name]
	if !ok {
		return
	}
	logger.SetLevel(levelName)
}

func (l *Logger) SetLevel(levelName string) *Logger {
	level := ParseLogLevel(levelName)
	l.mu.Lock()
	logger := l.logger.Level(level)
	l.logger = &logger
	l.mu.Unlock()
	return l
}

func (l *Logger) Trace() *Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return &Event{event: l.logger.Trace()}
}

func (l *Logger) Debug() *Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return &Event{event: l.logger.Debug()}
}

func (l *Logger) Info() *Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return &Event{event: l.logger.Info()}
}

func (l *Logger) Warn() *Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return &Event{event: l.logger.Warn()}
}

func (l *Logger) Error() *Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return &Event{event: l.logger.Error()}
}

func (l *Logger) Err(err error) *Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return &Event{event: l.logger.Err(err)}
}

func (l *Logger) Fatal() *Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return &Event{event: l.logger.Fatal()}
}

func (l *Logger) Panic() *Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return &Event{event: l.logger.Panic()}
}
