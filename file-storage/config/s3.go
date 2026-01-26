package config

import (
	"fmt"
	"time"
)

// S3Config holds AWS S3 configuration
type S3Config struct {
	Region                 string        // AWS region (e.g., "ap-southeast-1")
	Bucket                 string        // S3 bucket name
	BaseURL                string        // Base URL for downloads (e.g., "https://bucket.s3.amazonaws.com")
	AccessKey              string        // AWS access key ID
	SecretKey              string        // AWS secret access key
	MaxSize                int64         // Maximum file size in bytes
	AllowedTypes           []string      // Allowed MIME types
	ACL                    string        // S3 ACL (e.g., "private", "public-read")
	StorageClass           string        // S3 storage class (e.g., "STANDARD", "INTELLIGENT_TIERING")
	PathPrefix             string        // Path prefix for S3 keys (e.g., "uploads", "files/documents")
	PresignedURLExpiration time.Duration // Expiration duration for presigned URLs (default: 15 minutes)
}

// Driver returns the driver name
func (sc *S3Config) Driver() string {
	return "s3"
}

// Validate validates S3 configuration
func (sc *S3Config) Validate() error {
	if sc.Region == "" {
		return fmt.Errorf("s3 driver: region is required")
	}
	if sc.Bucket == "" {
		return fmt.Errorf("s3 driver: bucket is required")
	}
	if sc.BaseURL == "" {
		return fmt.Errorf("s3 driver: base_url is required")
	}
	if sc.AccessKey == "" {
		return fmt.Errorf("s3 driver: access_key is required")
	}
	if sc.SecretKey == "" {
		return fmt.Errorf("s3 driver: secret_key is required")
	}
	if sc.MaxSize <= 0 {
		return fmt.Errorf("s3 driver: max_size must be greater than 0")
	}
	if len(sc.AllowedTypes) == 0 {
		return fmt.Errorf("s3 driver: allowed_types cannot be empty")
	}
	return nil
}

// WithS3Driver adds an S3 driver configuration
func (c *Config) WithS3Driver(region, bucket, baseURL string, maxSize int64, allowedTypes []string) *Config {
	if maxSize == 0 {
		maxSize = DefaultMaxSize()
	}
	if len(allowedTypes) == 0 {
		allowedTypes = DefaultAllowedTypes()
	}
	c.Drivers["s3"] = &S3Config{
		Region:                 region,
		Bucket:                 bucket,
		BaseURL:                baseURL,
		MaxSize:                maxSize,
		AllowedTypes:           allowedTypes,
		ACL:                    "private",
		StorageClass:           "STANDARD",
		PathPrefix:             "uploads",        // Default prefix
		PresignedURLExpiration: 15 * time.Minute, // Default expiration
	}
	return c
}

// WithS3DriverFull adds an S3 driver configuration with all options
func (c *Config) WithS3DriverFull(region, bucket, baseURL, pathPrefix string, maxSize int64, allowedTypes []string, acl, storageClass string, presignedURLExpiration time.Duration) *Config {
	if maxSize == 0 {
		maxSize = DefaultMaxSize()
	}
	if len(allowedTypes) == 0 {
		allowedTypes = DefaultAllowedTypes()
	}
	if acl == "" {
		acl = "private"
	}
	if storageClass == "" {
		storageClass = "STANDARD"
	}
	if pathPrefix == "" {
		pathPrefix = "uploads"
	}
	if presignedURLExpiration == 0 {
		presignedURLExpiration = 15 * time.Minute
	}
	c.Drivers["s3"] = &S3Config{
		Region:                 region,
		Bucket:                 bucket,
		BaseURL:                baseURL,
		MaxSize:                maxSize,
		AllowedTypes:           allowedTypes,
		ACL:                    acl,
		StorageClass:           storageClass,
		PathPrefix:             pathPrefix,
		PresignedURLExpiration: presignedURLExpiration,
	}
	return c
}
