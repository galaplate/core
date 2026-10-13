package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/galaplate/core/config"
	"github.com/galaplate/core/supports"
	"github.com/galaplate/core/timezone"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/stdlib"
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

	// GORM defaults NowFunc to time.Now().Local() and never reads the app
	// config, so CreatedAt/UpdatedAt would follow the OS clock. Inject the
	// app-timezone clock when the app did not supply its own.
	if cfg.GormConfig == nil {
		cfg.GormConfig = &gorm.Config{}
	}
	if cfg.GormConfig.NowFunc == nil {
		cfg.GormConfig.NowFunc = timezone.Now
	}

	loc := timezone.Location()

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

		var pgCfg *pgx.ConnConfig
		pgCfg, err = pgx.ParseConfig(dsn)
		if err != nil {
			break
		}

		db, err = gorm.Open(
			postgres.New(postgres.Config{Conn: openPostgres(pgCfg, loc)}),
			cfg.GormConfig,
		)

	case "mysql":
		// loc pins the driver's time.Location for DATETIME parsing; it must be
		// URL-escaped (Asia/Jakarta -> Asia%2FJakarta).
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=%s",
			username, password, host, port, database, url.QueryEscape(loc.String()),
		)

		db, err = gorm.Open(
			mysql.Open(dsn),
			cfg.GormConfig,
		)

	case "sqlite":
		// SQLite has no timezone-aware decoding: the convention is UTC in /
		// UTC out. Values are stored verbatim and read back as written.
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

// openPostgres opens a pgx-backed *sql.DB pinned to loc:
//   - session TimeZone=loc, so server-side NOW()/::date casts use app tz;
//   - TimestampCodec{ScanLocation: loc} for `timestamp WITHOUT time zone`;
//   - TimestamptzCodec{ScanLocation: loc} for `timestamptz`, otherwise scans
//     land in process-local and RFC3339 formatting emits the OS offset.
//
// ScanLocation on timestamptz changes the label only, never the instant.
// The pool is built here rather than via a `TimeZone=` DSN param because the
// gorm driver then appends its own OptionAfterConnect, which overwrites ours
// (OptionAfterConnect sets a single field) and drops the timestamptz codec.
func openPostgres(pgCfg *pgx.ConnConfig, loc *time.Location) *sql.DB {
	pgCfg.RuntimeParams["timezone"] = loc.String()

	return stdlib.OpenDB(*pgCfg, stdlib.OptionAfterConnect(func(_ context.Context, conn *pgx.Conn) error {
		conn.TypeMap().RegisterType(&pgtype.Type{
			Name:  "timestamptz",
			OID:   pgtype.TimestamptzOID,
			Codec: &pgtype.TimestamptzCodec{ScanLocation: loc},
		})
		conn.TypeMap().RegisterType(&pgtype.Type{
			Name:  "timestamp",
			OID:   pgtype.TimestampOID,
			Codec: &pgtype.TimestampCodec{ScanLocation: loc},
		})
		return nil
	}))
}

func GetDriver(dbConn string) string {
	return config.ConfigString(fmt.Sprintf("database.connections.%s.driver", dbConn))
}
