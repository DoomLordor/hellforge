package redis

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"

	"github.com/DoomLordor/hellforge/locker"
	"github.com/DoomLordor/hellforge/logger"
	"github.com/DoomLordor/hellforge/redis"
)

type redisLocker struct {
	logger     zerolog.Logger
	client     redisClient
	ttl        time.Duration
	refreshTTL time.Duration
	retryTTL   time.Duration
	errorCount int
}

// NewLocker creates a new distributed lock
func NewLocker(client redisClient, options ...Option) locker.Locker {
	cfg := newConfig()
	for _, option := range options {
		option(cfg)
	}

	return &redisLocker{
		logger:     logger.NewLogger("redis-locker"),
		client:     client,
		ttl:        cfg.ttl,
		refreshTTL: cfg.refreshTTL,
		retryTTL:   cfg.retryTTL,
		errorCount: cfg.errorCount,
	}
}

// Lock key with callback on obtain/lose lock
func (l *redisLocker) Lock(ctx context.Context, task *locker.LockKey) error {
	for {
		l.logger.Info().Str("key", task.Key).Msg("trying to hold lock")

		ok, err := l.client.SetNX(ctx, task.Key, "lockerValue", l.ttl)
		if err == nil && ok {
			l.logger.Info().Str("key", task.Key).Msg("obtained lock")

			if task.OnStart != nil {
				l.logger.Info().Str("key", task.Key).Msg("started task")
				err = task.OnStart()
				if err != nil {
					l.logger.Err(err).Str("key", task.Key).Msg("error while starting task")
					_ = l.client.Del(ctx, task.Key)

					return err
				}
			}

			go l.holdLock(ctx, task)

			return nil
		}

		retryTimer := time.NewTimer(l.retryTTL)

		select {
		case <-retryTimer.C:
			continue
		case <-ctx.Done():
			l.logger.Err(errors.New("main context done")).Str("key", task.Key).Msg("error obtaining redis lock")
			return nil
		}
	}
}

func (l *redisLocker) holdLock(ctx context.Context, task *locker.LockKey) {
	ticker := time.NewTicker(l.refreshTTL)
	errCounter := 0

	for {
		select {
		case <-ticker.C:
			err := l.client.Expire(ctx, task.Key, l.ttl)
			if err != nil {
				l.logger.Err(err).Str("key", task.Key).Msg("error while holding lock")
				errCounter++

				if errCounter >= l.errorCount {
					if task.OnStop != nil {
						l.logger.Debug().Str("key", task.Key).Msg("doing on error task")

						err = task.OnStop()
						if err != nil {
							l.logger.Err(err).Str("key", task.Key).Msg("error while doing on error task")
							return
						}
					}

					l.logger.Err(err).Str("key", task.Key).Msg("stopped holding lock, going to retry lock")

					go func() {
						err := l.Lock(ctx, task)
						if err != nil {
							l.logger.Err(err).Str("key", task.Key).Msg("error while retrying lock")
						}
					}()

					return
				}
				continue
			}

			errCounter = 0

		case <-ctx.Done():
			if task.OnStop != nil {
				l.logger.Debug().Str("key", task.Key).Msg("doing on stop task")

				err := task.OnStop()
				if err != nil {
					l.logger.Err(err).Str("key", task.Key).Msg("error while doing on stop task")
				}
			}

			delCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			err := l.client.Del(delCtx, task.Key)
			if err != nil && !(errors.Is(err, redis.ErrNotSet) && errors.Is(err, redis.Nil)) {
				l.logger.Err(err).Str("key", task.Key).Msg("error while releasing lock")
				return
			}

			l.logger.Debug().Str("key", task.Key).Msg("deleted lock")
			return
		}
	}
}
