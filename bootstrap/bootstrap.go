package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/galaplate/core/config"
	"github.com/galaplate/core/database"
	filestorage "github.com/galaplate/core/file-storage"
	fileStorageConfig "github.com/galaplate/core/file-storage/config"
	fileStorageFactory "github.com/galaplate/core/file-storage/factory"
	"github.com/galaplate/core/file-storage/providers"
	"github.com/galaplate/core/logger"
	"github.com/galaplate/core/queue"
	"github.com/galaplate/core/scheduler"
	"github.com/galaplate/core/supports"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"gorm.io/gorm"
)

// AppConfig holds configuration for creating the Fiber app
type AppConfig struct {
	SetupRoutes         func(*fiber.App)
	StartBackgroundJobs bool
	QueueSize           int
	WorkerCount         int
	GormConfig          *gorm.Config
	FiberConfig         *fiber.Config
	IsConsoleMode       bool
	ConfigPath          string
	ShutdownTimeout     time.Duration
}

// OptFunc is a functional option for configuring AppConfig
type OptFunc func(*AppConfig)

// AppInstance holds the app and its background services for lifecycle management
type AppInstance struct {
	Fiber     *fiber.App
	Queue     *queue.Queue
	Scheduler *scheduler.Scheduler
}

// Shutdown gracefully shuts down the application
func (ai *AppInstance) Shutdown(ctx context.Context) error {
	if err := ai.Fiber.Shutdown(); err != nil {
		logger.Error("bootstrap@Shutdown", map[string]any{
			"component": "fiber",
			"error":     err.Error(),
		})
	}

	if ai.Queue != nil {
		if err := ai.Queue.Shutdown(ctx); err != nil {
			logger.Error("bootstrap@Shutdown", map[string]any{
				"component": "queue",
				"error":     err.Error(),
			})
		}
	}

	if ai.Scheduler != nil {
		if err := ai.Scheduler.Shutdown(ctx); err != nil {
			logger.Error("bootstrap@Shutdown", map[string]any{
				"component": "scheduler",
				"error":     err.Error(),
			})
		}
	}

	return nil
}

// DefaultConfig returns default configuration
func DefaultConfig() *AppConfig {
	isConsoleMode := len(os.Args) > 1 && os.Args[1] == "console"

	return &AppConfig{
		StartBackgroundJobs: true,
		QueueSize:           100,
		WorkerCount:         5,
		IsConsoleMode:       isConsoleMode,
		ConfigPath:          "./config",
		ShutdownTimeout:     30 * time.Second,
		FiberConfig: &fiber.Config{
			Views: html.New("./templates", ".html"),
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				var errResponse supports.GlobalErrorHandlerResp
				if json.Unmarshal([]byte(err.Error()), &errResponse) != nil {
					var e *fiber.Error
					code := fiber.StatusInternalServerError
					message := fmt.Sprintf("Internal Server Error: %v", err.Error())
					if errors.As(err, &e) {
						code = e.Code
						message = e.Message
					}
					logger.Error("bootstrap@DefaultConfig", map[string]any{
						"error": err.Error(),
					})

					return c.Status(code).JSON(fiber.Map{
						"success": false,
						"message": message,
						"error":   message,
					})
				}
				return c.Status(errResponse.Status).JSON(errResponse)
			},
		},
	}
}

// NewApp creates a new Fiber app with optional configurations
// Graceful shutdown is automatically set up when background jobs are enabled
func NewApp(opts ...OptFunc) *fiber.App {
	cfg := DefaultConfig()

	for _, opt := range opts {
		opt(cfg)
	}

	appInstance := appWithConfig(cfg)
	return appInstance.Fiber
}

func appWithConfig(cfg *AppConfig) *AppInstance {
	// Load configuration from config files
	loader := config.NewLoader(cfg.ConfigPath)
	configData, err := loader.Load()
	if err != nil {
		logger.Warn(fmt.Sprintf("Config loading warning: %v", err))
	}

	config.InitializeGlobal(configData)

	secret := config.ConfigString("app.key")
	if secret == "" {
		panic("You must generate the secret key first")
	}

	var fiberCfg *fiber.Config
	if cfg.FiberConfig != nil {
		fiberCfg = cfg.FiberConfig
	}

	app := fiber.New(*fiberCfg)
	if cfg.GormConfig != nil {
		database.New(func(c *database.Config) {
			c.GormConfig = cfg.GormConfig
		})
	} else {
		database.New()
	}

	// Initialize file storage
	initializeFileStorage()

	if cfg.SetupRoutes != nil {
		cfg.SetupRoutes(app)
	}

	appInstance := &AppInstance{
		Fiber: app,
	}

	if cfg.StartBackgroundJobs && !cfg.IsConsoleMode {
		q := queue.New(cfg.QueueSize)
		q.Start(cfg.WorkerCount)
		appInstance.Queue = q

		sch := scheduler.New()
		sch.RunTasks()
		sch.Start()
		appInstance.Scheduler = sch

		setupGracefulShutdown(app, appInstance, cfg.ShutdownTimeout)
	}

	return appInstance
}

