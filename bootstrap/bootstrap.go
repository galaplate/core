package bootstrap

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/galaplate/core/database"
	config "github.com/galaplate/core/env"
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
}

// OptFunc is a functional option for configuring AppConfig
type OptFunc func(*AppConfig)

// DefaultConfig returns default configuration
func DefaultConfig() *AppConfig {
	// Auto-detect console mode based on command line arguments
	isConsoleMode := len(os.Args) > 1 && os.Args[1] == "console"

	return &AppConfig{
		StartBackgroundJobs: true,
		QueueSize:           100,
		WorkerCount:         5,
		IsConsoleMode:       isConsoleMode,
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
					logger.Error(err)

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
func NewApp(opts ...OptFunc) *fiber.App {
	cfg := DefaultConfig()

	for _, opt := range opts {
		opt(cfg)
	}

	return appWithConfig(cfg)
}

func appWithConfig(cfg *AppConfig) *fiber.App {
	screet := config.Get("APP_SECRET")
	if screet == "" {
		panic("You must generate the screet key first")
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

	if cfg.SetupRoutes != nil {
		cfg.SetupRoutes(app)
	}

	// Disable background jobs when running console commands
	if cfg.StartBackgroundJobs && !cfg.IsConsoleMode {
		q := queue.New(cfg.QueueSize)
		q.Start(cfg.WorkerCount)

		sch := scheduler.New()
		sch.RunTasks()
		sch.Start()
	}

	return app
}
