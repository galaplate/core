package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type DbCreateCommand struct {
	BaseCommand
}

func (c *DbCreateCommand) GetSignature() string {
	return "db:create"
}

func (c *DbCreateCommand) GetDescription() string {
	return "Create a new Go-based migration file"
}

func (c *DbCreateCommand) Execute(args []string) error {
	var migrationName string

	if len(args) == 0 {
		migrationName = c.AskRequired("Enter migration name (e.g., create_users_table)")
	} else {
		migrationName = args[0]
	}

	if migrationName == "" {
		return fmt.Errorf("migration name cannot be empty")
	}

	return c.createMigration(migrationName)
}

func (c *DbCreateCommand) createMigration(name string) error {
	migrationsDir := "db/migrations"

	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %v", err)
	}

	timestamp := time.Now().Unix()
	fileName := fmt.Sprintf("%d_%s.go", timestamp, name)
	filePath := filepath.Join(migrationsDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("migration file %s already exists", filePath)
	}

	moduleName, err := c.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %v", err)
	}

	tableName := extractTableName(name)

	content := fmt.Sprintf(`package migrations

import (
    "github.com/galaplate/core/database"
)

type Migration%d struct {
    database.BaseMigration
}

func init() {
    migration := &Migration%d{
        BaseMigration: database.BaseMigration{
            Name:      "%s",
            Timestamp: %d,
        },
    }
    database.Register(migration)
}

func (m *Migration%d) Up(schema *database.Schema) error {
    return schema.Create("%s", func(table *database.Blueprint) {
        table.ID()
        // Add your columns here
        // table.String("name")
        // table.String("email").Unique()
        // table.Boolean("is_active").Default(true)
        // table.Timestamps()
    })
}

func (m *Migration%d) Down(schema *database.Schema) error {
    return schema.DropIfExists("%s")
}
`, timestamp, timestamp, name, timestamp, timestamp, tableName, timestamp, tableName)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write migration file: %v", err)
	}

	c.PrintSuccess(fmt.Sprintf("Migration created: %s", filePath))
	c.PrintInfo("Remember to import this migration in your main.go")
	c.PrintInfo(fmt.Sprintf("Example: _ \"%s/db/migrations\"", moduleName))

	return nil
}

func extractTableName(name string) string {
	// Extract table name from migration name
	// e.g., "create_users_table" -> "users"
	if len(name) > 7 && name[:7] == "create_" {
		tableName := name[7:]
		if len(tableName) > 6 && tableName[len(tableName)-6:] == "_table" {
			return tableName[:len(tableName)-6]
		}
		return tableName
	}
	return "example_table"
}
