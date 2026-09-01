package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/galaplate/core/database"
	"github.com/galaplate/core/models"
)

var (
	ErrQueueAlreadyStarted = errors.New("queue already started")
	ErrQueueNotStarted     = errors.New("queue not started")
)

// Task is kept for backward compatibility but unused by workers.
type Task func()

// Job is the interface all queue jobs must implement.
type Job interface {
	Type() string
	Handle(ctx context.Context, payload json.RawMessage) error
	MaxAttempts() int
	RetryAfter() time.Duration
	Timeout() time.Duration
}

// JobEnqueueRequest is used to persist a job to the database.
type JobEnqueueRequest struct {
	Type        string
	Payload     any
	AvailableAt time.Time
}

// Hooks allow observability into the queue lifecycle.
type Hooks struct {
	OnJobStart   func(job models.Job)
	OnJobSuccess func(job models.Job, duration time.Duration)
	OnJobFailure func(job models.Job, duration time.Duration, err error)
}

type Queue struct {
	db     *gorm.DB
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	started bool
	mu      sync.Mutex

	ShutdownTimeout time.Duration
	Hooks           Hooks
}

// New creates a Queue backed by the given DB.
// If db is nil it falls back to database.Connect.
func New(db *gorm.DB) *Queue {
	if db == nil {
		db = database.Connect
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Queue{
		db:              db,
		ctx:             ctx,
		cancel:          cancel,
		ShutdownTimeout: 30 * time.Second,
	}
}

// Start begins the worker pool. It returns an error if the queue is
// already started or if the jobs table does not exist.
func (q *Queue) Start(workerCount int) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.started {
		return ErrQueueAlreadyStarted
	}

	if q.db == nil {
		return errors.New("database connection not available")
	}

	if !q.db.Migrator().HasTable(&models.Job{}) {
		return errors.New("jobs table does not exist")
	}

	q.started = true
	for range workerCount {
		q.wg.Add(1)
		go q.worker()
	}
	return nil
}

func (q *Queue) worker() {
	defer q.wg.Done()

	for {
		select {
		case <-q.ctx.Done():
			return
		default:
		}

		jobRecord, ok := q.claimJob()
		if !ok {
			select {
			case <-q.ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		q.processJob(jobRecord)
	}
}

// claimJob atomically picks the oldest pending job using SELECT FOR UPDATE.
func (q *Queue) claimJob() (models.Job, bool) {
	var jobRecord models.Job

	err := q.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("state = ? AND available_at <= ?", models.JobPending, time.Now()).
			Order("created_at ASC").
			First(&jobRecord)

		if result.Error != nil {
			return result.Error
		}

		now := time.Now()
		jobRecord.Attempts++
		jobRecord.StartedAt = &now
		return tx.Model(&jobRecord).Updates(models.Job{
			State:     models.JobStarted,
			Attempts:  jobRecord.Attempts,
			StartedAt: &now,
		}).Error
	})

	if err != nil {
		return models.Job{}, false
	}

	return jobRecord, true
}

func (q *Queue) processJob(jobRecord models.Job) {
	start := time.Now()

	if q.Hooks.OnJobStart != nil {
		q.Hooks.OnJobStart(jobRecord)
	}

	job, err := ResolveJob(jobRecord.Type, jobRecord.Payload)
	if err != nil {
		q.failJob(jobRecord, err, start)
		return
	}

	timeout := job.Timeout()
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}

	jobCtx, cancel := context.WithTimeout(q.ctx, timeout)
	defer cancel()

	// Recover from panics so a single bad job cannot kill a worker.
	err = func() (runErr error) {
		defer func() {
			if r := recover(); r != nil {
				runErr = fmt.Errorf("panic: %v", r)
			}
		}()
		return job.Handle(jobCtx, jobRecord.Payload)
	}()

	duration := time.Since(start)

	if err != nil {
		if q.Hooks.OnJobFailure != nil {
			q.Hooks.OnJobFailure(jobRecord, duration, err)
		}

		if jobRecord.Attempts >= job.MaxAttempts() {
			q.failJob(jobRecord, err, start)
		} else {
			q.db.Model(&jobRecord).Updates(models.Job{
				State:       models.JobPending,
				ErrorMsg:    err.Error(),
				AvailableAt: time.Now().Add(job.RetryAfter()),
			})
		}
	} else {
		finishedAt := time.Now()
		q.db.Model(&jobRecord).Updates(models.Job{
			State:      models.JobFinished,
			FinishedAt: &finishedAt,
		})

		if q.Hooks.OnJobSuccess != nil {
			q.Hooks.OnJobSuccess(jobRecord, duration)
		}
	}
}

