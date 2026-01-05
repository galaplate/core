package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/galaplate/core/config"
	"github.com/galaplate/core/supports"
)

type DbDumpCommand struct {
	BaseCommand
}

func (c *DbDumpCommand) GetSignature() string {
	return "db:dump"
}

func (c *DbDumpCommand) GetDescription() string {
	return "Dump database to SQL file"
}

func (c *DbDumpCommand) Execute(args []string) error {
	if err := c.LoadEnvVariables(); err != nil {
		c.PrintError(fmt.Sprintf("Failed to load environment variables: %v", err))
		return err
	}

	dbConnection := supports.MapPostgres(config.ConfigString("database.default"))
	dbHost := config.ConfigString(fmt.Sprintf("database.connections.%s.host", dbConnection))
	dbPort := config.ConfigString(fmt.Sprintf("database.connections.%s.port", dbConnection))
	dbDatabase := config.ConfigString(fmt.Sprintf("database.connections.%s.database", dbConnection))
	dbUsername := config.ConfigString(fmt.Sprintf("database.connections.%s.username", dbConnection))
	dbPassword := config.ConfigString(fmt.Sprintf("database.connections.%s.password", dbConnection))

	if dbConnection == "" {
		c.PrintError("Missing DB_CONNECTION in .env file")
		return fmt.Errorf("missing database connection type")
	}

	// SQLite validation
	if dbConnection == "sqlite" {
		if dbDatabase == "" {
			dbDatabase = "database.sqlite"
		}
	} else {
		// MySQL/PostgreSQL validation
		if dbHost == "" || dbPort == "" || dbDatabase == "" {
			c.PrintError("Missing database configuration in .env file")
			return fmt.Errorf("missing database configuration")
		}
	}

	// Create dumps directory if it doesn't exist
	dumpsDir := "./db/dumps"
	if err := os.MkdirAll(dumpsDir, 0755); err != nil {
		c.PrintError(fmt.Sprintf("Failed to create dumps directory: %v", err))
		return err
	}

	// Generate filename with timestamp
	var filename string
	if len(args) > 0 {
		filename = args[0]
		if !filepath.IsAbs(filename) && filepath.Ext(filename) != ".sql" {
			filename = filepath.Join(dumpsDir, filename+".sql")
		}
	} else {
		timestamp := time.Now().Format("20060102_150405")
		filename = filepath.Join(dumpsDir, fmt.Sprintf("%s_%s.sql", dbDatabase, timestamp))
	}

	c.PrintInfo(fmt.Sprintf("Dumping database '%s' to %s...", dbDatabase, filename))

	var cmd *exec.Cmd
	switch dbConnection {
	case "sqlite":
		if err := c.checkCommand("sqlite3"); err != nil {
			return err
		}
		cmd = exec.Command("sqlite3", dbDatabase, ".dump")
	case "mysql":
		if err := c.checkCommand("mysqldump"); err != nil {
			return err
		}
		cmd = exec.Command("mysqldump",
			fmt.Sprintf("--host=%s", dbHost),
			fmt.Sprintf("--port=%s", dbPort),
			fmt.Sprintf("--user=%s", dbUsername),
			fmt.Sprintf("--password=%s", dbPassword),
			"--single-transaction",
			"--routines",
			"--triggers",
			dbDatabase,
		)
	case "postgres":
		if err := c.checkCommand("pg_dump"); err != nil {
			return err
		}
		os.Setenv("PGPASSWORD", dbPassword)
		cmd = exec.Command("pg_dump",
			fmt.Sprintf("--host=%s", dbHost),
			fmt.Sprintf("--port=%s", dbPort),
			fmt.Sprintf("--username=%s", dbUsername),
			"--no-password",
			"--verbose",
			"--clean",
			"--no-acl",
			"--no-owner",
			dbDatabase,
		)
	default:
		c.PrintError(fmt.Sprintf("Unsupported database driver: %s (supported: sqlite, mysql, postgres)", dbConnection))
		return fmt.Errorf("unsupported database driver: %s", dbConnection)
	}

	// Create output file
	outFile, err := os.Create(filename)
	if err != nil {
		c.PrintError(fmt.Sprintf("Failed to create dump file: %v", err))
		return err
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		c.PrintError(fmt.Sprintf("Database dump failed: %v", err))
		// Clean up failed dump file
		os.Remove(filename)
		return err
	}

	c.PrintSuccess(fmt.Sprintf("Database dumped successfully to %s", filename))
	return nil
}

func (c *DbDumpCommand) checkCommand(command string) error {
	_, err := exec.LookPath(command)
	if err != nil {
		c.PrintError(fmt.Sprintf("%s command not found", command))
		switch command {
		case "sqlite3":
			c.PrintInfo("Install SQLite3 command-line tool:")
			c.PrintInfo("  macOS: brew install sqlite3")
			c.PrintInfo("  Ubuntu/Debian: sudo apt-get install sqlite3")
			c.PrintInfo("  CentOS/RHEL: sudo yum install sqlite")
		case "mysqldump":
			c.PrintInfo("Install MySQL client tools to use mysqldump")
		case "pg_dump":
			c.PrintInfo("Install PostgreSQL client tools to use pg_dump")
		}
		return fmt.Errorf("%s is not installed", command)
	}
	return nil
}
