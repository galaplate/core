package scheduler

import (
	"context"

	"github.com/galaplate/core/logger"
	"github.com/robfig/cron/v3"
)

type Handler interface {
	Handle() (string, func())
}

type Scheduler struct {
	cron    *cron.Cron
	started bool
}

func New() *Scheduler {
	return &Scheduler{
		cron:    cron.New(cron.WithSeconds()),
		started: false,
	}
}

func (s *Scheduler) RunTasks() error {
	for name, task := range SchedulerRegistry {
		_, err := s.AddTask(task.Handle())
		if err != nil {
			logger.Error("Scheduler@RunTasks", map[string]any{
				"messages": "Failed to register scheduler",
				"name":     name,
				"error":    err.Error(),
			})
		}
	}
	return nil
}

func (s *Scheduler) AddTask(spec string, task func()) (cron.EntryID, error) {
	return s.cron.AddFunc(spec, task)
}

func (s *Scheduler) Start() {
	if !s.started {
		s.cron.Start()
		s.started = true
	}
}

func (s *Scheduler) Stop() context.Context {
	if s.started {
		return s.cron.Stop()
	}
	return context.Background()
}

// Shutdown gracefully shuts down the scheduler
func (s *Scheduler) Shutdown(ctx context.Context) error {
	if !s.started {
		return nil
	}

	// Stop the scheduler - returns a context that's done when all running tasks finish
	stopCtx := s.Stop()

	select {
	case <-stopCtx.Done():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

var SchedulerRegistry = map[string]Handler{}

func RegisterScheduler(name string, scheduler Handler) {
	SchedulerRegistry[name] = scheduler
}
