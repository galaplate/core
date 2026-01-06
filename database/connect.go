package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/galaplate/core/config"
	"github.com/galaplate/core/supports"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Connect *gorm.DB

type Config struct {
	GormConfig *gorm.Config
}

func ConnectDB() {
	// ConnectWithConfig(nil)
}

type OptFunc func(*Config)

func New(opts ...OptFunc) {
	cfg := DefaultGormConfig()

	// Apply all provided options
	for _, opt := range opts {
		opt(cfg)
	}

	ConnectWithConfig(cfg)
}

// DefaultGormConfig returns default GORM configuration
func DefaultGormConfig() *Config {
	var logLevel logger.LogLevel
	logLevelStr := "warn" // default

	switch strings.ToLower(logLevelStr) {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:
		logLevel = logger.Warn
	}

	return &Config{
		GormConfig: &gorm.Config{
			Logger: logger.New(
				log.New(os.Stdout, "\r\n", log.LstdFlags),
				logger.Config{
					SlowThreshold:             time.Second,
					LogLevel:                  logLevel,
					IgnoreRecordNotFoundError: true,
					ParameterizedQueries:      true,
					Colorful:                  true,
				},
			),
			DisableForeignKeyConstraintWhenMigrating: true,
		},
	}
}
func ConnectWithConfig(cfg *Config) {
	var err error
	var db *gorm.DB

	// Use provided config or fall back to environment variables
	var dbConn, host, port, username, password, database, driver string

	dbConn = config.ConfigString("database.default")

	host = config.ConfigString(fmt.Sprintf("database.connections.%s.host", dbConn))
	port = config.ConfigString(fmt.Sprintf("database.connections.%s.port", dbConn))
	username = config.ConfigString(fmt.Sprintf("database.connections.%s.username", dbConn))
	password = config.ConfigString(fmt.Sprintf("database.connections.%s.password", dbConn))
	database = config.ConfigString(fmt.Sprintf("database.connections.%s.database", dbConn))
	driver = supports.MapPostgres(GetDriver(dbConn))

	switch driver {
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, username, password, database,
		)

		db, err = gorm.Open(
			postgres.Open(dsn),
			cfg.GormConfig,
		)

	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			username, password, host, port, database,
		)

		db, err = gorm.Open(
			mysql.Open(dsn),
			cfg.GormConfig,
		)

	case "sqlite":
		dsn := database
		if dsn == "" {
			dsn = "db/database.sqlite"
		}

		db, err = gorm.Open(
			sqlite.Open(dsn),
			cfg.GormConfig,
		)
	default:
		log.Panic("Unsupported database type", driver)
	}

	if err != nil {
		log.Panic(err.Error())
		fmt.Println("Failed to connect to the database")
	}

	// Apply connection pool settings from config
	sqlDB, err := db.DB()
	if err == nil {
		poolSize := config.ConfigInt(fmt.Sprintf("database.connections.%s.pool_size", dbConn))
		maxIdle := config.ConfigInt(fmt.Sprintf("database.connections.%s.max_idle_connections", dbConn))

		if poolSize > 0 {
			sqlDB.SetMaxOpenConns(poolSize)
		}
		if maxIdle > 0 {
			sqlDB.SetMaxIdleConns(maxIdle)
		}
	}

	Connect = db
}

func GetDriver(dbConn string) string {
	return config.ConfigString(fmt.Sprintf("database.connections.%s.driver", dbConn))
}
