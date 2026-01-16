package config

// This file contains examples of how to use the filesystem configuration.
// It's not meant to be imported, but serves as documentation.

/*
============================================================
FILESYSTEM CONFIG - USAGE EXAMPLES
============================================================

1. BASIC LOCAL STORAGE SETUP
============================================================

In your main.go:

	import "bisma-api/modules/file-storage/config"

	func init() {
		cfg := config.New().
			WithLocalDriver("storage/app/uploads", 0, nil).
			SetDefault("local")

		if err := cfg.Validate(); err != nil {
			panic(err)
		}
	}


2. LOAD FROM ENVIRONMENT VARIABLES
============================================================

In your main.go:

	cfg, err := config.LoadFromEnv()
	if err != nil {
		panic(err)
	}

	if err := cfg.Validate(); err != nil {
		panic(err)
	}

Environment Variables:

	# Default driver
	FILESYSTEM_DRIVER=local

	# Local storage
	FILESYSTEM_LOCAL_PATH=storage/app/uploads
	FILESYSTEM_MAX_SIZE=10485760  # 10MB in bytes

	# AWS S3
	AWS_REGION=ap-southeast-1
	S3_BUCKET=my-bucket
	S3_BASE_URL=https://my-bucket.s3.amazonaws.com
	AWS_ACCESS_KEY_ID=your-key
	AWS_SECRET_ACCESS_KEY=your-secret
	S3_ACL=private
	S3_STORAGE_CLASS=INTELLIGENT_TIERING

	# Google Cloud Storage
	GCP_PROJECT=my-project
	GCS_BUCKET=my-bucket
	GCS_BASE_URL=https://storage.googleapis.com/my-bucket
	GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json


3. MULTI-DRIVER CONFIGURATION
============================================================

In your main.go:

	cfg := config.New().
		WithLocalDriver("storage/app/uploads", 0, nil).
		WithS3Driver("ap-southeast-1", "my-bucket", "https://my-bucket.s3.amazonaws.com", 0, nil).
		WithGCSDriver("my-project", "my-bucket", "https://storage.googleapis.com/my-bucket",
			"/path/to/service-account.json", 0, nil).
		SetDefault("local")

	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	// Get default driver
	defaultDriver, _ := cfg.GetDefaultDriver()

	// Get specific driver
	s3Driver, _ := cfg.GetDriver("s3")


4. CUSTOM ALLOWED TYPES
============================================================

	customTypes := []string{"image/jpeg", "application/pdf"}

	cfg := config.New().
		WithLocalDriver("storage/uploads", 0, customTypes).
		SetDefault("local")

	if err := cfg.Validate(); err != nil {
		panic(err)
	}


5. CUSTOM MAX FILE SIZE
============================================================

	// 50MB max size
	maxSize := int64(50 * 1024 * 1024)

	cfg := config.New().
		WithLocalDriver("storage/uploads", maxSize, nil).
		SetDefault("local")


6. BUILDER PATTERN
============================================================

	cfg := config.New().
		WithLocalDriver("storage/local", 0, nil).
		WithS3Driver("ap-southeast-1", "prod-bucket", "", 0, nil).
		WithGCSDriver("prod-project", "prod-bucket", "", "", 0, nil).
		SetDefault("s3")

	// Fluent API allows method chaining


7. ACCESS CONFIGURATION
============================================================

	// Get configuration from environment
	cfg, err := config.LoadFromEnv()
	if err != nil {
		panic(err)
	}

	// Get local driver config
	localDriver, err := cfg.GetDriver("local")
	if err != nil {
		panic(err)
	}

	// Cast to LocalConfig
	localCfg := localDriver.(*config.LocalConfig)
	println("Upload path:", localCfg.Path)
	println("Max size:", localCfg.MaxSize)


8. ERROR HANDLING
============================================================

	cfg := config.New().
		WithLocalDriver("storage/uploads", 0, nil)

	// Missing default driver
	if err := cfg.Validate(); err != nil {
		println("Config error:", err) // "no default driver configured"
	}

	cfg.SetDefault("nonexistent")
	if err := cfg.Validate(); err != nil {
		println("Config error:", err) // "default driver nonexistent not configured"
	}

	cfg = config.New().WithLocalDriver("", 0, nil).SetDefault("local")
	if err := cfg.Validate(); err != nil {
		println("Config error:", err) // "driver local validation failed: local driver: path is required"
	}


9. ENVIRONMENT-BASED DRIVER SELECTION
============================================================

	func loadFilesystemConfig(env string) *config.Config {
		cfg := config.New()

		switch env {
		case "production":
			cfg.WithS3Driver(
				os.Getenv("AWS_REGION"),
				os.Getenv("S3_BUCKET"),
				os.Getenv("S3_BASE_URL"),
				0, nil,
			).SetDefault("s3")

		case "staging":
			cfg.WithGCSDriver(
				os.Getenv("GCP_PROJECT"),
				os.Getenv("GCS_BUCKET"),
				os.Getenv("GCS_BASE_URL"),
				os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
				0, nil,
			).SetDefault("gcs")

		default: // development
			cfg.WithLocalDriver("storage/app/uploads", 0, nil).SetDefault("local")
		}

		return cfg
	}


10. VALIDATE AND LOG CONFIGURATION
============================================================

	cfg, err := config.LoadFromEnv()
	if err != nil {
		panic(err)
	}

	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	// Print configuration
	println(cfg.String())
	// Output:
	// Filesystems Configuration:
	// Default Driver: local
	// Configured Drivers:
	//   - local (local)


11. PROGRAMMATIC CONFIG WITH FALLBACKS
============================================================

	// Start with local storage
	cfg := config.New().
		WithLocalDriver("storage/app/uploads", 0, nil).
		SetDefault("local")

	// Try to add S3 if credentials available
	if os.Getenv("S3_BUCKET") != "" {
		cfg.WithS3Driver(
			os.Getenv("AWS_REGION"),
			os.Getenv("S3_BUCKET"),
			os.Getenv("S3_BASE_URL"),
			0, nil,
		)
	}

	// Try to add GCS if credentials available
	if os.Getenv("GCS_BUCKET") != "" {
		cfg.WithGCSDriver(
			os.Getenv("GCP_PROJECT"),
			os.Getenv("GCS_BUCKET"),
			os.Getenv("GCS_BASE_URL"),
			os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
			0, nil,
		)
	}

	if err := cfg.Validate(); err != nil {
		panic(err)
	}


12. CUSTOM DRIVER CONFIGURATION
============================================================

	// Create custom driver config
	customConfig := &config.LocalConfig{
		Path:        "/custom/path",
		MaxSize:     50 * 1024 * 1024, // 50MB
		AllowedTypes: []string{"image/jpeg", "application/pdf"},
	}

	cfg := config.New()
	cfg.Drivers["custom"] = customConfig
	cfg.SetDefault("custom")

	if err := cfg.Validate(); err != nil {
		panic(err)
	}


============================================================
ENVIRONMENT VARIABLES REFERENCE
============================================================

FILESYSTEM_DRIVER
  Description: Default filesystem driver to use
  Values: "local", "s3", "gcs"
  Default: "local"
  Example: FILESYSTEM_DRIVER=s3

FILESYSTEM_LOCAL_PATH
  Description: Path for local storage uploads
  Example: FILESYSTEM_LOCAL_PATH=storage/app/uploads

FILESYSTEM_MAX_SIZE
  Description: Maximum file size in bytes
  Default: 10485760 (10MB)
  Example: FILESYSTEM_MAX_SIZE=52428800 (50MB)

AWS_REGION
  Description: AWS region for S3
  Required for: S3 driver
  Example: AWS_REGION=ap-southeast-1

S3_BUCKET
  Description: S3 bucket name
  Required for: S3 driver
  Example: S3_BUCKET=my-bucket

S3_BASE_URL
  Description: Base URL for S3 file downloads
  Default: https://{bucket}.s3.{region}.amazonaws.com
  Example: S3_BASE_URL=https://cdn.example.com

AWS_ACCESS_KEY_ID
  Description: AWS access key for S3
  Required for: S3 driver
  Example: AWS_ACCESS_KEY_ID=AKIA...

AWS_SECRET_ACCESS_KEY
  Description: AWS secret key for S3
  Required for: S3 driver
  Example: AWS_SECRET_ACCESS_KEY=...

S3_ACL
  Description: S3 object ACL
  Default: "private"
  Values: "private", "public-read", etc.
  Example: S3_ACL=private

S3_STORAGE_CLASS
  Description: S3 storage class
  Default: "STANDARD"
  Values: "STANDARD", "INTELLIGENT_TIERING", etc.
  Example: S3_STORAGE_CLASS=INTELLIGENT_TIERING

GCP_PROJECT
  Description: GCP project ID
  Required for: GCS driver
  Example: GCP_PROJECT=my-project

GCS_BUCKET
  Description: GCS bucket name
  Required for: GCS driver
  Example: GCS_BUCKET=my-bucket

GCS_BASE_URL
  Description: Base URL for GCS file downloads
  Default: https://storage.googleapis.com/{bucket}
  Example: GCS_BASE_URL=https://storage.googleapis.com/my-bucket

GOOGLE_APPLICATION_CREDENTIALS
  Description: Path to GCS service account JSON
  Required for: GCS driver
  Example: GOOGLE_APPLICATION_CREDENTIALS=/etc/gcp/service-account.json


============================================================
CONFIGURATION VALIDATION
============================================================

The Config type validates:
  ✓ Default driver is set
  ✓ Default driver is configured
  ✓ At least one driver is configured
  ✓ Each driver configuration is valid

Validation errors:
  - "no default driver configured"
  - "default driver X not configured"
  - "no drivers configured"
  - "driver X validation failed: ..."
  - "driver X: field Y is required"

Example validation:
  if err := cfg.Validate(); err != nil {
      log.Fatalf("Configuration error: %v", err)
  }

*/
