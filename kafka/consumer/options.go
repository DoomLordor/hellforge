package consumer

import (
	"go.opentelemetry.io/otel/trace"

	"github.com/DoomLordor/hellforge/kafka"
)

type Option func(c *config)

func WithTracer(tracer trace.Tracer) Option {
	return func(c *config) {
		c.tracer = tracer
	}
}

func WithKafkaOptions(options ...kafka.Option) Option {
	return func(c *config) {
		c.kafkaOptions = append(c.kafkaOptions, options...)
	}
}
