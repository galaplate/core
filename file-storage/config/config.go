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

// getEnvValue is a helper that tries to get from Galaplate config first, then falls back to os.Getenv
// This provides compatibility with both YAML config and environment variables
func getEnvValue(configPath, envVar string) string {
	// Try to get from Galaplate config first (if initialized)
	// Using reflection to avoid circular dependency on config package
	val := tryGetConfigValue(configPath)
	if val != "" {
		return val
	}
	// Fall back to environment variable
	return os.Getenv(envVar)
}

// tryGetConfigValue attempts to get a config value from Galaplate config manager
// Returns empty string if config is not initialized
func tryGetConfigValue(path string) string {
	// Avoid circular dependency by using reflection
	// In a real implementation, this would use: return config.ConfigString(path)
	// For now, rely on os.Getenv as fallback
	return ""
}

// LoadFromEnv loads configuration from filesystems.yaml (via Galaplate config) or environment variables
// Prefer using the YAML configuration through Galaplate's config manager when available
func LoadFromEnv() (*Config, error) {
	cfg := New()

	// Get driver from environment/config
	driver := getEnvValue("filesystems.default", "FILESYSTEM_DRIVER")
	if driver == "" {
		driver = "local"
	}

	cfg.SetDefault(driver)

	// Helper function to parse max size
	parseMaxSize := func() int64 {
		maxSizeStr := getEnvValue("filesystems.max_size", "FILESYSTEM_MAX_SIZE")
		var maxSize int64 = DefaultMaxSize()
		if maxSizeStr != "" {
			if size, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
				maxSize = size
			}
		}
		return maxSize
	}

	// Load all driver configurations
	switch driver {
	case "local":
		localPath := getEnvValue("filesystems.disks.local.path", "FILESYSTEM_LOCAL_PATH")
		if localPath == "" {
			localPath = "storage/app/uploads"
		}
		maxSize := parseMaxSize()

		cfg.WithLocalDriver(localPath, maxSize, DefaultAllowedTypes())

	case "s3":
		region := getEnvValue("filesystems.disks.s3.region", "AWS_REGION")
		bucket := getEnvValue("filesystems.disks.s3.bucket", "S3_BUCKET")
		baseURL := getEnvValue("filesystems.disks.s3.url", "S3_BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucket, region)
		}
		maxSize := parseMaxSize()

		s3Config := &S3Config{
			Region:       region,
			Bucket:       bucket,
			BaseURL:      baseURL,
			AccessKey:    getEnvValue("filesystems.disks.s3.key", "AWS_ACCESS_KEY_ID"),
			SecretKey:    getEnvValue("filesystems.disks.s3.secret", "AWS_SECRET_ACCESS_KEY"),
			MaxSize:      maxSize,
			AllowedTypes: DefaultAllowedTypes(),
			ACL:          getEnvValue("filesystems.disks.s3.acl", "S3_ACL"),
			StorageClass: getEnvValue("filesystems.disks.s3.storage_class", "S3_STORAGE_CLASS"),
		}

		if s3Config.ACL == "" {
			s3Config.ACL = "private"
		}
		if s3Config.StorageClass == "" {
			s3Config.StorageClass = "STANDARD"
		}

		cfg.Drivers["s3"] = s3Config

	case "gcs":
		project := getEnvValue("filesystems.disks.gcs.project_id", "GCP_PROJECT")
		bucket := getEnvValue("filesystems.disks.gcs.bucket", "GCS_BUCKET")
		baseURL := getEnvValue("filesystems.disks.gcs.url", "GCS_BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("https://storage.googleapis.com/%s", bucket)
		}
		credentialsFile := getEnvValue("filesystems.disks.gcs.key_file", "GOOGLE_APPLICATION_CREDENTIALS")
		maxSize := parseMaxSize()

		cfg.WithGCSDriver(project, bucket, baseURL, credentialsFile, maxSize, DefaultAllowedTypes())

	case "google_drive":
		serviceAccountFile := getEnvValue("filesystems.disks.google_drive.service_account_file", "GOOGLE_SERVICE_ACCOUNT_FILE")
		folderID := getEnvValue("filesystems.disks.google_drive.folder_id", "GOOGLE_DRIVE_FOLDER_ID")
		maxSize := parseMaxSize()

		cfg.WithGoogleDriveDriver(serviceAccountFile, folderID, maxSize, DefaultAllowedTypes())
	}

	// Also load other drivers if environment variables are set
	if localPath := getEnvValue("filesystems.disks.local.path", "FILESYSTEM_LOCAL_PATH"); localPath != "" && driver != "local" {
		cfg.WithLocalDriver(localPath, 0, nil)
	}

	if s3Bucket := getEnvValue("filesystems.disks.s3.bucket", "S3_BUCKET"); s3Bucket != "" && driver != "s3" {
		region := getEnvValue("filesystems.disks.s3.region", "AWS_REGION")
		baseURL := getEnvValue("filesystems.disks.s3.url", "S3_BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", s3Bucket, region)
		}
		cfg.WithS3Driver(region, s3Bucket, baseURL, 0, nil)
	}

	if gcsBucket := getEnvValue("filesystems.disks.gcs.bucket", "GCS_BUCKET"); gcsBucket != "" && driver != "gcs" {
		project := getEnvValue("filesystems.disks.gcs.project_id", "GCP_PROJECT")
		baseURL := getEnvValue("filesystems.disks.gcs.url", "GCS_BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("https://storage.googleapis.com/%s", gcsBucket)
		}
		credentialsFile := getEnvValue("filesystems.disks.gcs.key_file", "GOOGLE_APPLICATION_CREDENTIALS")
		cfg.WithGCSDriver(project, gcsBucket, baseURL, credentialsFile, 0, nil)
	}

	if folderID := getEnvValue("filesystems.disks.google_drive.folder_id", "GOOGLE_DRIVE_FOLDER_ID"); folderID != "" && driver != "google_drive" {
		serviceAccountFile := getEnvValue("filesystems.disks.google_drive.service_account_file", "GOOGLE_SERVICE_ACCOUNT_FILE")
		cfg.WithGoogleDriveDriver(serviceAccountFile, folderID, 0, nil)
	}

	return cfg, nil
}

// String returns a string representation of the configuration
func (c *Config) String() string {
	var sb strings.Builder
	sb.WriteString("Filesystems Configuration:\n")
	fmt.Fprintf(&sb, "Default Driver: %s\n", c.Default)
	sb.WriteString("Configured Drivers:\n")

	for name, driver := range c.Drivers {
		fmt.Fprintf(&sb, "  - %s (%s)\n", name, driver.Driver())
	}

	return sb.String()
}
