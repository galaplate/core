package providers

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"slices"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

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

// Upload stores a file to S3 and saves metadata to database
func (s3 *S3Storage) Upload(file *multipart.FileHeader, userID uint, db *gorm.DB) filestorage.FileUploadResult {
	// Validate file size
	if file.Size > s3.config.BaseConfig.MaxSize {
		return filestorage.FileUploadResult{
			Error: "file_too_large",
		}
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	isAllowed := slices.Contains(s3.config.BaseConfig.AllowedTypes, contentType)

	if !isAllowed {
		return filestorage.FileUploadResult{
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
		return filestorage.FileUploadResult{
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

	// Save file metadata to database
	fileMetadata := FileMetadata{
		OriginalName: file.Filename,
		FileName:     fileName,
		FilePath:     filePath,
		FileSize:     file.Size,
		MimeType:     contentType,
		UploadedBy:   userID,
	}

	if err := db.Table("file_uploads").Create(map[string]interface{}{
		"original_name": fileMetadata.OriginalName,
		"file_name":     fileMetadata.FileName,
		"file_path":     fileMetadata.FilePath,
		"file_size":     fileMetadata.FileSize,
		"mime_type":     fileMetadata.MimeType,
		"uploaded_by":   fileMetadata.UploadedBy,
	}).Error; err != nil {
		// TODO: Delete from S3 if database save fails
		return filestorage.FileUploadResult{
			Error: "database_save_failed",
		}
	}

	return filestorage.FileUploadResult{
		FileUpload: &fileMetadata,
	}
}

// Delete removes a file from S3 and database
func (s3 *S3Storage) Delete(fileID uint, userID uint, db *gorm.DB) error {
	var fileData map[string]interface{}
	if err := db.Table("file_uploads").First(&fileData, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("file_not_found")
		}
		return fmt.Errorf("database_error")
	}

	// Check if user owns the file
	uploadedBy, ok := fileData["uploaded_by"].(uint)
	if !ok {
		if val, ok := fileData["uploaded_by"].(int); ok {
			uploadedBy = uint(val)
		} else {
			return fmt.Errorf("invalid_file_data")
		}
	}

	if uploadedBy != userID {
		return fmt.Errorf("forbidden")
	}

	// TODO: Delete from S3 using s3Client
	// filePath, ok := fileData["file_path"].(string)
	// Extract S3 key from filePath and delete

	// Delete from database
	if err := db.Table("file_uploads").Delete(map[string]interface{}{"id": fileID}).Error; err != nil {
		return fmt.Errorf("database_delete_failed")
	}

	return nil
}

// GetFileByID retrieves file metadata by ID
func (s3 *S3Storage) GetFileByID(fileID uint, db *gorm.DB) (any, error) {
	var fileData map[string]interface{}
	if err := db.Table("file_uploads").First(&fileData, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("file_not_found")
		}
		return nil, fmt.Errorf("database_error")
	}

	return fileData, nil
}

// ValidateFileExists checks if file exists in S3
func (s3 *S3Storage) ValidateFileExists(fileUpload any) bool {
	// TODO: Implement S3 HEAD request to verify file exists
	return true
}

// GetDownloadPath returns the S3 presigned URL for download
func (s3 *S3Storage) GetDownloadPath(fileUpload any) string {
	fileData, ok := fileUpload.(map[string]interface{})
	if !ok {
		return ""
	}

	// TODO: Generate presigned URL when S3 client is available
	filePath, ok := fileData["file_path"].(string)
	if !ok {
		return ""
	}

	return filePath
}

// GetProviderName returns the provider name
func (s3 *S3Storage) GetProviderName() string {
	return "s3"
}
