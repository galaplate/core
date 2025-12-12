package database

import (
	"fmt"
	"time"
)

// Migration interface
type Migration interface {
	Up(schema *Schema) error
	Down(schema *Schema) error
	GetName() string
	GetTimestamp() int64
	GetFileName() string
}

// BaseMigration provides default implementation
type BaseMigration struct {
	Name      string
	Timestamp int64
}

func (m *BaseMigration) GetName() string {
	return m.Name
}

func (m *BaseMigration) GetTimestamp() int64 {
	return m.Timestamp
}

func (m *BaseMigration) GetFileName() string {
	return fmt.Sprintf("%d_%s", m.Timestamp, m.Name)
}

// MigrationInfo holds metadata about a migration
type MigrationInfo struct {
	ID        int64     `json:"id"`
	Migration string    `json:"migration"`
	Batch     int       `json:"batch"`
	CreatedAt time.Time `json:"created_at"`
}

// MigrationRegistry holds all registered migrations
type MigrationRegistry struct {
	migrations []Migration
}

var DefaultRegistry = &MigrationRegistry{}

// Register adds a migration to the registry
func (r *MigrationRegistry) Register(migration Migration) {
	r.migrations = append(r.migrations, migration)
}

// GetMigrations returns all registered migrations sorted by timestamp
func (r *MigrationRegistry) GetMigrations() []Migration {
	// Sort migrations by timestamp
	for i := 0; i < len(r.migrations); i++ {
		for j := i + 1; j < len(r.migrations); j++ {
			if r.migrations[i].GetTimestamp() > r.migrations[j].GetTimestamp() {
				r.migrations[i], r.migrations[j] = r.migrations[j], r.migrations[i]
			}
		}
	}
	return r.migrations
}

// GetMigrationByName finds a migration by name
func (r *MigrationRegistry) GetMigrationByName(name string) Migration {
	for _, migration := range r.migrations {
		if migration.GetName() == name {
			return migration
		}
	}
	return nil
}

// Register is a helper function to register migrations
func Register(migration Migration) {
	DefaultRegistry.Register(migration)
}

// CreateMigrationTemplate generates a Go migration template
func CreateMigrationTemplate(name string) string {
	timestamp := time.Now().Unix()

	template := fmt.Sprintf(`package migrations

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
`, timestamp, timestamp, name, timestamp, timestamp, extractTableName(name), timestamp, extractTableName(name))

	return template
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
