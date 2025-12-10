package apiserver

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	"github.com/DoomLordor/hellforge/nats-gateway/gateway"
)

type Configurator func(ctx context.Context) ([]Option, error)

type Grpc interface {
	RegisterServer(grpcServer *grpc.Server)
	RegisterHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error
	RegisterHandlerNats() []gateway.Adapter
}

type Api interface {
	RegistrationApi() RouteApiMap
}
