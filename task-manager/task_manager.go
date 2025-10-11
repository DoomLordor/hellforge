package taskmanager

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/DoomLordor/hellforge/logger"
)

//TODO: add leader election

type TaskManager struct {
	ctx       context.Context
	mu        *sync.Mutex
	logger    zerolog.Logger
	cfg       *config
	scheduler gocron.Scheduler
	tracer    trace.Tracer
	tasks     map[string]*task
}

func NewTaskManager(ctx context.Context, enable bool, options ...Option) (*TaskManager, error) {
	cfg := newConfig()
	for _, option := range options {
		option(cfg)
	}

	scheduler, err := gocron.NewScheduler(cfg.schedulerOptions...)
	if err != nil {
		return nil, err
	}

	if enable {
		scheduler.Start()
	}

	return &TaskManager{
		ctx:       ctx,
		mu:        &sync.Mutex{},
		logger:    logger.NewLogger("task-manager"),
		cfg:       cfg,
		scheduler: scheduler,
		tracer:    cfg.tracer,
		tasks:     make(map[string]*task, 10),
	}, nil
}

func (m *TaskManager) AddTask(taskName string, cronSchedule string, worker Worker) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if t, ok := m.tasks[taskName]; ok {
		err := m.scheduler.RemoveJob(t.job.ID())
		if err != nil {
			m.logger.Err(err).Msg("failed to remove task")
		}
	}

	t := &task{
		name:         taskName,
		cronSchedule: cronSchedule,
		worker:       m.basicDecorator(m.ctx, worker, taskName),
	}

	job, err := m.scheduler.NewJob(
		gocron.CronJob(t.cronSchedule, true),
		gocron.NewTask(t.worker),
		m.cfg.jobOptions...,
	)
	if err != nil {
		m.logger.Err(err).Msg("failed to create job")
		return SchedulerError
	}

	t.job = job
	m.tasks[taskName] = t
	return nil
}

func (m *TaskManager) StopTask(taskName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tasks[taskName]
	if !ok {
		return TaskNotFoundError
	}

	delete(m.tasks, taskName)
	err := m.scheduler.RemoveJob(t.job.ID())
	if err != nil {
		m.logger.Err(err).Msg("failed to stop task")
		return SchedulerError
	}

	return nil
}

func (m *TaskManager) Shutdown() error {
	return m.scheduler.Shutdown()
}

func (m *TaskManager) StopJobs() error {
	return m.scheduler.StopJobs()
}

func (m *TaskManager) withRecover(worker Worker) Worker {
	return func(ctx context.Context) error {
		defer func() {
			r := recover()
			if r != nil {
				m.logger.Err(errors.New("panic")).
					Ctx(ctx).
					Str("panic", fmt.Sprintf("%v", r)).
					Msgf("recovered stack: %s", string(debug.Stack()))
			}
		}()

		return worker(ctx)
	}
}

func (m *TaskManager) withTracing(worker Worker, taskName string) Worker {
	if m.tracer == nil {
		return worker
	}

	return func(ctx context.Context) error {
		ctx, span := m.tracer.Start(ctx, taskName)
		defer span.End()

		err := worker(ctx)
		if err != nil {
			span.SetStatus(otelcodes.Error, err.Error())
		} else {
			span.SetStatus(otelcodes.Ok, "succeeded")
		}

		return err
	}
}

func (m *TaskManager) basicDecorator(ctx context.Context, worker Worker, taskName string) func() {
	worker = m.withRecover(worker)
	worker = m.withTracing(worker, taskName)

	return func() {
		var cancel func()
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()

		m.logger.Debug().
			Ctx(ctx).
			Str("task", taskName).
			Msgf("start periodic task")

		err := worker(ctx)
		if err != nil {
			m.logger.Err(err).
				Ctx(ctx).
				Str("task", taskName).
				Msgf("error in periodic task")
		}

		m.logger.Debug().
			Ctx(ctx).
			Str("task", taskName).
			Msg("stop periodic task")
	}
}
