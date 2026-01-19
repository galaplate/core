package filestorage

import (
	"mime/multipart"
)

// UploadMetadata contains file upload metadata returned by the storage provider
// This is what the provider returns after uploading a file
type UploadMetadata struct {
	FileName      string  // Sanitized filename (generated)
	FilePath      string  // Disk path or cloud identifier
	FileSize      int64   // File size in bytes
	MimeType      string  // MIME type of the file
	StorageType   string  // 'local', 'google_drive', 's3', 'gcs'
	GoogleDriveID *string // Google Drive file ID (if using Google Drive)
	Error         string  // Error message if upload failed
}

// FileUploadResult represents the result of a file upload operation
// Deprecated: Use UploadMetadata instead
type FileUploadResult struct {
	FileUpload any // Generic file upload model (can be from any package)
	Error      string
}

// FileStorageProvider defines the interface for file storage implementations
// v2.0: Decoupled - providers only handle storage, not database operations
type FileStorageProvider interface {
	// Upload stores a file to the storage provider (disk/cloud)
	// Returns metadata only - does NOT save to database
	// The caller is responsible for saving the metadata to database if needed
	Upload(file *multipart.FileHeader) UploadMetadata

	// Delete removes a file from the storage provider
	// filePath: The file path or identifier (returned from Upload)
	// storageType: The storage type ('local', 'google_drive', etc.)
	Delete(filePath string, storageType string) error

	// GetDownloadURL returns a URL/path for downloading the file
	// metadata: The UploadMetadata returned from Upload
	GetDownloadURL(metadata UploadMetadata) string

	// ValidateFileExists checks if a file exists in storage
	// metadata: The UploadMetadata returned from Upload
	ValidateFileExists(metadata UploadMetadata) bool

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
