package configs

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Host           string `env:"HOST" envDefault:"localhost"`
	Port           uint16 `env:"PORT" envDefault:"5432"`
	User           string `env:"USER" envDefault:"postgres"`
	Password       string `env:"PASSWORD" envDefault:"12345"`
	Name           string `env:"NAME" envDefault:"postgres"`
	SslMode        string `env:"SSL_MODE" envDefault:"disable"`
	MaxConnections int32  `env:"MAX_CONNECTIONS" envDefault:"10"`
}

func (c *Postgres) DSN() string {
	format := "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Europe/Moscow"
	return fmt.Sprintf(format, c.Host, c.Port, c.User, c.Password, c.Name, c.SslMode)
}

func (c *Postgres) Pool(ctx context.Context) (*pgxpool.Pool, error) {
	postgresConfig, err := pgxpool.ParseConfig(c.DSN())
	if err != nil {
		return nil, errors.New("failed to parse pg dsn config string")
	}

	postgresConfig.MaxConns = c.MaxConnections

	postgresPool, err := pgxpool.NewWithConfig(ctx, postgresConfig)
	if err != nil {
		return nil, err
	}

	err = postgresPool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return postgresPool, nil
}
