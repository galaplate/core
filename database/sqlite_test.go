package database

import (
	"os"
	"testing"
	"time"

	"github.com/galaplate/core/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSQLiteConnection(t *testing.T) {
	// Create temporary SQLite database
	tempDB := "test.sqlite"
	defer os.Remove(tempDB)

	// Test SQLite connection
	db, err := gorm.Open(sqlite.Open(tempDB), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}

	// Test auto-migration with Job model
	err = db.AutoMigrate(&models.Job{})
	if err != nil {
		t.Fatalf("Failed to migrate Job model: %v", err)
	}

	// Test creating a job record
	job := models.Job{
		Type:        "test_job",
		Payload:     []byte(`{"test": "data"}`),
		State:       models.JobPending,
		Attempts:    0,
		AvailableAt: time.Now(),
		CreatedAt:   time.Now(),
	}

	err = db.Create(&job).Error
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Test retrieving the job
	var retrievedJob models.Job
	err = db.First(&retrievedJob, job.ID).Error
	if err != nil {
		t.Fatalf("Failed to retrieve job: %v", err)
	}

	if retrievedJob.Type != "test_job" {
		t.Errorf("Expected job type 'test_job', got '%s'", retrievedJob.Type)
	}

	if retrievedJob.State != models.JobPending {
		t.Errorf("Expected job state 'pending', got '%s'", retrievedJob.State)
	}

	t.Log("✅ SQLite retry provider test passed")
}
