package helpers

import (
	"go.opentelemetry.io/otel/trace"
)

func ProviderToTracer(provider trace.TracerProvider, name string, opts ...trace.TracerOption) trace.Tracer {
	if provider == nil {
		return nil
	}

	return provider.Tracer(name, opts...)
}
