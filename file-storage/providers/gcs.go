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

type GCSConfig struct {
	BaseConfig          filestorage.FileUploadConfig
	Client              any // *storage.Client - use interface to avoid SDK dependency
	BucketName          string
	ProjectID           string
	BaseURL             string
	PathPrefix          string        // Path prefix for GCS objects (e.g., "uploads", "files/documents")
	SignedURLExpiration time.Duration // Expiration duration for signed URLs (default: 15 minutes)
}

type GCSStorage struct {
	config GCSConfig
}

func NewGCSStorage(cfg GCSConfig) *GCSStorage {
	return &GCSStorage{
		config: cfg,
	}
}

func (gcs *GCSStorage) Upload(file *multipart.FileHeader) filestorage.UploadMetadata {
	if file.Size > gcs.config.BaseConfig.MaxSize {
		return filestorage.UploadMetadata{
			Error: "file_too_large",
		}
	}

	contentType := file.Header.Get("Content-Type")
	isAllowed := slices.Contains(gcs.config.BaseConfig.AllowedTypes, contentType)

	if !isAllowed {
		return filestorage.UploadMetadata{
			Error: "invalid_file_type",
		}
	}

	ext := filepath.Ext(file.Filename)
	uniqueID := uuid.New().String()
	timestamp := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s%s", timestamp, uniqueID, ext)

	// Create GCS path with configurable prefix
	pathPrefix := gcs.config.PathPrefix
	if pathPrefix == "" {
		pathPrefix = "uploads" // Default prefix
	}
	gcsPath := fmt.Sprintf("%s/%s", pathPrefix, fileName)

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

	storageType := "gcs"
	return filestorage.UploadMetadata{
		FileName:    fileName,
		FilePath:    filePath,
		FileSize:    file.Size,
		MimeType:    contentType,
		StorageType: storageType,
	}
}

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

func (gcs *GCSStorage) ValidateFileExists(metadata filestorage.UploadMetadata) bool {
	// TODO: Implement GCS HEAD request to verify file exists
	return true
}

func (gcs *GCSStorage) GetDownloadURL(metadata filestorage.UploadMetadata) string {
	// TODO: Generate signed URL when GCS client is available
	return metadata.FilePath
}

func (gcs *GCSStorage) GetProviderName() string {
	return "gcs"
}
