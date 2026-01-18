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
	"gorm.io/gorm"

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

// Upload stores a file on Google Drive and saves metadata to database
func (gd *GoogleDriveStorage) Upload(file *multipart.FileHeader, userID uint, db *gorm.DB) filestorage.FileUploadResult {
	// Validate file size
	if file.Size > gd.config.MaxSize {
		return filestorage.FileUploadResult{
			Error: "file_too_large",
		}
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	isAllowed := slices.Contains(gd.config.AllowedTypes, contentType)

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

	// Open file
	src, err := file.Open()
	if err != nil {
		return filestorage.FileUploadResult{
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
		return filestorage.FileUploadResult{
			Error: "google_drive_upload_failed",
		}
	}

	// Create file metadata for database
	fileMetadata := FileMetadata{
		OriginalName: file.Filename,
		FileName:     fileName,
		FilePath:     res.Id, // Store Google Drive file ID as path
		FileSize:     file.Size,
		MimeType:     contentType,
		UploadedBy:   userID,
	}

	// Add custom field for Google Drive URL
	fileMetadataWithURL := map[string]interface{}{
		"original_name":    fileMetadata.OriginalName,
		"file_name":        fileMetadata.FileName,
		"file_path":        fileMetadata.FilePath,
		"file_size":        fileMetadata.FileSize,
		"mime_type":        fileMetadata.MimeType,
		"uploaded_by":      fileMetadata.UploadedBy,
		"storage_type":     "google_drive",
		"google_drive_id":  res.Id,
		"google_drive_url": res.WebViewLink,
		"download_url":     res.WebContentLink,
	}

	// Save file info to database
	if err := db.Table("file_uploads").Create(fileMetadataWithURL).Error; err != nil {
		// Try to delete from Google Drive if database save fails
		deleteCtx, deleteCancel := context.WithTimeout(gd.ctx, 30*time.Second)
		gd.service.Files.Delete(res.Id).Context(deleteCtx).Do()
		deleteCancel()

		return filestorage.FileUploadResult{
			Error: "database_save_failed",
		}
	}

	// Return the metadata with Google Drive info
	return filestorage.FileUploadResult{
		FileUpload: &fileMetadata,
	}
}

// Delete removes a file from Google Drive and its database record
func (gd *GoogleDriveStorage) Delete(fileID uint, userID uint, db *gorm.DB) error {
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

	// Get Google Drive file ID
	filePath, ok := fileData["file_path"].(string)
	if !ok {
		return fmt.Errorf("invalid_file_path")
	}

	// Delete from Google Drive
	ctx, cancel := context.WithTimeout(gd.ctx, 30*time.Second)
	defer cancel()

	if err := gd.service.Files.Delete(filePath).Context(ctx).Do(); err != nil {
		// Log the error but continue with database deletion
		fmt.Printf("warning: failed to delete from Google Drive: %v\n", err)
	}

	// Delete from database
	if err := db.Table("file_uploads").Delete(map[string]interface{}{"id": fileID}).Error; err != nil {
		return fmt.Errorf("database_delete_failed")
	}

	return nil
}

// GetFileByID retrieves file metadata by ID
func (gd *GoogleDriveStorage) GetFileByID(fileID uint, db *gorm.DB) (any, error) {
	var fileData map[string]interface{}
	if err := db.Table("file_uploads").First(&fileData, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("file_not_found")
		}
		return nil, fmt.Errorf("database_error")
	}

	return fileData, nil
}

// ValidateFileExists checks if a file exists on Google Drive
func (gd *GoogleDriveStorage) ValidateFileExists(fileUpload any) bool {
	fileData, ok := fileUpload.(map[string]interface{})
	if !ok {
		return false
	}

	filePath, ok := fileData["file_path"].(string)
	if !ok {
		return false
	}

	ctx, cancel := context.WithTimeout(gd.ctx, 10*time.Second)
	defer cancel()

	_, err := gd.service.Files.Get(filePath).Fields("id").Context(ctx).Do()
	return err == nil
}

// GetDownloadPath returns the download URL for Google Drive files
func (gd *GoogleDriveStorage) GetDownloadPath(fileUpload any) string {
	fileData, ok := fileUpload.(map[string]interface{})
	if !ok {
		return ""
	}

	// Try to get direct download URL first
	if downloadURL, ok := fileData["download_url"].(string); ok && downloadURL != "" {
		return downloadURL
	}

	// Fall back to Google Drive file ID
	filePath, ok := fileData["file_path"].(string)
	if !ok {
		return ""
	}

	// Return Google Drive direct download URL
	return fmt.Sprintf("https://drive.google.com/uc?id=%s&export=download", filePath)
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
