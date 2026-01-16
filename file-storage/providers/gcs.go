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

// Upload stores a file to GCS and saves metadata to database
func (gcs *GCSStorage) Upload(file *multipart.FileHeader, userID uint, db *gorm.DB) filestorage.FileUploadResult {
	// Validate file size
	if file.Size > gcs.config.BaseConfig.MaxSize {
		return filestorage.FileUploadResult{
			Error: "file_too_large",
		}
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	isAllowed := slices.Contains(gcs.config.BaseConfig.AllowedTypes, contentType)

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

	// Create GCS object path
	gcsPath := fmt.Sprintf("uploads/%s", fileName)

	// Open file for upload
	src, err := file.Open()
	if err != nil {
		return filestorage.FileUploadResult{
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
		// TODO: Delete from GCS if database save fails
		return filestorage.FileUploadResult{
			Error: "database_save_failed",
		}
	}

	return filestorage.FileUploadResult{
		FileUpload: &fileMetadata,
	}
}

// Delete removes a file from GCS and database
func (gcs *GCSStorage) Delete(fileID uint, userID uint, db *gorm.DB) error {
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

	// TODO: Delete from GCS using gcsClient
	// filePath, ok := fileData["file_path"].(string)
	// Extract GCS path from filePath and delete

	// Delete from database
	if err := db.Table("file_uploads").Delete(map[string]interface{}{"id": fileID}).Error; err != nil {
		return fmt.Errorf("database_delete_failed")
	}

	return nil
}

// GetFileByID retrieves file metadata by ID
func (gcs *GCSStorage) GetFileByID(fileID uint, db *gorm.DB) (any, error) {
	var fileData map[string]interface{}
	if err := db.Table("file_uploads").First(&fileData, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("file_not_found")
		}
		return nil, fmt.Errorf("database_error")
	}

	return fileData, nil
}

// ValidateFileExists checks if file exists in GCS
func (gcs *GCSStorage) ValidateFileExists(fileUpload any) bool {
	// TODO: Implement GCS HEAD request to verify file exists
	return true
}

// GetDownloadPath returns the GCS signed URL for download
func (gcs *GCSStorage) GetDownloadPath(fileUpload any) string {
	fileData, ok := fileUpload.(map[string]interface{})
	if !ok {
		return ""
	}

	// TODO: Generate signed URL when GCS client is available
	filePath, ok := fileData["file_path"].(string)
	if !ok {
		return ""
	}

	return filePath
}

// GetProviderName returns the provider name
func (gcs *GCSStorage) GetProviderName() string {
	return "gcs"
}
