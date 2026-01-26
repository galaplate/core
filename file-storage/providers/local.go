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

type LocalStorage struct {
	config filestorage.FileUploadConfig
}

func NewLocalStorage(config ...filestorage.FileUploadConfig) *LocalStorage {
	cfg := filestorage.DefaultFileUploadConfig
	if len(config) > 0 {
		cfg = config[0]
	}
	return &LocalStorage{
		config: cfg,
	}
}

func (ls *LocalStorage) Upload(file *multipart.FileHeader) filestorage.UploadMetadata {
	if file.Size > ls.config.MaxSize {
		return filestorage.UploadMetadata{
			Error: "file_too_large",
		}
	}

	contentType := file.Header.Get("Content-Type")
	isAllowed := slices.Contains(ls.config.AllowedTypes, contentType)

	if !isAllowed {
		return filestorage.UploadMetadata{
			Error: "invalid_file_type",
		}
	}

	ext := filepath.Ext(file.Filename)
	uniqueID := uuid.New().String()
	timestamp := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s%s", timestamp, uniqueID, ext)

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

	storageType := "local"
	return filestorage.UploadMetadata{
		FileName:    fileName,
		FilePath:    filePath,
		FileSize:    file.Size,
		MimeType:    contentType,
		StorageType: storageType,
	}
}

func (ls *LocalStorage) Delete(filePath string, storageType string) error {
	if filePath == "" {
		return fmt.Errorf("invalid_file_path")
	}

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("file_delete_failed: %w", err)
	}

	return nil
}

func (ls *LocalStorage) ValidateFileExists(metadata filestorage.UploadMetadata) bool {
	if metadata.FilePath == "" {
		return false
	}

	_, err := os.Stat(metadata.FilePath)
	return !os.IsNotExist(err)
}

func (ls *LocalStorage) GetDownloadURL(metadata filestorage.UploadMetadata) string {
	return metadata.FilePath
}

func (ls *LocalStorage) GetProviderName() string {
	return "local"
}

func (ls *LocalStorage) GetConfig() filestorage.FileUploadConfig {
	return ls.config
}

func (ls *LocalStorage) SetConfig(config filestorage.FileUploadConfig) {
	ls.config = config
}
