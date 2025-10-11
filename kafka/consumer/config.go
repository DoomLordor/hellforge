package consumer

import (
	"go.opentelemetry.io/otel/trace"

	"github.com/DoomLordor/hellforge/kafka"
)

type config struct {
	tracer       trace.Tracer
	kafkaOptions []kafka.Option
}

func newConfig() *config {
	return &config{}
}
