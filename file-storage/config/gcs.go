package config

import (
	"fmt"
	"os"
	"time"
)

// GCSConfig holds Google Cloud Storage configuration
type GCSConfig struct {
	Project             string        // GCP project ID
	Bucket              string        // GCS bucket name
	BaseURL             string        // Base URL for downloads
	CredentialsFile     string        // Path to service account JSON
	MaxSize             int64         // Maximum file size in bytes
	AllowedTypes        []string      // Allowed MIME types
	PathPrefix          string        // Path prefix for GCS objects (e.g., "uploads", "files/documents")
	SignedURLExpiration time.Duration // Expiration duration for signed URLs (default: 15 minutes)
}

// Driver returns the driver name
func (gc *GCSConfig) Driver() string {
	return "gcs"
}

// Validate validates GCS configuration
func (gc *GCSConfig) Validate() error {
	if gc.Project == "" {
		return fmt.Errorf("gcs driver: project is required")
	}
	if gc.Bucket == "" {
		return fmt.Errorf("gcs driver: bucket is required")
	}
	if gc.BaseURL == "" {
		return fmt.Errorf("gcs driver: base_url is required")
	}
	if gc.CredentialsFile == "" {
		return fmt.Errorf("gcs driver: credentials_file is required")
	}
	if _, err := os.Stat(gc.CredentialsFile); err != nil {
		return fmt.Errorf("gcs driver: credentials_file not found: %w", err)
	}
	if gc.MaxSize <= 0 {
		return fmt.Errorf("gcs driver: max_size must be greater than 0")
	}
	if len(gc.AllowedTypes) == 0 {
		return fmt.Errorf("gcs driver: allowed_types cannot be empty")
	}
	return nil
}

// WithGCSDriver adds a GCS driver configuration
func (c *Config) WithGCSDriver(project, bucket, baseURL, credentialsFile string, maxSize int64, allowedTypes []string) *Config {
	if maxSize == 0 {
		maxSize = DefaultMaxSize()
	}
	if len(allowedTypes) == 0 {
		allowedTypes = DefaultAllowedTypes()
	}
	c.Drivers["gcs"] = &GCSConfig{
		Project:             project,
		Bucket:              bucket,
		BaseURL:             baseURL,
		CredentialsFile:     credentialsFile,
		MaxSize:             maxSize,
		AllowedTypes:        allowedTypes,
		PathPrefix:          "uploads",        // Default prefix
		SignedURLExpiration: 15 * time.Minute, // Default expiration
	}
	return c
}

// WithGCSDriverFull adds a GCS driver configuration with all options
func (c *Config) WithGCSDriverFull(project, bucket, baseURL, credentialsFile, pathPrefix string, maxSize int64, allowedTypes []string, signedURLExpiration time.Duration) *Config {
	if maxSize == 0 {
		maxSize = DefaultMaxSize()
	}
	if len(allowedTypes) == 0 {
		allowedTypes = DefaultAllowedTypes()
	}
	if pathPrefix == "" {
		pathPrefix = "uploads"
	}
	if signedURLExpiration == 0 {
		signedURLExpiration = 15 * time.Minute
	}
	c.Drivers["gcs"] = &GCSConfig{
		Project:             project,
		Bucket:              bucket,
		BaseURL:             baseURL,
		CredentialsFile:     credentialsFile,
		MaxSize:             maxSize,
		AllowedTypes:        allowedTypes,
		PathPrefix:          pathPrefix,
		SignedURLExpiration: signedURLExpiration,
	}
	return c
}
