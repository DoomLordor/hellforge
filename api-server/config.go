package apiserver

import (
	"context"
	"fmt"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type config struct {
	httpSystem httpSystemConfig
	http       httpConfig
	grpc       grpcConfig
	tracer     trace.Tracer
	metrics    []prometheus.Collector
	closers    []func(ctx context.Context) error
}

type httpSystemConfig struct {
	enabled bool
	port    uint16
}

type httpConfig struct {
	enabled         bool
	host            string
	port            uint16
	writeTimeout    time.Duration
	readTimeout     time.Duration
	idleTimeout     time.Duration
	swaggerFile     string
	baseSwaggerPath string
	api             []Api
	erc             ErrorResponseConstructor
}

type grpcConfig struct {
	enabled            bool
	gatewayEnabled     bool
	host               string
	port               uint16
	implementations    []GRPC
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	muxOptions         []runtime.ServeMuxOption
	withReflect        bool
}

type jaegerConfig struct {
	address string
	name    string
}

func newConfig() *config {
	return &config{
		httpSystem: httpSystemConfig{
			enabled: true,
			port:    8081,
		},
		http: httpConfig{
			enabled:         true,
			port:            8080,
			writeTimeout:    15 * time.Second,
			readTimeout:     15 * time.Second,
			idleTimeout:     15 * time.Second,
			swaggerFile:     "swagger.json",
			baseSwaggerPath: "/",
		},
		grpc: grpcConfig{
			enabled:        true,
			gatewayEnabled: true,
			port:           8000,
			withReflect:    true,
		},
	}
}

func (c *httpSystemConfig) address() string {
	return fmt.Sprintf(":%d", c.port)
}

func (c *httpConfig) address() string {
	return fmt.Sprintf("%s:%d", c.host, c.port)
}

func (c *grpcConfig) address() string {
	return fmt.Sprintf("%s:%d", c.host, c.port)
}