func (q *Queue) failJob(jobRecord models.Job, err error, start time.Time) {
	finishedAt := time.Now()
	q.db.Model(&jobRecord).Updates(models.Job{
		State:      models.JobFailed,
		ErrorMsg:   err.Error(),
		FinishedAt: &finishedAt,
	})

	if q.Hooks.OnJobFailure != nil {
		q.Hooks.OnJobFailure(jobRecord, time.Since(start), err)
	}
}

// Shutdown gracefully stops all workers.
func (q *Queue) Shutdown(ctx context.Context) error {
	q.mu.Lock()
	if !q.started {
		q.mu.Unlock()
		return nil
	}
	q.cancel()
	q.started = false
	q.mu.Unlock()

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

// Dispatch enqueues a job to be processed immediately.
func Dispatch(job Job, payload any) error {
	_, err := SaveJobToDB(JobEnqueueRequest{
		Type:    job.Type(),
		Payload: payload,
	})
	if err != nil {
		return fmt.Errorf("failed to save job to DB: %w", err)
	}
	return nil
}

// DispatchDelayed enqueues a job to be processed after a delay.
func DispatchDelayed(job Job, payload any, delay time.Duration) error {
	_, err := SaveJobToDB(JobEnqueueRequest{
		Type:        job.Type(),
		Payload:     payload,
		AvailableAt: time.Now().Add(delay),
	})
	if err != nil {
		return fmt.Errorf("failed to save job to DB: %w", err)
	}
	return nil
}

// SaveJobToDB persists a JobEnqueueRequest to the database.
func SaveJobToDB(req JobEnqueueRequest) (*models.Job, error) {
	payloadJSON, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, err
	}

	availableAt := req.AvailableAt
	if availableAt.IsZero() {
		availableAt = time.Now()
	}

	job := models.Job{
		Type:        req.Type,
		Payload:     payloadJSON,
		State:       models.JobPending,
		Attempts:    0,
		AvailableAt: availableAt,
		CreatedAt:   time.Now(),
	}

	if err := database.Connect.Create(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

// RetryJob resets a failed job back to pending so it can be retried.
func RetryJob(jobID uint) error {
	result := database.Connect.Model(&models.Job{}).
		Where("id = ? AND state = ?", jobID, models.JobFailed).
		Updates(models.Job{
			State:       models.JobPending,
			ErrorMsg:    "",
			AvailableAt: time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("job %d not found or not in failed state", jobID)
	}
	return nil
}

// PruneJobs deletes finished/failed jobs older than the given age.
func PruneJobs(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result := database.Connect.Where(
		"state IN ? AND finished_at < ?",
		[]models.JobState{models.JobFinished, models.JobFailed},
		cutoff,
	).Delete(&models.Job{})

	return result.RowsAffected, result.Error
}

// Stats returns counts of jobs by state.
func Stats() (map[models.JobState]int64, error) {
	stats := make(map[models.JobState]int64)

	var rows []struct {
		State models.JobState
		Count int64
	}
	err := database.Connect.Model(&models.Job{}).
		Select("state, COUNT(*) as count").
		Group("state").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, r := range rows {
		stats[r.State] = r.Count
	}
	return stats, nil
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
	return job, nil
}

func ptr[T any](v T) *T {
	return &v
}
