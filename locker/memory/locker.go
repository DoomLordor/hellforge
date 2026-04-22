package memory

import (
	"context"
	"sync"

	"github.com/rs/zerolog"

	"github.com/DoomLordor/hellforge/locker"
	"github.com/DoomLordor/hellforge/logger"
)

type lockEntry struct {
	ch   chan struct{}
	refs int
}

type memoryLocker struct {
	logger zerolog.Logger
	mu     sync.Mutex
	locks  map[string]*lockEntry
}

func NewLocker() locker.Locker {
	return &memoryLocker{
		logger: logger.NewLogger("memory-locker"),
		locks:  make(map[string]*lockEntry),
	}
}

func (l *memoryLocker) Lock(ctx context.Context, task *locker.LockKey) error {
	l.mu.Lock()
	entry, ok := l.locks[task.Key]
	if !ok {
		entry = &lockEntry{ch: make(chan struct{}, 1)}
		l.locks[task.Key] = entry
	}
	entry.refs++
	l.mu.Unlock()

	l.logger.Info().Str("key", task.Key).Msg("trying to hold lock")
	select {
	case entry.ch <- struct{}{}:
	case <-ctx.Done():
		l.deleteKey(task.Key)
		return ctx.Err()
	}

	if task.OnStart != nil {
		err := task.OnStart()
		if err != nil {
			<-entry.ch
			l.deleteKey(task.Key)
			return err
		}
	}

	go func() {
		defer l.deleteKey(task.Key)
		<-ctx.Done()
		if task.OnStop != nil {
			err := task.OnStop()
			if err != nil {
				l.logger.Err(err).Str("key", task.Key).Msg("error in OnStop")
			}
		}
		<-entry.ch
	}()

	return nil
}

func (l *memoryLocker) deleteKey(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.locks[key]
	if ok {
		entry.refs--
		if entry.refs == 0 {
			delete(l.locks, key)
		}
	}
}
