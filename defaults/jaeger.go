package defaults

import (
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/DoomLordor/hellforge/tracer"
)

type Jaeger struct {
	GrpcAddr    string `env:"GRPC_ADDR" envDefault:"localhost:4317"`
	ServiceName string `env:"SERVICE_NAME"`
}

func (c *Jaeger) TracerProvider() (*sdktrace.TracerProvider, error) {
	return tracer.NewJaegerClient(c.GrpcAddr, c.ServiceName)
}
