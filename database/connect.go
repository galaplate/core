package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/galaplate/core/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Connect *gorm.DB

func ConnectDB() {
	var err error
	var dsn string
	var dbType = config.Get("DB_CONNECTION")
	var db *gorm.DB

	var gormConfig = &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				ParameterizedQueries:      true,
				Colorful:                  true,
			},
		),
	}

	switch dbType {
	case "postgres":
		p := config.Get("DB_PORT")

		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			config.Get("DB_HOST"),
			p,
			config.Get("DB_USERNAME"),
			config.Get("DB_PASSWORD"),
			config.Get("DB_DATABASE"),
		)

		db, err = gorm.Open(
			postgres.Open(dsn),
			gormConfig,
		)

	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Get("DB_USERNAME"),
			config.Get("DB_PASSWORD"),
			config.Get("DB_HOST"),
			config.Get("DB_PORT"),
			config.Get("DB_DATABASE"),
		)

		db, err = gorm.Open(
			mysql.Open(dsn),
			gormConfig,
		)
	default:
		log.Panic("Unsupported database type")
	}

	if err != nil {
		log.Panic(err.Error())
		fmt.Println("Failed to connect to the database")
	}
	Connect = db
}
