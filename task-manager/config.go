package taskmanager

import (
	"github.com/go-co-op/gocron/v2"
	"go.opentelemetry.io/otel/trace"
)

type config struct {
	tracer           trace.Tracer
	schedulerOptions []gocron.SchedulerOption
	jobOptions       []gocron.JobOption
}

func newConfig() *config {
	return &config{}
}
