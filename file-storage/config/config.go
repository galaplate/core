package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all filesystem configuration
type Config struct {
	Default string                  // Default driver ("local", "s3", "gcs")
	Drivers map[string]DriverConfig // Driver-specific configurations
}

// DriverConfig is the base configuration for any driver
type DriverConfig interface {
	Driver() string
	Validate() error
}

// LocalConfig holds local filesystem configuration
type LocalConfig struct {
	Path         string   // Directory path for uploads
	MaxSize      int64    // Maximum file size in bytes
	AllowedTypes []string // Allowed MIME types
}

// Driver returns the driver name
func (lc *LocalConfig) Driver() string {
	return "local"
}

// Validate validates local configuration
func (lc *LocalConfig) Validate() error {
	if lc.Path == "" {
		return fmt.Errorf("local driver: path is required")
	}
	if lc.MaxSize <= 0 {
		return fmt.Errorf("local driver: max_size must be greater than 0")
	}
	if len(lc.AllowedTypes) == 0 {
		return fmt.Errorf("local driver: allowed_types cannot be empty")
	}
	return nil
}

// S3Config holds AWS S3 configuration
type S3Config struct {
	Region       string   // AWS region (e.g., "ap-southeast-1")
	Bucket       string   // S3 bucket name
	BaseURL      string   // Base URL for downloads (e.g., "https://bucket.s3.amazonaws.com")
	AccessKey    string   // AWS access key ID
	SecretKey    string   // AWS secret access key
	MaxSize      int64    // Maximum file size in bytes
	AllowedTypes []string // Allowed MIME types
	ACL          string   // S3 ACL (e.g., "private", "public-read")
	StorageClass string   // S3 storage class (e.g., "STANDARD", "INTELLIGENT_TIERING")
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

// GCSConfig holds Google Cloud Storage configuration
type GCSConfig struct {
	Project         string   // GCP project ID
	Bucket          string   // GCS bucket name
	BaseURL         string   // Base URL for downloads
	CredentialsFile string   // Path to service account JSON
	MaxSize         int64    // Maximum file size in bytes
	AllowedTypes    []string // Allowed MIME types
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

// DefaultAllowedTypes returns the default list of allowed MIME types
func DefaultAllowedTypes() []string {
	return []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}
}

// DefaultMaxSize returns the default maximum file size (10MB)
func DefaultMaxSize() int64 {
	return 10 * 1024 * 1024
}

// NewConfig creates a new Config instance
func New() *Config {
	return &Config{
		Drivers: make(map[string]DriverConfig),
	}
}

// WithLocalDriver adds a local driver configuration
func (c *Config) WithLocalDriver(path string, maxSize int64, allowedTypes []string) *Config {
	if maxSize == 0 {
		maxSize = DefaultMaxSize()
	}
	if len(allowedTypes) == 0 {
		allowedTypes = DefaultAllowedTypes()
	}
	c.Drivers["local"] = &LocalConfig{
		Path:         path,
		MaxSize:      maxSize,
		AllowedTypes: allowedTypes,
	}
	return c
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
		Region:       region,
		Bucket:       bucket,
		BaseURL:      baseURL,
		MaxSize:      maxSize,
		AllowedTypes: allowedTypes,
		ACL:          "private",
		StorageClass: "STANDARD",
	}
	return c
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
		Project:         project,
		Bucket:          bucket,
		BaseURL:         baseURL,
		CredentialsFile: credentialsFile,
		MaxSize:         maxSize,
		AllowedTypes:    allowedTypes,
	}
	return c
}

// SetDefault sets the default driver
func (c *Config) SetDefault(driver string) *Config {
	c.Default = driver
	return c
}

// GetDriver returns driver configuration by name
func (c *Config) GetDriver(name string) (DriverConfig, error) {
	driver, exists := c.Drivers[name]
	if !exists {
		return nil, fmt.Errorf("driver not found: %s", name)
	}
	return driver, nil
}

// GetDefaultDriver returns the default driver configuration
func (c *Config) GetDefaultDriver() (DriverConfig, error) {
	if c.Default == "" {
		return nil, fmt.Errorf("no default driver configured")
	}
	return c.GetDriver(c.Default)
}

