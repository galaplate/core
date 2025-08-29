package bootstrap

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/galaplate/core/config"
	"github.com/galaplate/core/database"
	"github.com/galaplate/core/logger"
	"github.com/galaplate/core/queue"
	"github.com/galaplate/core/scheduler"
	"github.com/galaplate/core/utils"
)

// AppConfig holds configuration for creating the Fiber app
type AppConfig struct {
	TemplateDir         string
	TemplateExt         string
	SetupRoutes         func(*fiber.App)
	StartBackgroundJobs bool
	QueueSize           int
	WorkerCount         int
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
	database.ConnectDB()

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
	screet := config.Get("APP_SCREET")
	if screet == "" {
		logger.Fatal("You must generate the screet key first")
	}

	// Connect DB for console commands
	database.ConnectDB()
}
