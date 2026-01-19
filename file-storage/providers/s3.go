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

// S3Config holds S3-specific configuration
type S3Config struct {
	BaseConfig   filestorage.FileUploadConfig
	S3Client     interface{} // *s3.Client - use interface to avoid SDK dependency
	BucketName   string
	Region       string
	BaseURL      string
	ACL          string
	StorageClass string
}

// S3Storage implements FileStorageProvider for AWS S3
type S3Storage struct {
	config S3Config
}

// NewS3Storage creates a new S3 storage provider
// Note: S3Client should be *s3.Client from github.com/aws/aws-sdk-go-v2/service/s3
// Using interface{} to avoid forcing the SDK dependency
func NewS3Storage(cfg S3Config) *S3Storage {
	return &S3Storage{
		config: cfg,
	}
}

// Upload stores a file to S3 and returns metadata
// NOTE: This does NOT save to database - that's the caller's responsibility
func (s3 *S3Storage) Upload(file *multipart.FileHeader) filestorage.UploadMetadata {
	// Validate file size
	if file.Size > s3.config.BaseConfig.MaxSize {
		return filestorage.UploadMetadata{
			Error: "file_too_large",
		}
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	isAllowed := slices.Contains(s3.config.BaseConfig.AllowedTypes, contentType)

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

	// Create S3 key path
	s3Key := fmt.Sprintf("uploads/%s", fileName)

	// Open file for upload
	src, err := file.Open()
	if err != nil {
		return filestorage.UploadMetadata{
			Error: "file_open_failed",
		}
	}
	defer src.Close()

	// TODO: Implement S3 upload using provided s3Client
	// Example implementation when you have AWS SDK:
	// uploader := manager.NewUploader(s3.config.S3Client)
	// result, err := uploader.Upload(context.Background(), &s3.PutObjectInput{...})

	// For now, return placeholder result
	filePath := fmt.Sprintf("%s/%s", s3.config.BaseURL, s3Key)

	// Return metadata only - caller is responsible for saving to database
	storageType := "s3"
	return filestorage.UploadMetadata{
		FileName:    fileName,
		FilePath:    filePath,
		FileSize:    file.Size,
		MimeType:    contentType,
		StorageType: storageType,
	}
}

// Delete removes a file from S3
// NOTE: This does NOT delete from database - that's the caller's responsibility
func (s3 *S3Storage) Delete(filePath string, storageType string) error {
	if filePath == "" {
		return fmt.Errorf("invalid_file_path")
	}

	// TODO: Delete from S3 using s3Client
	// Example:
	// _, err := s3.config.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
	//     Bucket: aws.String(s3.config.BucketName),
	//     Key:    aws.String(s3Key),
	// })

	// For now, just return success (implement when AWS SDK is available)
	return nil
}

// ValidateFileExists checks if file exists in S3
func (s3 *S3Storage) ValidateFileExists(metadata filestorage.UploadMetadata) bool {
	// TODO: Implement S3 HEAD request to verify file exists
	return true
}

// GetDownloadURL returns the S3 presigned URL for download
func (s3 *S3Storage) GetDownloadURL(metadata filestorage.UploadMetadata) string {
	// TODO: Generate presigned URL when S3 client is available
	return metadata.FilePath
}

// GetProviderName returns the provider name
func (s3 *S3Storage) GetProviderName() string {
	return "s3"
}
