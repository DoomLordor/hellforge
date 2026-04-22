package postgres

import (
	"context"

	"cirello.io/pglock"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"

	"github.com/DoomLordor/hellforge/locker"
	"github.com/DoomLordor/hellforge/logger"
)

type postgresLocker struct {
	logger zerolog.Logger
	client *pglock.Client
}

func NewLocker(pool *pgxpool.Pool, options ...Option) (locker.Locker, error) {
	cfg := newConfig()
	for _, option := range options {
		option(cfg)
	}

	client, err := pglock.UnsafeNew(
		stdlib.OpenDBFromPool(pool),
		pglock.WithLeaseDuration(cfg.leaseDuration),
		pglock.WithHeartbeatFrequency(cfg.heartbeatFrequency),
		pglock.WithCustomTable(cfg.tableName),
	)
	if err != nil {
		return nil, err
	}

	if cfg.tryCreateTable {
		err = client.TryCreateTable()
		if err != nil {
			return nil, err
		}
	}

	return &postgresLocker{
		logger: logger.NewLogger("postgres-locker"),
		client: client,
	}, nil
}

func (l *postgresLocker) Lock(ctx context.Context, task *locker.LockKey) error {
	l.logger.Debug().Str("key", task.Key).Msg("trying to hold pg lock")

	lock, err := l.client.AcquireContext(ctx, task.Key)
	if err != nil {
		l.logger.Err(err).Str("key", task.Key).Msg("failed to obtain pg lock")
		return err
	}

	l.logger.Debug().Str("key", task.Key).Msg("obtained pg lock")

	if task.OnStart != nil {
		err = task.OnStart()
		if err != nil {
			_ = lock.Close()
			return err
		}
	}

	go l.monitorLock(ctx, lock, task)

	return nil
}

func (l *postgresLocker) monitorLock(ctx context.Context, lock *pglock.Lock, task *locker.LockKey) {
	defer func() {
		err := lock.Close()
		if err != nil {
			l.logger.Err(err).Str("key", task.Key).Msg("failed to close lock")
		}
	}()

	select {
	case <-ctx.Done():
		l.logger.Debug().Str("key", task.Key).Msg("context done, releasing lock")
	}

	if task.OnStop != nil {
		err := task.OnStop()
		if err != nil {
			l.logger.Err(err).Str("key", task.Key).Msg("error in OnStop")
		}
	}
}
