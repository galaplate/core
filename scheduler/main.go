package scheduler

import (
	"context"
	"fmt"

	"github.com/galaplate/core/logger"
	"github.com/robfig/cron/v3"
)


type Handler interface {
    Handle() (string, func())
}

type Scheduler struct {
	cron *cron.Cron
}

func New() *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithSeconds()),
	}
}

func (s *Scheduler) RunTasks() error {
	for name, task := range SchedulerRegistry {
		_, err := s.AddTask(task.Handle())
		if err != nil {
			logger.Fatal("Failed to register scheduler:", name, err)
		}
		fmt.Printf("Registering scheduler: %s\n", name)
	}
	return nil
}

func (s *Scheduler) AddTask(spec string, task func()) (cron.EntryID, error) {
	return s.cron.AddFunc(spec, task)
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() context.Context {
	return s.cron.Stop()
}

var SchedulerRegistry = map[string]Handler{}

func RegisterScheduler(name string, scheduler Handler) {
	SchedulerRegistry[name] = scheduler
}
