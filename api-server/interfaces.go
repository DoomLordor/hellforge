package apiserver

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type Configurator interface {
	Configurate(ctx context.Context) ([]Option, error)
}

type Grpc interface {
	RegisterServer(grpcServer *grpc.Server)
	RegisterHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error
}

type Api interface {
	RegistrationApi() RouteApiMap
}
