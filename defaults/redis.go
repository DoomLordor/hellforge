package defaults

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	"github.com/DoomLordor/hellforge/redis"
)

type Redis struct {
	Address  []string `env:"address" envDefault:"localhost:6379"`
	Password string   `env:"password" envDefault:""`
}

func (c *Redis) ClusterClient(ctx context.Context, provider trace.TracerProvider) (redis.ClusterClient, error) {
	return redis.NewClusterClient(ctx, c.Address, c.Password, redis.WithTracing(provider))
}
