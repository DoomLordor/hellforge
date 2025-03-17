package apiserver

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	"github.com/DoomLordor/hellforge/apiserver/grpc"
	"github.com/DoomLordor/hellforge/apiserver/rest"
)

type Adapter struct {
	Auth   rest.AuthFunc
	Api    []rest.Api
	Grps   []grpc.Grps
	Tracer trace.Tracer
}

type Configurator interface {
	Configure(ctx context.Context) (*Adapter, error)
}

type Shutdown interface {
	Stop(ctx context.Context) []error
}
