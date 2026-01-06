package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/galaplate/core/database"
	"github.com/galaplate/core/models"
)

type Task func()

type Queue struct {
	tasks           chan Task
	wg              sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
	started         bool
	mu              sync.Mutex
	ShutdownTimeout time.Duration
}

func New(bufferSize int) *Queue {
	ctx, cancel := context.WithCancel(context.Background())
	return &Queue{
		tasks:           make(chan Task, bufferSize),
		ctx:             ctx,
		cancel:          cancel,
		started:         false,
		ShutdownTimeout: 30 * time.Second,
	}
}

func (q *Queue) Start(workerCount int) {
	q.mu.Lock()
	if q.started {
		q.mu.Unlock()
		return
	}
	q.started = true
	q.mu.Unlock()

	if database.Connect.Migrator().HasTable(&models.Job{}) {
		for range workerCount {
			q.wg.Add(1)
			go q.worker()
		}
	}
}

func (q *Queue) worker() {
	defer q.wg.Done()

	for {
		select {
		case <-q.ctx.Done():
			return
		default:
		}

		var jobRecord models.Job
		start := time.Now()

		result := database.Connect.
			Where("state = ? AND available_at <= ?", models.JobPending, time.Now()).
			Order("created_at ASC").
			First(&jobRecord)

		if result.Error != nil {
			select {
			case <-q.ctx.Done():
				return
			case <-time.After(1 * time.Second):
				continue
			}
		}

		// Increment attempts when starting the job
		jobRecord.Attempts++
		updateResult := database.Connect.Model(&jobRecord).
			Where("id = ? AND state = ?", jobRecord.ID, models.JobPending).
			Updates(models.Job{
				State:     models.JobStarted,
				StartedAt: &start,
				Attempts:  jobRecord.Attempts,
			})

		if updateResult.RowsAffected == 0 {
			continue
		}

		job, err := ResolveJob(jobRecord.Type, jobRecord.Payload)
		if err != nil {
			failJob(&jobRecord, err)
			continue
		}

		err = job.Handle(jobRecord.Payload)

		if err != nil {
			if jobRecord.Attempts >= job.MaxAttempts() {
				failJob(&jobRecord, err)
			} else {
				database.Connect.Model(&jobRecord).Updates(models.Job{
					State:       models.JobPending,
					ErrorMsg:    err.Error(),
					AvailableAt: time.Now().Add(job.RetryAfter()),
				})
			}
		} else {
			database.Connect.Model(&jobRecord).Updates(models.Job{
				State:      models.JobFinished,
				FinishedAt: ptr(time.Now()),
			})
		}
	}
}

// Shutdown gracefully shuts down the queue, waiting for all workers to finish
func (q *Queue) Shutdown(ctx context.Context) error {
	q.mu.Lock()
	if !q.started {
		q.mu.Unlock()
		return nil
	}
	q.mu.Unlock()

	// Signal all workers to stop
	q.cancel()

	// Wait for all workers to finish with timeout
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func failJob(job *models.Job, err error) {
	database.Connect.Model(job).Updates(models.Job{
		State:      models.JobFailed,
		ErrorMsg:   err.Error(),
		FinishedAt: ptr(time.Now()),
	})
}

func ptr[T any](v T) *T {
	return &v
}

var registry = map[string]func() Job{}

func RegisterJob(job Job) {
	registry[job.Type()] = func() Job {
		return job
	}
}

func ResolveJob(typeName string, payload json.RawMessage) (Job, error) {
	creator, exists := registry[typeName]
	if !exists {
		return nil, fmt.Errorf("job type '%s' not registered", typeName)
	}

	job := creator()

	// // optional: decode payload into the job struct if needed
	// if err := json.Unmarshal(payload, &job); err != nil {
	// 	return nil, fmt.Errorf("failed to unmarshal payload for '%s': %w", typeName, err)
	// }

	return job, nil
}

type JobEnqueueRequest struct {
	Type    string
	Payload any
}

type Job interface {
	Type() string
	Handle(payload json.RawMessage) error
	MaxAttempts() int
	RetryAfter() time.Duration
}

func Dispatch(job Job, params ...any) error {
	_, err := SaveJobToDB(JobEnqueueRequest{
		Type:    job.Type(),
		Payload: params,
	})
	if err != nil {
		return fmt.Errorf("failed to save job to DB: %w", err)
	}

	return nil
}

func SaveJobToDB(req JobEnqueueRequest) (*models.Job, error) {
	payloadJSON, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	job := models.Job{
		Type:        req.Type,
		Payload:     payloadJSON,
		State:       models.JobPending,
		Attempts:    0,
		AvailableAt: now,
		CreatedAt:   now,
	}

	if err := database.Connect.Create(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}
