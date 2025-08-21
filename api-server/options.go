package apiserver

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type Option func(c *config)

func WithHttpSystem(enabled bool) Option {
	return func(c *config) {
		c.httpSystem.enabled = enabled
	}
}

func WithHttpSystemPort(port uint16) Option {
	return func(c *config) {
		c.httpSystem.port = port
	}
}

func WithMetrics(metrics ...prometheus.Collector) Option {
	return func(c *config) {
		c.metrics = append(c.metrics, metrics...)
	}
}

func WithHttp(enabled bool) Option {
	return func(c *config) {
		c.http.enabled = enabled
	}
}

func WithHttpHost(host string) Option {
	return func(c *config) {
		c.http.host = host
	}
}

func WithHttpPort(port uint16) Option {
	return func(c *config) {
		c.http.port = port
	}
}

func WithErrorResponseConstructor(erc ErrorResponseConstructor) Option {
	return func(c *config) {
		c.http.erc = erc
	}
}

func WithApi(api ...Api) Option {
	return func(c *config) {
		c.http.api = append(c.http.api, api...)
	}
}

func WithGrpc(enabled bool) Option {
	return func(c *config) {
		c.grpc.enabled = enabled
	}
}

func WithGrpcHost(host string) Option {
	return func(c *config) {
		c.grpc.host = host
	}
}

func WithGPRCPort(port uint16) Option {
	return func(c *config) {
		c.grpc.port = port
	}
}

func WithGrpcImplementation(implementations ...Grpc) Option {
	return func(c *config) {
		c.grpc.implementations = append(c.grpc.implementations, implementations...)
	}
}

func WithGrpcReflection(enabled bool) Option {
	return func(c *config) {
		c.grpc.withReflect = enabled
	}
}

func WithGrpcUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) Option {
	return func(c *config) {
		c.grpc.unaryInterceptors = append(c.grpc.unaryInterceptors, interceptors...)
	}
}

func WithGrpcStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) Option {
	return func(c *config) {
		c.grpc.streamInterceptors = append(c.grpc.streamInterceptors, interceptors...)
	}
}

func WithGrpcMuxOptions(muxOptions ...runtime.ServeMuxOption) Option {
	return func(c *config) {
		c.grpc.muxOptions = append(c.grpc.muxOptions, muxOptions...)
	}
}

func WithClosers(closers ...func(ctx context.Context) error) Option {
	return func(c *config) {
		c.closers = append(c.closers, closers...)
	}
}

func WithTracer(tracer trace.Tracer) Option {
	return func(c *config) {
		c.tracer = tracer
	}
}