// Validate validates all driver configurations
func (c *Config) Validate() error {
	if c.Default == "" {
		return fmt.Errorf("no default driver configured")
	}

	if len(c.Drivers) == 0 {
		return fmt.Errorf("no drivers configured")
	}

	for name, driver := range c.Drivers {
		if err := driver.Validate(); err != nil {
			return fmt.Errorf("driver %s validation failed: %w", name, err)
		}
	}

	// Verify default driver exists
	if _, exists := c.Drivers[c.Default]; !exists {
		return fmt.Errorf("default driver %s not configured", c.Default)
	}

	return nil
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	cfg := New()

	driver := os.Getenv("FILESYSTEM_DRIVER")
	if driver == "" {
		driver = "local"
	}

	cfg.SetDefault(driver)

	// Load all driver configurations
	switch driver {
	case "local":
		localPath := os.Getenv("FILESYSTEM_LOCAL_PATH")
		if localPath == "" {
			localPath = "storage/app/uploads"
		}
		maxSizeStr := os.Getenv("FILESYSTEM_MAX_SIZE")
		var maxSize int64 = DefaultMaxSize()
		if maxSizeStr != "" {
			if size, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
				maxSize = size
			}
		}

		cfg.WithLocalDriver(localPath, maxSize, DefaultAllowedTypes())

	case "s3":
		region := os.Getenv("AWS_REGION")
		bucket := os.Getenv("S3_BUCKET")
		baseURL := os.Getenv("S3_BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucket, region)
		}

		maxSizeStr := os.Getenv("FILESYSTEM_MAX_SIZE")
		var maxSize int64 = DefaultMaxSize()
		if maxSizeStr != "" {
			if size, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
				maxSize = size
			}
		}

		s3Config := &S3Config{
			Region:       region,
			Bucket:       bucket,
			BaseURL:      baseURL,
			AccessKey:    os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretKey:    os.Getenv("AWS_SECRET_ACCESS_KEY"),
			MaxSize:      maxSize,
			AllowedTypes: DefaultAllowedTypes(),
			ACL:          os.Getenv("S3_ACL"),
			StorageClass: os.Getenv("S3_STORAGE_CLASS"),
		}

		if s3Config.ACL == "" {
			s3Config.ACL = "private"
		}
		if s3Config.StorageClass == "" {
			s3Config.StorageClass = "STANDARD"
		}

		cfg.Drivers["s3"] = s3Config

	case "gcs":
		project := os.Getenv("GCP_PROJECT")
		bucket := os.Getenv("GCS_BUCKET")
		baseURL := os.Getenv("GCS_BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("https://storage.googleapis.com/%s", bucket)
		}
		credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

		maxSizeStr := os.Getenv("FILESYSTEM_MAX_SIZE")
		var maxSize int64 = DefaultMaxSize()
		if maxSizeStr != "" {
			if size, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
				maxSize = size
			}
		}

		cfg.WithGCSDriver(project, bucket, baseURL, credentialsFile, maxSize, DefaultAllowedTypes())
	}

	// Also load other drivers if environment variables are set
	if os.Getenv("FILESYSTEM_LOCAL_PATH") != "" {
		localPath := os.Getenv("FILESYSTEM_LOCAL_PATH")
		cfg.WithLocalDriver(localPath, 0, nil)
	}

	if os.Getenv("S3_BUCKET") != "" && driver != "s3" {
		region := os.Getenv("AWS_REGION")
		bucket := os.Getenv("S3_BUCKET")
		baseURL := os.Getenv("S3_BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucket, region)
		}
		cfg.WithS3Driver(region, bucket, baseURL, 0, nil)
	}

	if os.Getenv("GCS_BUCKET") != "" && driver != "gcs" {
		project := os.Getenv("GCP_PROJECT")
		bucket := os.Getenv("GCS_BUCKET")
		baseURL := os.Getenv("GCS_BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("https://storage.googleapis.com/%s", bucket)
		}
		credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		cfg.WithGCSDriver(project, bucket, baseURL, credentialsFile, 0, nil)
	}

	return cfg, nil
}

// String returns a string representation of the configuration
func (c *Config) String() string {
	var sb strings.Builder
	sb.WriteString("Filesystems Configuration:\n")
	sb.WriteString(fmt.Sprintf("Default Driver: %s\n", c.Default))
	sb.WriteString("Configured Drivers:\n")

	for name, driver := range c.Drivers {
		sb.WriteString(fmt.Sprintf("  - %s (%s)\n", name, driver.Driver()))
	}

	return sb.String()
}
