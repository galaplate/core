package providers

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	filestorage "github.com/galaplate/core/file-storage"
)

// LocalStorage implements FileStorageProvider for local disk storage
type LocalStorage struct {
	config filestorage.FileUploadConfig
}

// NewLocalStorage creates a new local storage provider
func NewLocalStorage(config ...filestorage.FileUploadConfig) *LocalStorage {
	cfg := filestorage.DefaultFileUploadConfig
	if len(config) > 0 {
		cfg = config[0]
	}
	return &LocalStorage{
		config: cfg,
	}
}

// FileMetadata holds generic file metadata for database storage
type FileMetadata struct {
	OriginalName string
	FileName     string
	FilePath     string
	FileSize     int64
	MimeType     string
	UploadedBy   uint
}

// Upload stores a file on local disk and saves metadata to database
func (ls *LocalStorage) Upload(file *multipart.FileHeader, userID uint, db *gorm.DB) filestorage.FileUploadResult {
	// Validate file size
	if file.Size > ls.config.MaxSize {
		return filestorage.FileUploadResult{
			Error: "file_too_large",
		}
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	isAllowed := slices.Contains(ls.config.AllowedTypes, contentType)

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

	// Create upload directory if not exists
	uploadDir := ls.config.UploadDir
	if !filepath.IsAbs(uploadDir) {
		cwd, err := os.Getwd()
		if err == nil {
			uploadDir = filepath.Join(cwd, uploadDir)
		}
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return filestorage.FileUploadResult{
			Error: "directory_creation_failed",
		}
	}

	// Save file to disk
	filePath := filepath.Join(uploadDir, fileName)
	src, err := file.Open()
	if err != nil {
		return filestorage.FileUploadResult{
			Error: "file_open_failed",
		}
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return filestorage.FileUploadResult{
			Error: "file_create_failed",
		}
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(src); err != nil {
		os.Remove(filePath) // Clean up on failure
		return filestorage.FileUploadResult{
			Error: "file_save_failed",
		}
	}

	// Create file metadata - this is a generic struct
	// Applications should map this to their own FileUpload model
	fileMetadata := FileMetadata{
		OriginalName: file.Filename,
		FileName:     fileName,
		FilePath:     filePath,
		FileSize:     file.Size,
		MimeType:     contentType,
		UploadedBy:   userID,
	}

	// Save file info to database using the generic model
	// Applications will handle the actual model type via model hooks or wrappers
	if err := db.Table("file_uploads").Create(map[string]interface{}{
		"original_name": fileMetadata.OriginalName,
		"file_name":     fileMetadata.FileName,
		"file_path":     fileMetadata.FilePath,
		"file_size":     fileMetadata.FileSize,
		"mime_type":     fileMetadata.MimeType,
		"uploaded_by":   fileMetadata.UploadedBy,
	}).Error; err != nil {
		// Remove uploaded file if database save fails
		os.Remove(filePath)
		return filestorage.FileUploadResult{
			Error: "database_save_failed",
		}
	}

	// Return the metadata - let the application convert to their model type
	return filestorage.FileUploadResult{
		FileUpload: &fileMetadata,
	}
}

// Delete removes a file from disk and its database record
func (ls *LocalStorage) Delete(fileID uint, userID uint, db *gorm.DB) error {
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
		// Try to convert from other types
		if val, ok := fileData["uploaded_by"].(int); ok {
			uploadedBy = uint(val)
		} else {
			return fmt.Errorf("invalid_file_data")
		}
	}

	if uploadedBy != userID {
		return fmt.Errorf("forbidden")
	}

	// Delete file from disk
	filePath, ok := fileData["file_path"].(string)
	if !ok {
		return fmt.Errorf("invalid_file_path")
	}

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("file_delete_failed")
	}

	// Delete from database
	if err := db.Table("file_uploads").Delete(map[string]interface{}{"id": fileID}).Error; err != nil {
		return fmt.Errorf("database_delete_failed")
	}

	return nil
}

// GetFileByID retrieves file metadata by ID
func (ls *LocalStorage) GetFileByID(fileID uint, db *gorm.DB) (any, error) {
	var fileData map[string]interface{}
	if err := db.Table("file_uploads").First(&fileData, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("file_not_found")
		}
		return nil, fmt.Errorf("database_error")
	}

	return fileData, nil
}

// ValidateFileExists checks if a file exists on disk
func (ls *LocalStorage) ValidateFileExists(fileUpload any) bool {
	fileData, ok := fileUpload.(map[string]interface{})
	if !ok {
		return false
	}

	filePath, ok := fileData["file_path"].(string)
	if !ok {
		return false
	}

	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// GetDownloadPath returns the local file path for download
func (ls *LocalStorage) GetDownloadPath(fileUpload any) string {
	fileData, ok := fileUpload.(map[string]interface{})
	if !ok {
		return ""
	}

	filePath, ok := fileData["file_path"].(string)
	if !ok {
		return ""
	}

	return filePath
}

// GetProviderName returns the provider name
func (ls *LocalStorage) GetProviderName() string {
	return "local"
}

// GetConfig returns the storage configuration
func (ls *LocalStorage) GetConfig() filestorage.FileUploadConfig {
	return ls.config
}

// SetConfig updates the storage configuration
func (ls *LocalStorage) SetConfig(config filestorage.FileUploadConfig) {
	ls.config = config
}
