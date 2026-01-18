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

	"github.com/galaplate/core/config"
	"github.com/galaplate/core/database"
	"github.com/galaplate/core/file-storage"
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
	// Get filesystem driver from config (defaults to "local")
	driver := config.ConfigString("filesystems.default")
	if driver == "" {
		driver = "local"
	}

	// Get max file size from config (defaults to 10MB)
	maxSize := config.ConfigInt("filesystems.max_size")
	if maxSize <= 0 {
		maxSize = int(filestorage.DefaultFileUploadConfig.MaxSize)
	}

	// Create file upload configuration
	fileConfig := filestorage.FileUploadConfig{
		MaxSize:      int64(maxSize),
		AllowedTypes: filestorage.DefaultFileUploadConfig.AllowedTypes,
	}

	// Initialize storage providers
	providersMap := make(map[string]filestorage.FileStorageProvider)

	// Local storage provider
	uploadDir := config.ConfigString("filesystems.disks.local.path")
	if uploadDir == "" {
		uploadDir = filestorage.DefaultFileUploadConfig.UploadDir
	}
	fileConfig.UploadDir = uploadDir
	providersMap["local"] = providers.NewLocalStorage(fileConfig)

	// S3 storage provider
	s3Bucket := config.ConfigString("filesystems.disks.s3.bucket")
	if s3Bucket != "" {
		logger.Warn("S3 storage provider available but not fully integrated", map[string]any{
			"bucket": s3Bucket,
		})
	}

	// Google Cloud Storage provider
	gcsBucket := config.ConfigString("filesystems.disks.gcs.bucket")
	if gcsBucket != "" {
		logger.Warn("Google Cloud Storage provider available but not fully integrated", map[string]any{
			"bucket": gcsBucket,
		})
	}

	// Google Drive storage provider
	gdFolderID := config.ConfigString("filesystems.disks.google_drive.folder_id")
	if gdFolderID != "" {
		gdServiceAccountFile := config.ConfigString("filesystems.disks.google_drive.service_account_file")

		gdConfig := &filestorage.FileUploadConfig{
			MaxSize:      fileConfig.MaxSize,
			AllowedTypes: fileConfig.AllowedTypes,
		}

		gdStorage, err := providers.NewGoogleDriveStorage(gdServiceAccountFile, gdFolderID, *gdConfig)
		if err != nil {
			logger.Error("Failed to initialize Google Drive storage", map[string]any{
				"error":  err.Error(),
				"folder": gdFolderID,
			})
		} else {
			providersMap["google_drive"] = gdStorage
		}
	}

	// Initialize global factory with selected driver
	fileStorageFactory.Initialize(driver, providersMap)

	logger.Info("File storage initialized", map[string]any{
		"driver":    driver,
		"providers": len(providersMap),
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
