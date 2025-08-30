package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	config "github.com/galaplate/core/env"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Connect *gorm.DB

type GormConfig struct {
	gorm.Config

	// Additional logger-specific configs for easier configuration
	SlowThreshold             time.Duration
	LogLevel                  string // Silent, Error, Warn, Info
	IgnoreRecordNotFoundError bool
	ParameterizedQueries      bool
	Colorful                  bool
}

type Config struct {
	GormConfig *GormConfig
}

func ConnectDB() {
	ConnectWithConfig(nil)
}

func ConnectWithConfig(cfg *Config) {
	var err error
	var dsn string
	var db *gorm.DB

	// Use provided config or fall back to environment variables
	var dbType, host, port, username, password, database string

	dbType = config.Get("DB_CONNECTION")
	host = config.Get("DB_HOST")
	port = config.Get("DB_PORT")
	username = config.Get("DB_USERNAME")
	password = config.Get("DB_PASSWORD")
	database = config.Get("DB_DATABASE")

	// Configure GORM settings
	var gormCfg *GormConfig
	if cfg != nil && cfg.GormConfig != nil {
		gormCfg = cfg.GormConfig
	} else {
		// Default GORM configuration
		gormCfg = &GormConfig{
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

	// Convert log level string to logger level
	var logLevel logger.LogLevel
	logLevelStr := "warn" // default
	if gormCfg.LogLevel != "" {
		logLevelStr = gormCfg.LogLevel
	}

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

	// Start with the embedded gorm.Config
	gormConfig := &gormCfg.Config

	// Set/override the logger if not already set
	if gormConfig.Logger == nil {
		// Set default values if not provided
		slowThreshold := gormCfg.SlowThreshold
		if slowThreshold == 0 {
			slowThreshold = time.Second
		}

		gormConfig.Logger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             slowThreshold,
				LogLevel:                  logLevel,
				IgnoreRecordNotFoundError: gormCfg.IgnoreRecordNotFoundError,
				ParameterizedQueries:      gormCfg.ParameterizedQueries,
				Colorful:                  gormCfg.Colorful,
			},
		)
	}

	switch dbType {
	case "postgres":
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, username, password, database,
		)

		db, err = gorm.Open(
			postgres.Open(dsn),
			gormConfig,
		)

	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			username, password, host, port, database,
		)

		db, err = gorm.Open(
			mysql.Open(dsn),
			gormConfig,
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
