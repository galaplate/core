package models

import (
	"encoding/json"
	"time"
)

type JobState string

const (
	JobPending  JobState = "pending"
	JobStarted  JobState = "started"
	JobFinished JobState = "finished"
	JobFailed   JobState = "failed"
)

type Job struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	Type        string          `gorm:"not null" json:"type"`
	Payload     json.RawMessage `gorm:"type:text" json:"payload"`
	State       JobState        `gorm:"type:varchar(16);not null" json:"state"`
	ErrorMsg    string          `json:"error_msg"`
	Attempts    int             `json:"attempts"`
	AvailableAt time.Time       `json:"available_at"`
	CreatedAt   time.Time       `json:"created_at"`
	StartedAt   *time.Time      `json:"started_at"`
	FinishedAt  *time.Time      `json:"finished_at"`
}
