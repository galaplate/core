package config

import (
	"fmt"
	"strconv"
	"time"
)

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
		bucket := getEnvValue("filesystems.disks.s3.bucket", "AWS_BUCKET")
		baseURL := getEnvValue("filesystems.disks.s3.url", "AWS_BASE_URL")
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
			ACL:          getEnvValue("filesystems.disks.s3.acl", "AWS_ACL"),
			StorageClass: getEnvValue("filesystems.disks.s3.storage_class", "AWS_STORAGE_CLASS"),
			PathPrefix:   getEnvValue("filesystems.disks.s3.path_prefix", "AWS_PATH_PREFIX"),
		}

		if s3Config.ACL == "" {
			s3Config.ACL = "private"
		}
		if s3Config.StorageClass == "" {
			s3Config.StorageClass = "STANDARD"
		}
		if s3Config.PathPrefix == "" {
			s3Config.PathPrefix = "uploads"
		}

		// Parse presigned URL expiration (in minutes)
		if expirationStr := getEnvValue("filesystems.disks.s3.presigned_expiration", "AWS_PRESIGNED_EXPIRATION"); expirationStr != "" {
			if minutes, err := strconv.Atoi(expirationStr); err == nil {
				s3Config.PresignedURLExpiration = time.Duration(minutes) * time.Minute
			}
		}
		if s3Config.PresignedURLExpiration == 0 {
			s3Config.PresignedURLExpiration = 15 * time.Minute
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

	if s3Bucket := getEnvValue("filesystems.disks.s3.bucket", "AWS_BUCKET"); s3Bucket != "" && driver != "s3" {
		region := getEnvValue("filesystems.disks.s3.region", "AWS_REGION")
		baseURL := getEnvValue("filesystems.disks.s3.url", "AWS_BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", s3Bucket, region)
		}
		pathPrefix := getEnvValue("filesystems.disks.s3.path_prefix", "AWS_PATH_PREFIX")
		acl := getEnvValue("filesystems.disks.s3.acl", "AWS_ACL")
		storageClass := getEnvValue("filesystems.disks.s3.storage_class", "AWS_STORAGE_CLASS")

		// Parse presigned URL expiration (in minutes)
		var presignedExpiration time.Duration
		if expirationStr := getEnvValue("filesystems.disks.s3.presigned_expiration", "AWS_PRESIGNED_EXPIRATION"); expirationStr != "" {
			if minutes, err := strconv.Atoi(expirationStr); err == nil {
				presignedExpiration = time.Duration(minutes) * time.Minute
			}
		}

		cfg.WithS3DriverFull(region, s3Bucket, baseURL, pathPrefix, 0, nil, acl, storageClass, presignedExpiration)
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
