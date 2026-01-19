package providers

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"slices"
	"time"

	"github.com/google/uuid"

	filestorage "github.com/galaplate/core/file-storage"
)

// GCSConfig holds Google Cloud Storage-specific configuration
type GCSConfig struct {
	BaseConfig filestorage.FileUploadConfig
	Client     interface{} // *storage.Client - use interface to avoid SDK dependency
	BucketName string
	ProjectID  string
	BaseURL    string
}

// GCSStorage implements FileStorageProvider for Google Cloud Storage
type GCSStorage struct {
	config GCSConfig
}

// NewGCSStorage creates a new Google Cloud Storage provider
// Note: Client should be *storage.Client from cloud.google.com/go/storage
// Using interface{} to avoid forcing the SDK dependency
func NewGCSStorage(cfg GCSConfig) *GCSStorage {
	return &GCSStorage{
		config: cfg,
	}
}

// Upload stores a file to GCS and returns metadata
// NOTE: This does NOT save to database - that's the caller's responsibility
func (gcs *GCSStorage) Upload(file *multipart.FileHeader) filestorage.UploadMetadata {
	// Validate file size
	if file.Size > gcs.config.BaseConfig.MaxSize {
		return filestorage.UploadMetadata{
			Error: "file_too_large",
		}
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	isAllowed := slices.Contains(gcs.config.BaseConfig.AllowedTypes, contentType)

	if !isAllowed {
		return filestorage.UploadMetadata{
			Error: "invalid_file_type",
		}
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	uniqueID := uuid.New().String()
	timestamp := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s%s", timestamp, uniqueID, ext)

	// Create GCS object path
	gcsPath := fmt.Sprintf("uploads/%s", fileName)

	// Open file for upload
	src, err := file.Open()
	if err != nil {
		return filestorage.UploadMetadata{
			Error: "file_open_failed",
		}
	}
	defer src.Close()

	// TODO: Implement GCS upload using provided gcsClient
	// Example implementation when you have GCS SDK:
	// ctx := context.Background()
	// wc := gcs.config.Client.Bucket(gcs.config.BucketName).Object(gcsPath).NewWriter(ctx)
	// if _, err := io.Copy(wc, src); err != nil { ... }

	// For now, return placeholder result
	filePath := fmt.Sprintf("%s/%s", gcs.config.BaseURL, gcsPath)

	// Return metadata only - caller is responsible for saving to database
	storageType := "gcs"
	return filestorage.UploadMetadata{
		FileName:    fileName,
		FilePath:    filePath,
		FileSize:    file.Size,
		MimeType:    contentType,
		StorageType: storageType,
	}
}

// Delete removes a file from GCS
// NOTE: This does NOT delete from database - that's the caller's responsibility
func (gcs *GCSStorage) Delete(filePath string, storageType string) error {
	if filePath == "" {
		return fmt.Errorf("invalid_file_path")
	}

	// TODO: Delete from GCS using gcsClient
	// Example:
	// ctx := context.Background()
	// err := gcs.config.Client.Bucket(gcs.config.BucketName).Object(gcsPath).Delete(ctx)

	// For now, just return success (implement when GCS SDK is available)
	return nil
}

// ValidateFileExists checks if file exists in GCS
func (gcs *GCSStorage) ValidateFileExists(metadata filestorage.UploadMetadata) bool {
	// TODO: Implement GCS HEAD request to verify file exists
	return true
}

// GetDownloadURL returns the GCS signed URL for download
func (gcs *GCSStorage) GetDownloadURL(metadata filestorage.UploadMetadata) string {
	// TODO: Generate signed URL when GCS client is available
	return metadata.FilePath
}

// GetProviderName returns the provider name
func (gcs *GCSStorage) GetProviderName() string {
	return "gcs"
}
