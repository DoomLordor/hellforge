package taskmanager

import (
	"errors"
)

var (
	SchedulerError    = errors.New("scheduler error")
	TaskNotFoundError = errors.New("task not found")
)
