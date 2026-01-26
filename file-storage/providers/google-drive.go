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

type GoogleDriveStorage struct {
	config   filestorage.FileUploadConfig
	service  *drive.Service
	folderID string
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewGoogleDriveStorage(serviceAccountJSON string, folderID string, config ...filestorage.FileUploadConfig) (*GoogleDriveStorage, error) {
	cfg := filestorage.DefaultFileUploadConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	ctx := context.Background()

	jsonData, err := os.ReadFile(serviceAccountJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account file: %w", err)
	}

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

func (gd *GoogleDriveStorage) Upload(file *multipart.FileHeader) filestorage.UploadMetadata {
	if file.Size > gd.config.MaxSize {
		return filestorage.UploadMetadata{
			Error: "file_too_large",
		}
	}

	contentType := file.Header.Get("Content-Type")
	isAllowed := slices.Contains(gd.config.AllowedTypes, contentType)

	if !isAllowed {
		return filestorage.UploadMetadata{
			Error: "invalid_file_type",
		}
	}

	ext := filepath.Ext(file.Filename)
	uniqueID := uuid.New().String()
	timestamp := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s%s", timestamp, uniqueID, ext)

	src, err := file.Open()
	if err != nil {
		return filestorage.UploadMetadata{
			Error: "file_open_failed",
		}
	}
	defer src.Close()

	driveFile := &drive.File{
		Name:     fileName,
		MimeType: contentType,
	}

	if gd.folderID != "" {
		driveFile.Parents = []string{gd.folderID}
	}

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

func (gd *GoogleDriveStorage) Delete(filePath string, storageType string) error {
	if filePath == "" {
		return fmt.Errorf("invalid_file_path")
	}

	ctx, cancel := context.WithTimeout(gd.ctx, 30*time.Second)
	defer cancel()

	if err := gd.service.Files.Delete(filePath).Context(ctx).Do(); err != nil {
		return fmt.Errorf("google_drive_delete_failed: %w", err)
	}

	return nil
}

func (gd *GoogleDriveStorage) ValidateFileExists(metadata filestorage.UploadMetadata) bool {
	if metadata.FilePath == "" {
		return false
	}

	ctx, cancel := context.WithTimeout(gd.ctx, 10*time.Second)
	defer cancel()

	_, err := gd.service.Files.Get(metadata.FilePath).Fields("id").Context(ctx).Do()
	return err == nil
}

func (gd *GoogleDriveStorage) GetDownloadURL(metadata filestorage.UploadMetadata) string {
	if metadata.FilePath == "" {
		return ""
	}

	return fmt.Sprintf("https://drive.google.com/uc?id=%s&export=download", metadata.FilePath)
}

func (gd *GoogleDriveStorage) GetProviderName() string {
	return "google_drive"
}

func (gd *GoogleDriveStorage) GetConfig() filestorage.FileUploadConfig {
	return gd.config
}

func (gd *GoogleDriveStorage) SetConfig(config filestorage.FileUploadConfig) {
	gd.config = config
}

func (gd *GoogleDriveStorage) Close() error {
	gd.cancel()
	return nil
}
