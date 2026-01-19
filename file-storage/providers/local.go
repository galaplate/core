package providers

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/google/uuid"

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

// Upload stores a file on local disk and returns metadata
// NOTE: This does NOT save to database - that's the caller's responsibility
func (ls *LocalStorage) Upload(file *multipart.FileHeader) filestorage.UploadMetadata {
	// Validate file size
	if file.Size > ls.config.MaxSize {
		return filestorage.UploadMetadata{
			Error: "file_too_large",
		}
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	isAllowed := slices.Contains(ls.config.AllowedTypes, contentType)

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

	// Create upload directory if not exists
	uploadDir := ls.config.UploadDir
	if !filepath.IsAbs(uploadDir) {
		cwd, err := os.Getwd()
		if err == nil {
			uploadDir = filepath.Join(cwd, uploadDir)
		}
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return filestorage.UploadMetadata{
			Error: "directory_creation_failed",
		}
	}

	// Save file to disk
	filePath := filepath.Join(uploadDir, fileName)
	src, err := file.Open()
	if err != nil {
		return filestorage.UploadMetadata{
			Error: "file_open_failed",
		}
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return filestorage.UploadMetadata{
			Error: "file_create_failed",
		}
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(src); err != nil {
		os.Remove(filePath) // Clean up on failure
		return filestorage.UploadMetadata{
			Error: "file_save_failed",
		}
	}

	// Return metadata only - caller is responsible for saving to database
	storageType := "local"
	return filestorage.UploadMetadata{
		FileName:    fileName,
		FilePath:    filePath,
		FileSize:    file.Size,
		MimeType:    contentType,
		StorageType: storageType,
	}
}

// Delete removes a file from disk
// NOTE: This does NOT delete from database - that's the caller's responsibility
func (ls *LocalStorage) Delete(filePath string, storageType string) error {
	if filePath == "" {
		return fmt.Errorf("invalid_file_path")
	}

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("file_delete_failed: %w", err)
	}

	return nil
}

// ValidateFileExists checks if a file exists on disk
func (ls *LocalStorage) ValidateFileExists(metadata filestorage.UploadMetadata) bool {
	if metadata.FilePath == "" {
		return false
	}

	_, err := os.Stat(metadata.FilePath)
	return !os.IsNotExist(err)
}

// GetDownloadURL returns the local file path for download
func (ls *LocalStorage) GetDownloadURL(metadata filestorage.UploadMetadata) string {
	return metadata.FilePath
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
