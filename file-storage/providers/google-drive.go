package providers

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"slices"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	filestorage "github.com/galaplate/core/file-storage"
	"github.com/google/uuid"
)

// GoogleDriveStorage implements FileStorageProvider for Google Drive
type GoogleDriveStorage struct {
	config   filestorage.FileUploadConfig
	service  *drive.Service
	folderID string
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewGoogleDriveStorage creates a new Google Drive storage provider
func NewGoogleDriveStorage(serviceAccountJSON string, folderID string, config ...filestorage.FileUploadConfig) (*GoogleDriveStorage, error) {
	cfg := filestorage.DefaultFileUploadConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	ctx := context.Background()

	// Parse service account JSON
	jsonData, err := os.ReadFile(serviceAccountJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account file: %w", err)
	}

	// Create drive service with service account credentials
	service, err := drive.NewService(ctx, option.WithCredentialsJSON(jsonData), option.WithScopes(drive.DriveScope))
	if err != nil {
		return nil, fmt.Errorf("failed to create drive service: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &GoogleDriveStorage{
		config:   cfg,
		service:  service,
		folderID: folderID,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Upload stores a file on Google Drive and returns metadata
// NOTE: This does NOT save to database - that's the caller's responsibility
func (gd *GoogleDriveStorage) Upload(file *multipart.FileHeader) filestorage.UploadMetadata {
	// Validate file size
	if file.Size > gd.config.MaxSize {
		return filestorage.UploadMetadata{
			Error: "file_too_large",
		}
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	isAllowed := slices.Contains(gd.config.AllowedTypes, contentType)

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

	// Open file
	src, err := file.Open()
	if err != nil {
		return filestorage.UploadMetadata{
			Error: "file_open_failed",
		}
	}
	defer src.Close()

	// Create file metadata for Google Drive
	driveFile := &drive.File{
		Name:     fileName,
		MimeType: contentType,
	}

	// Add parent folder if specified
	if gd.folderID != "" {
		driveFile.Parents = []string{gd.folderID}
	}

	// Upload to Google Drive
	ctx, cancel := context.WithTimeout(gd.ctx, 60*time.Second)
	defer cancel()

	res, err := gd.service.Files.Create(driveFile).
		Media(src).
		Fields("id, webViewLink, webContentLink").
		Context(ctx).
		Do()

	if err != nil {
		return filestorage.UploadMetadata{
			Error: "google_drive_upload_failed",
		}
	}

	// Return metadata only - caller is responsible for saving to database
	storageType := "google_drive"
	return filestorage.UploadMetadata{
		FileName:      fileName,
		FilePath:      res.Id, // Store Google Drive file ID as path
		FileSize:      file.Size,
		MimeType:      contentType,
		StorageType:   storageType,
		GoogleDriveID: &res.Id,
	}
}

// Delete removes a file from Google Drive
// NOTE: This does NOT delete from database - that's the caller's responsibility
func (gd *GoogleDriveStorage) Delete(filePath string, storageType string) error {
	if filePath == "" {
		return fmt.Errorf("invalid_file_path")
	}

	// Delete from Google Drive
	ctx, cancel := context.WithTimeout(gd.ctx, 30*time.Second)
	defer cancel()

	if err := gd.service.Files.Delete(filePath).Context(ctx).Do(); err != nil {
		return fmt.Errorf("google_drive_delete_failed: %w", err)
	}

	return nil
}

// ValidateFileExists checks if a file exists on Google Drive
func (gd *GoogleDriveStorage) ValidateFileExists(metadata filestorage.UploadMetadata) bool {
	if metadata.FilePath == "" {
		return false
	}

	ctx, cancel := context.WithTimeout(gd.ctx, 10*time.Second)
	defer cancel()

	_, err := gd.service.Files.Get(metadata.FilePath).Fields("id").Context(ctx).Do()
	return err == nil
}

// GetDownloadURL returns the download URL for Google Drive files
func (gd *GoogleDriveStorage) GetDownloadURL(metadata filestorage.UploadMetadata) string {
	// Google Drive provides direct download URL in metadata
	// Using file ID as backup method
	if metadata.FilePath == "" {
		return ""
	}

	// Return Google Drive direct download URL
	return fmt.Sprintf("https://drive.google.com/uc?id=%s&export=download", metadata.FilePath)
}

// GetProviderName returns the provider name
func (gd *GoogleDriveStorage) GetProviderName() string {
	return "google_drive"
}

// GetConfig returns the storage configuration
func (gd *GoogleDriveStorage) GetConfig() filestorage.FileUploadConfig {
	return gd.config
}

// SetConfig updates the storage configuration
func (gd *GoogleDriveStorage) SetConfig(config filestorage.FileUploadConfig) {
	gd.config = config
}

// Close closes the context and cleans up resources
func (gd *GoogleDriveStorage) Close() error {
	gd.cancel()
	return nil
}
