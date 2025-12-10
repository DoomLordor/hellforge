package gateway

import (
	"runtime"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type config struct {
	tracer         trace.Tracer
	adapters       []Adapter
	dialOpts       []grpc.DialOption
	limit          int
	maxMessageSize int
}

func newConfig() *config {
	return &config{
		limit:          runtime.NumCPU(),
		maxMessageSize: 1024 * 1024 * 10,
	}
}

func (c *config) getAdaptersMap() map[string]Adapter {
	adapterMap := make(map[string]Adapter, len(c.adapters))
	for _, adapter := range c.adapters {
		adapterMap[adapter.GetMethod()] = adapter
	}

	return adapterMap
}
