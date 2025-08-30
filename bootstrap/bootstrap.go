package bootstrap

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/galaplate/core/database"
	config "github.com/galaplate/core/env"
	"github.com/galaplate/core/logger"
	"github.com/galaplate/core/queue"
	"github.com/galaplate/core/scheduler"
	"github.com/galaplate/core/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"gorm.io/gorm"
)

// GormConfig holds GORM-specific configuration
type GormConfig struct {
	gorm.Config

	// Additional logger-specific configs for easier configuration
	SlowThreshold             time.Duration
	LogLevel                  string // Silent, Error, Warn, Info
	IgnoreRecordNotFoundError bool
	ParameterizedQueries      bool
	Colorful                  bool
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	GormConfig *GormConfig
}

// AppConfig holds configuration for creating the Fiber app
type AppConfig struct {
	TemplateDir         string
	TemplateExt         string
	SetupRoutes         func(*fiber.App)
	StartBackgroundJobs bool
	QueueSize           int
	WorkerCount         int
	DatabaseConfig      *DatabaseConfig
}

// DefaultGormConfig returns default GORM configuration
func DefaultGormConfig() *GormConfig {
	return &GormConfig{
		Config: gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		},
		SlowThreshold:             time.Second,
		LogLevel:                  "Warn",
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      true,
		Colorful:                  true,
	}
}

// DefaultConfig returns default configuration
func DefaultConfig() *AppConfig {
	return &AppConfig{
		TemplateDir:         "./templates",
		TemplateExt:         ".html",
		StartBackgroundJobs: true,
		QueueSize:           100,
		WorkerCount:         5,
	}
}

func App(cfg *AppConfig) *fiber.App {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	screet := config.Get("APP_SCREET")
	if screet == "" {
		logger.Fatal("You must generate the screet key first")
	}

	engine := html.New(cfg.TemplateDir, cfg.TemplateExt)

	app := fiber.New(fiber.Config{
		Views: engine,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var errResponse utils.GlobalErrorHandlerResp
			if json.Unmarshal([]byte(err.Error()), &errResponse) != nil {
				var e *fiber.Error
				code := fiber.StatusInternalServerError
				message := "Internal Server Error"
				if errors.As(err, &e) {
					code = e.Code
					message = e.Message
				}
				logger.Error(err)

				return c.Status(code).JSON(fiber.Map{
					"success": false,
					"message": message,
					"error":   message,
				})
			}
			return c.Status(errResponse.Status).JSON(errResponse)
		},
	})

	// Connect DB (can be swapped with test DB)
	if cfg.DatabaseConfig != nil && cfg.DatabaseConfig.GormConfig != nil {
		dbConfig := &database.Config{
			GormConfig: &database.GormConfig{
				Config:                    cfg.DatabaseConfig.GormConfig.Config,
				SlowThreshold:             cfg.DatabaseConfig.GormConfig.SlowThreshold,
				LogLevel:                  cfg.DatabaseConfig.GormConfig.LogLevel,
				IgnoreRecordNotFoundError: cfg.DatabaseConfig.GormConfig.IgnoreRecordNotFoundError,
				ParameterizedQueries:      cfg.DatabaseConfig.GormConfig.ParameterizedQueries,
				Colorful:                  cfg.DatabaseConfig.GormConfig.Colorful,
			},
		}

		database.ConnectWithConfig(dbConfig)
	} else {
		database.ConnectDB()
	}

	// Setup routes (provided by application)
	if cfg.SetupRoutes != nil {
		cfg.SetupRoutes(app)
	}

	// Start background jobs (can be skipped in tests)
	if cfg.StartBackgroundJobs {
		q := queue.New(cfg.QueueSize)
		q.Start(cfg.WorkerCount)

		sch := scheduler.New()
		sch.RunTasks()
		sch.Start()
	}

	return app
}

func Init() {
	InitWithConfig(nil)
}

func InitWithConfig(dbConfig *DatabaseConfig) {
	screet := config.Get("APP_SCREET")
	if screet == "" {
		logger.Fatal("You must generate the screet key first")
	}

	// Connect DB for console commands
	if dbConfig != nil && dbConfig.GormConfig != nil {
		dbCfg := &database.Config{
			GormConfig: &database.GormConfig{
				Config:                    dbConfig.GormConfig.Config,
				SlowThreshold:             dbConfig.GormConfig.SlowThreshold,
				LogLevel:                  dbConfig.GormConfig.LogLevel,
				IgnoreRecordNotFoundError: dbConfig.GormConfig.IgnoreRecordNotFoundError,
				ParameterizedQueries:      dbConfig.GormConfig.ParameterizedQueries,
				Colorful:                  dbConfig.GormConfig.Colorful,
			},
		}

		database.ConnectWithConfig(dbCfg)
	} else {
		database.ConnectDB()
	}
}
