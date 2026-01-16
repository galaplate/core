package filestorage

import (
	"mime/multipart"

	"gorm.io/gorm"
)

// FileUploadResult represents the result of a file upload operation
type FileUploadResult struct {
	FileUpload any // Generic file upload model (can be from any package)
	Error      string
}

// FileStorageProvider defines the interface for file storage implementations
type FileStorageProvider interface {
	// Upload stores a file and returns FileUploadResult
	Upload(file *multipart.FileHeader, userID uint, db *gorm.DB) FileUploadResult

	// Delete removes a file and its database record
	Delete(fileID uint, userID uint, db *gorm.DB) error

	// GetFileByID retrieves file metadata by ID
	GetFileByID(fileID uint, db *gorm.DB) (any, error)

	// ValidateFileExists checks if a file exists in storage
	ValidateFileExists(fileUpload any) bool

	// GetDownloadPath returns the path/URL for file download
	GetDownloadPath(fileUpload any) string

	// GetProviderName returns the name of the storage provider
	GetProviderName() string
}

// FileUploadConfig contains configuration for file uploads
type FileUploadConfig struct {
	MaxSize      int64    // Maximum file size in bytes
	AllowedTypes []string // Allowed MIME types
	UploadDir    string   // Upload directory
}

// DefaultFileUploadConfig is the default configuration for file uploads
var DefaultFileUploadConfig = FileUploadConfig{
	MaxSize: 10 * 1024 * 1024, // 10MB
	AllowedTypes: []string{
		"image/jpeg", "image/jpg", "image/png", "image/gif",
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	},
	UploadDir: "storage/app/uploads",
}

// Attachment represents a file attachment
type Attachment struct {
	FileID   uint64 `json:"file_id"`
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	Url      string `json:"url"`
}
