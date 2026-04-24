package defaults

import (
	"context"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/DoomLordor/hellforge/tracer"
)

type Jaeger struct {
	GrpcAddr    string `env:"GRPC_ADDR" envDefault:"localhost:4317"`
	ServiceName string `env:"SERVICE_NAME"`
	WithTLS     bool   `env:"WITH_TLS" envDefault:"false"`
}

func (c *Jaeger) TracerProvider(ctx context.Context) (*sdktrace.TracerProvider, error) {
	return tracer.NewJaegerClient(ctx, c.GrpcAddr, c.ServiceName, c.WithTLS)
}
