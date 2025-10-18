package configs

import (
	"github.com/caarlos0/env/v10"
)

func NewConfig[T any]() (*T, error) {
	var cfg T
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
