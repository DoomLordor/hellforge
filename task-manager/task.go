package taskmanager

import (
	"context"

	"github.com/go-co-op/gocron/v2"
)

type Worker func(ctx context.Context) error

type task struct {
	name         string
	cronSchedule string
	worker       func()
	job          gocron.Job
}
