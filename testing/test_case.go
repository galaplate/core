package testing

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/galaplate/core/bootstrap"
	"github.com/galaplate/core/config"
	"github.com/galaplate/core/database"
	"github.com/galaplate/core/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TestConfig struct {
	EnvFile           string
	SetupRoutes       func(*fiber.App)
	ConfigPath        string
	RefreshDatabase   bool
	ProjectRootOffset int
	CustomBootstrap   func(*TestCase)
	GormConfig        *gorm.Config
	FiberConfig       *fiber.Config
}

type TestCase struct {
	suite.Suite
	App               *fiber.App
	DB                *gorm.DB
	Config            *TestConfig
	refreshDatabase   bool
	databaseRefreshed bool
	projectRoot       string
}

func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		EnvFile:           ".env.testing",
		ConfigPath:        "./config",
		RefreshDatabase:   false,
		ProjectRootOffset: 1,
	}
}

func NewTestCase(opts ...func(*TestConfig)) *TestCase {
	cfg := DefaultTestConfig()

	for _, opt := range opts {
		opt(cfg)
	}

	return &TestCase{
		Config:          cfg,
		refreshDatabase: cfg.RefreshDatabase,
	}
}

func (tc *TestCase) SetupTest() {
	tc.ensureProjectRoot()
	tc.loadEnvironment()
	tc.handleDatabaseRefresh()
	tc.bootstrapApplication()

	if tc.Config.CustomBootstrap != nil {
		tc.Config.CustomBootstrap(tc)
	}
}

func (tc *TestCase) ensureProjectRoot() {
	if tc.projectRoot != "" {
		if err := os.Chdir(tc.projectRoot); err != nil {
			log.Panicf("Failed to change to project root: %v", err)
		}
		// Reinitialize logger to use correct logs directory
		if err := logger.ReinitializeForTesting(tc.projectRoot); err != nil {
			log.Printf("Warning: Failed to reinitialize logger: %v", err)
		}
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Panicf("Failed to get current directory: %v", err)
	}

	// Try to find project root by looking for go.mod or main.go
	projectRoot := cwd
	for i := 0; i <= 10; i++ {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		if _, err := os.Stat(filepath.Join(projectRoot, "main.go")); err == nil {
			break
		}
		parent := filepath.Join(projectRoot, "..")
		absParent, _ := filepath.Abs(parent)
		if absParent == projectRoot {
			break
		}
		projectRoot = parent
	}

	absProjectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		log.Panicf("Failed to resolve project root: %v", err)
	}

	tc.projectRoot = absProjectRoot

	if err := os.Chdir(absProjectRoot); err != nil {
		log.Panicf("Failed to change to project root: %v", err)
	}

	// Reinitialize logger to use correct logs directory
	if err := logger.ReinitializeForTesting(absProjectRoot); err != nil {
		log.Printf("Warning: Failed to reinitialize logger: %v", err)
	}
}

func (tc *TestCase) loadEnvironment() {
	if tc.Config.EnvFile != "" {
		cwd, _ := os.Getwd()

		possiblePaths := []string{
			filepath.Join(tc.projectRoot, "tests", tc.Config.EnvFile),
			filepath.Join(tc.projectRoot, tc.Config.EnvFile),
			filepath.Join(cwd, tc.Config.EnvFile),
			filepath.Join(cwd, "..", tc.Config.EnvFile),
			filepath.Join(cwd, "..", "..", tc.Config.EnvFile),
			filepath.Join(cwd, "..", "..", "..", tc.Config.EnvFile),
		}

		envLoaded := false
		for _, envPath := range possiblePaths {
			absPath, _ := filepath.Abs(envPath)
			if _, err := os.Stat(absPath); err == nil {
				if err := godotenv.Load(absPath); err != nil {
					log.Printf("Warning: Error loading env file %s: %v", absPath, err)
				} else {
					envLoaded = true
					break
				}
			}
		}

		if !envLoaded {
			log.Printf("Warning: Env file not found. Tried tests/.env.testing and project root")
		}
	}

	os.Setenv("APP_ENV", "testing")

	if os.Getenv("DB_DATABASE") == "" {
		dbName := config.ConfigString("database.connections." + config.ConfigString("database.default") + ".database")
		if dbName == "" {
			dbName = "galaplate_test"
		}
		os.Setenv("DB_DATABASE", dbName+"_test")
	}
}

func (tc *TestCase) handleDatabaseRefresh() {
	if tc.refreshDatabase && !tc.databaseRefreshed {
		if err := tc.RefreshDatabase(); err != nil {
			log.Printf("Warning: Failed to refresh database: %v", err)
			log.Printf("Continuing with existing database state...")
		} else {
			tc.databaseRefreshed = true
		}
	}
}

func (tc *TestCase) bootstrapApplication() {
	cfg := bootstrap.DefaultConfig()

	if tc.Config.SetupRoutes != nil {
		cfg.SetupRoutes = tc.Config.SetupRoutes
	}

	if tc.Config.GormConfig != nil {
		cfg.GormConfig = tc.Config.GormConfig
	}

	if tc.Config.FiberConfig != nil {
		cfg.FiberConfig = tc.Config.FiberConfig
	}

	if tc.Config.ConfigPath != "" {
		cfg.ConfigPath = tc.Config.ConfigPath
	}

	cfg.StartBackgroundJobs = false
	cfg.IsConsoleMode = true

	app := bootstrap.NewApp(func(ac *bootstrap.AppConfig) {
		*ac = *cfg
	})

	tc.App = app
	tc.DB = database.Connect
}

func (tc *TestCase) EnableRefreshDatabase() {
	tc.refreshDatabase = true
}

func (tc *TestCase) RefreshDatabase() error {
	cmd := exec.Command("go", "run", "main.go", "console", "db:fresh", "--force")
	cmd.Dir = tc.projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (tc *TestCase) RefreshDatabaseBetweenTests() {
	tc.databaseRefreshed = false
}

func (tc *TestCase) TearDownTest() {
	if tc.DB != nil {
		sqlDB, err := tc.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

func (tc *TestCase) GetDB() *gorm.DB {
	if tc.DB == nil {
		tc.DB = database.Connect
	}
	return tc.DB
}

func (tc *TestCase) GetApp() *fiber.App {
	return tc.App
}

func (tc *TestCase) GetProjectRoot() string {
	return tc.projectRoot
}
