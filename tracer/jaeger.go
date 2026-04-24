package tracer

import (
	"context"
	"crypto/tls"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc/credentials"
)

// NewJaegerClient initializes an OTLP exporter
func NewJaegerClient(ctx context.Context, address, name string, withTLS bool) (*sdktrace.TracerProvider, error) {
	credential := otlptracegrpc.WithInsecure()
	if withTLS {
		credential = otlptracegrpc.WithTLSCredentials(credentials.NewTLS(&tls.Config{}))
	}

	// Set up the OTLP trace exporter
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(address),
		credential,
	)
	if err != nil {
		return nil, err
	}

	// Create a new batch span processor
	spanProcessor := sdktrace.NewBatchSpanProcessor(exporter)

	rsc := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(name),
	)

	// Create new trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(spanProcessor),
		sdktrace.WithResource(rsc),
	)

	// Set the global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
	))

	// Set the global tracer provider
	otel.SetTracerProvider(tp)

	return tp, nil
}
