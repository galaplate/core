package providers

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	s3service "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"

	filestorage "github.com/galaplate/core/file-storage"
)

// S3Config holds S3-specific configuration
type S3Config struct {
	BaseConfig             filestorage.FileUploadConfig
	S3Client               *s3service.Client
	BucketName             string
	Region                 string
	BaseURL                string
	ACL                    string
	StorageClass           string
	DisableACL             bool          // Disable ACL parameter for S3-compatible services
	DisableStorageClass    bool          // Disable StorageClass parameter for S3-compatible services
	PathPrefix             string        // Path prefix for S3 keys (e.g., "uploads", "files/documents")
	PresignedURLExpiration time.Duration // Expiration duration for presigned URLs (default: 15 minutes)
}

type S3Storage struct {
	config S3Config
}

func NewS3Storage(cfg S3Config) *S3Storage {
	return &S3Storage{
		config: cfg,
	}
}

func (s3 *S3Storage) Upload(file *multipart.FileHeader) filestorage.UploadMetadata {
	if file.Size > s3.config.BaseConfig.MaxSize {
		return filestorage.UploadMetadata{
			Error: "file_too_large",
		}
	}

	contentType := file.Header.Get("Content-Type")
	isAllowed := slices.Contains(s3.config.BaseConfig.AllowedTypes, contentType)

	if !isAllowed {
		return filestorage.UploadMetadata{
			Error: "invalid_file_type",
		}
	}

	ext := filepath.Ext(file.Filename)
	uniqueID := uuid.New().String()
	timestamp := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s%s", timestamp, uniqueID, ext)

	pathPrefix := s3.config.PathPrefix
	if pathPrefix == "" {
		pathPrefix = "uploads" // Default prefix
	}
	s3Key := fmt.Sprintf("%s/%s", pathPrefix, fileName)

	src, err := file.Open()
	if err != nil {
		return filestorage.UploadMetadata{
			Error: "file_open_failed",
		}
	}
	defer src.Close()

	putObjectInput := &s3service.PutObjectInput{
		Bucket:      aws.String(s3.config.BucketName),
		Key:         aws.String(s3Key),
		Body:        src,
		ContentType: aws.String(contentType),
	}

	// Only set ACL if not disabled (some S3-compatible services don't support ACL)
	if !s3.config.DisableACL && s3.config.ACL != "" {
		putObjectInput.ACL = types.ObjectCannedACL(s3.config.ACL)
	}

	// Only set StorageClass if not disabled (some S3-compatible services don't support it)
	if !s3.config.DisableStorageClass && s3.config.StorageClass != "" {
		putObjectInput.StorageClass = types.StorageClass(s3.config.StorageClass)
	}

	uploader := manager.NewUploader(s3.config.S3Client)
	ctx := context.Background()
	_, err = uploader.Upload(ctx, putObjectInput)
	if err != nil {
		return filestorage.UploadMetadata{
			Error: fmt.Sprintf("file_save_failed: %s", err.Error()),
		}
	}

	filePath := fmt.Sprintf("%s/%s", s3.config.BaseURL, s3Key)

	storageType := "s3"
	return filestorage.UploadMetadata{
		FileName:    fileName,
		FilePath:    filePath,
		FileSize:    file.Size,
		MimeType:    contentType,
		StorageType: storageType,
	}
}

func (s3 *S3Storage) Delete(filePath string, storageType string) error {
	if filePath == "" {
		return fmt.Errorf("invalid_file_path")
	}

	s3Key := s3.extractS3KeyFromPath(filePath)
	if s3Key == "" {
		return fmt.Errorf("invalid_s3_key")
	}

	ctx := context.Background()
	_, err := s3.config.S3Client.DeleteObject(ctx, &s3service.DeleteObjectInput{
		Bucket: aws.String(s3.config.BucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fmt.Errorf("s3_delete_failed: %w", err)
	}

	return nil
}

func (s3 *S3Storage) extractS3KeyFromPath(filePath string) string {
	baseURL := strings.TrimSuffix(s3.config.BaseURL, "/")
	if after, ok := strings.CutPrefix(filePath, baseURL); ok {
		key := after
		key = strings.TrimPrefix(key, "/")
		return key
	}
	return filePath
}

func (s3 *S3Storage) ValidateFileExists(metadata filestorage.UploadMetadata) bool {
	if metadata.FilePath == "" {
		return false
	}

	s3Key := s3.extractS3KeyFromPath(metadata.FilePath)
	if s3Key == "" {
		return false
	}

	ctx := context.Background()
	_, err := s3.config.S3Client.HeadObject(ctx, &s3service.HeadObjectInput{
		Bucket: aws.String(s3.config.BucketName),
		Key:    aws.String(s3Key),
	})

	return err == nil
}

func (s3 *S3Storage) GetDownloadURL(metadata filestorage.UploadMetadata) string {
	if metadata.FilePath == "" {
		return ""
	}

	s3Key := s3.extractS3KeyFromPath(metadata.FilePath)
	if s3Key == "" {
		return metadata.FilePath
	}

	presignClient := s3service.NewPresignClient(s3.config.S3Client)

	// Use configured expiration or default to 15 minutes
	expiration := s3.config.PresignedURLExpiration
	if expiration == 0 {
		expiration = 15 * time.Minute
	}

	ctx := context.Background()
	presignedReq, err := presignClient.PresignGetObject(ctx, &s3service.GetObjectInput{
		Bucket: aws.String(s3.config.BucketName),
		Key:    aws.String(s3Key),
	}, s3service.WithPresignExpires(expiration))

	if err != nil {
		return metadata.FilePath
	}

	return presignedReq.URL
}

func (s3 *S3Storage) GetProviderName() string {
	return "s3"
}
