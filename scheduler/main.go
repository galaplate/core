package scheduler

import (
	"context"
	"fmt"

	"github.com/robfig/cron/v3"
	"github.com/galaplate/core/logger"
)

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
		_, err := s.AddTask(task())
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

var SchedulerRegistry = map[string]func() (spec string, task func()){}

func RegisterScheduler(name string, scheduler func() (spec string, task func())) {
	SchedulerRegistry[name] = scheduler
}