// initializeFileStorage initializes the file storage factory with configured providers
func initializeFileStorage() {
	// Load configuration from environment/YAML
	cfg, err := fileStorageConfig.LoadFromEnv()
	if err != nil {
		logger.Fatal("Failed to load file storage config", map[string]any{
			"error": err.Error(),
		})
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Fatal("File storage configuration validation failed", map[string]any{
			"error": err.Error(),
		})
	}

	// Initialize storage providers based on configuration
	providersMap := make(map[string]filestorage.FileStorageProvider)

	for name, driverCfg := range cfg.Drivers {
		switch driver := driverCfg.(type) {
		case *fileStorageConfig.LocalConfig:
			// Create local storage provider
			localConfig := filestorage.FileUploadConfig{
				UploadDir:    driver.Path,
				MaxSize:      driver.MaxSize,
				AllowedTypes: driver.AllowedTypes,
			}
			providersMap[name] = providers.NewLocalStorage(localConfig)
			logger.Info("Local storage provider initialized", map[string]any{
				"name": name,
				"path": driver.Path,
			})

		case *fileStorageConfig.S3Config:
			// Create S3 storage provider with AWS SDK integration
			// Initialize AWS credentials
			awsCfg := aws.Config{
				Region: driver.Region,
				Credentials: credentials.NewStaticCredentialsProvider(
					driver.AccessKey,
					driver.SecretKey,
					"",
				),
			}

			// Configure S3 client options
			s3ClientOptions := []func(*s3.Options){
				func(o *s3.Options) {
					o.Region = driver.Region
					o.Credentials = awsCfg.Credentials
				},
			}

			// Add custom endpoint if specified (for S3-compatible services like SeaweedFS, MinIO)
			if driver.Endpoint != "" {
				s3ClientOptions = append(s3ClientOptions, func(o *s3.Options) {
					o.BaseEndpoint = aws.String(driver.Endpoint)
				})
			}

			// Enable path-style URLs if specified (required for some S3-compatible services)
			if driver.UsePathStyleEndpoint {
				s3ClientOptions = append(s3ClientOptions, func(o *s3.Options) {
					o.UsePathStyle = true
				})
			}

			// Create S3 client
			s3Client := s3.NewFromConfig(awsCfg, s3ClientOptions...)

			s3Config := providers.S3Config{
				BaseConfig: filestorage.FileUploadConfig{
					MaxSize:      driver.MaxSize,
					AllowedTypes: driver.AllowedTypes,
				},
				S3Client:               s3Client,
				BucketName:             driver.Bucket,
				Region:                 driver.Region,
				BaseURL:                driver.BaseURL,
				ACL:                    driver.ACL,
				StorageClass:           driver.StorageClass,
				DisableACL:             driver.DisableACL,
				DisableStorageClass:    driver.DisableStorageClass,
				PathPrefix:             driver.PathPrefix,
				PresignedURLExpiration: driver.PresignedURLExpiration,
			}
			providersMap[name] = providers.NewS3Storage(s3Config)
			logger.Info("S3 storage provider initialized", map[string]any{
				"name":                   name,
				"bucket":                 driver.Bucket,
				"region":                 driver.Region,
				"path_prefix":            driver.PathPrefix,
				"endpoint":               driver.Endpoint,
				"use_path_style":         driver.UsePathStyleEndpoint,
				"presigned_url_duration": driver.PresignedURLExpiration.String(),
			})

		case *fileStorageConfig.GCSConfig:
			// Create GCS storage provider
			// Note: GCS Client initialization is skipped here - add GCP SDK integration if needed
			gcsConfig := providers.GCSConfig{
				BaseConfig: filestorage.FileUploadConfig{
					MaxSize:      driver.MaxSize,
					AllowedTypes: driver.AllowedTypes,
				},
				BucketName: driver.Bucket,
				ProjectID:  driver.Project,
				BaseURL:    driver.BaseURL,
			}
			providersMap[name] = providers.NewGCSStorage(gcsConfig)
			logger.Info("GCS storage provider initialized", map[string]any{
				"name":    name,
				"bucket":  driver.Bucket,
				"project": driver.Project,
			})

		case *fileStorageConfig.GoogleDriveConfig:
			// Create Google Drive storage provider
			gdConfig := filestorage.FileUploadConfig{
				MaxSize:      driver.MaxSize,
				AllowedTypes: driver.AllowedTypes,
			}
			gdStorage, err := providers.NewGoogleDriveStorage(
				driver.ServiceAccountFile,
				driver.FolderID,
				gdConfig,
			)
			if err != nil {
				logger.Error("Failed to initialize Google Drive storage", map[string]any{
					"name":   name,
					"error":  err.Error(),
					"folder": driver.FolderID,
				})
			} else {
				providersMap[name] = gdStorage
				logger.Info("Google Drive storage provider initialized", map[string]any{
					"name":   name,
					"folder": driver.FolderID,
				})
			}
		}
	}

	// Initialize global factory with default driver
	fileStorageFactory.Initialize(cfg.Default, providersMap)

	logger.Info("File storage factory initialized", map[string]any{
		"default_driver":  cfg.Default,
		"total_providers": len(providersMap),
	})
}

// setupGracefulShutdown sets up OS signal handling for graceful shutdown
func setupGracefulShutdown(app *fiber.App, appInstance *AppInstance, shutdownTimeout time.Duration) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received signal: %v, starting graceful shutdown", map[string]any{
			"signal": sig.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := appInstance.Shutdown(ctx); err != nil {
			logger.Error("Error during graceful shutdown", map[string]any{
				"error": err.Error(),
			})
		} else {
			logger.Info("Application shut down gracefully")
		}
	}()
}
