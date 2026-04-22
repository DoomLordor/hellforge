package taskmanager

import (
	"github.com/go-co-op/gocron/v2"
	"go.opentelemetry.io/otel/trace"
)

type Option func(c *config)

func WithTracer(provider trace.TracerProvider) Option {
	return func(c *config) {
		c.provider = provider
	}
}

func WithSchedulerOptions(schedulerOptions ...gocron.SchedulerOption) Option {
	return func(c *config) {
		c.schedulerOptions = append(c.schedulerOptions, schedulerOptions...)
	}
}

func WithJobOptions(jobOptions ...gocron.JobOption) Option {
	return func(c *config) {
		c.jobOptions = append(c.jobOptions, jobOptions...)
	}
}
