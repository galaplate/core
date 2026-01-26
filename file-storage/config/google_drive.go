package config

import (
	"fmt"
	"os"
)

// GoogleDriveConfig holds Google Drive configuration
type GoogleDriveConfig struct {
	ServiceAccountFile string   // Path to service account JSON credentials
	FolderID           string   // Google Drive folder ID where files will be uploaded
	MaxSize            int64    // Maximum file size in bytes
	AllowedTypes       []string // Allowed MIME types
}

// Driver returns the driver name
func (gd *GoogleDriveConfig) Driver() string {
	return "google_drive"
}

// Validate validates Google Drive configuration
func (gd *GoogleDriveConfig) Validate() error {
	if gd.ServiceAccountFile == "" {
		return fmt.Errorf("google_drive driver: service_account_file is required")
	}
	if _, err := os.Stat(gd.ServiceAccountFile); err != nil {
		return fmt.Errorf("google_drive driver: service_account_file not found: %w", err)
	}
	if gd.FolderID == "" {
		return fmt.Errorf("google_drive driver: folder_id is required")
	}
	if gd.MaxSize <= 0 {
		return fmt.Errorf("google_drive driver: max_size must be greater than 0")
	}
	if len(gd.AllowedTypes) == 0 {
		return fmt.Errorf("google_drive driver: allowed_types cannot be empty")
	}
	return nil
}

// WithGoogleDriveDriver adds a Google Drive driver configuration
func (c *Config) WithGoogleDriveDriver(serviceAccountFile, folderID string, maxSize int64, allowedTypes []string) *Config {
	if maxSize == 0 {
		maxSize = DefaultMaxSize()
	}
	if len(allowedTypes) == 0 {
		allowedTypes = DefaultAllowedTypes()
	}
	c.Drivers["google_drive"] = &GoogleDriveConfig{
		ServiceAccountFile: serviceAccountFile,
		FolderID:           folderID,
		MaxSize:            maxSize,
		AllowedTypes:       allowedTypes,
	}
	return c
}
