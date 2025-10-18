package defaults

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type Clickhouse struct {
	Addresses []string `env:"ADDRESSES" envDefault:"localhost:9000"`
	User      string   `env:"USER" envDefault:"default"`
	Password  string   `env:"PASSWORD" envDefault:""`
	Name      string   `env:"NAME" envDefault:"default"`
	TLS       bool     `env:"TLS" envDefault:"false"`
}

func (c *Clickhouse) Open(ctx context.Context) (clickhouse.Conn, error) {
	options := &clickhouse.Options{
		Addr: c.Addresses,
		Auth: clickhouse.Auth{
			Database: c.Name,
			Username: c.User,
			Password: c.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: time.Second * 30,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Debug:                false,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
	}

	if c.TLS {
		options.TLS = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	conn, err := clickhouse.Open(options)
	if err != nil {
		return nil, err
	}

	err = conn.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
