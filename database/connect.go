package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	config "github.com/galaplate/core/env"
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
	var dbType, host, port, username, password, database string

	dbType = supports.MapPostgres(config.Get("DB_CONNECTION"))
	host = config.Get("DB_HOST")
	port = config.Get("DB_PORT")
	username = config.Get("DB_USERNAME")
	password = config.Get("DB_PASSWORD")
	database = config.Get("DB_DATABASE")

	switch dbType {
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
		log.Panic("Unsupported database type", dbType)
	}

	if err != nil {
		log.Panic(err.Error())
		fmt.Println("Failed to connect to the database")
	}
	Connect = db
}
