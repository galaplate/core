package database

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/galaplate/core/config"
	"github.com/galaplate/core/supports"
	"gorm.io/gorm"
)

// Migrator handles running migrations
type Migrator struct {
	db       *gorm.DB
	schema   *Schema
	registry *MigrationRegistry
}

// NewMigrator creates a new migrator instance
func NewMigrator() *Migrator {
	return &Migrator{
		db:       Connect,
		schema:   NewSchema(),
		registry: DefaultRegistry,
	}
}

// disableForeignKeyChecks disables foreign key checks for the current database
func (m *Migrator) disableForeignKeyChecks() error {
	dbType := supports.MapPostgres(GetDriver(config.ConfigString("database.default")))

	switch dbType {
	case "mysql":
		return m.db.Exec("SET FOREIGN_KEY_CHECKS=0;").Error
	case "postgres":
		// PostgreSQL uses session-level setting
		return nil
	case "sqlite":
		return m.db.Exec("PRAGMA foreign_keys = OFF;").Error
	default:
		// Unknown database type, skip
		return nil
	}
}

// enableForeignKeyChecks enables foreign key checks for the current database
func (m *Migrator) enableForeignKeyChecks() error {
	dbType := supports.MapPostgres(GetDriver(config.ConfigString("database.default")))

	switch dbType {
	case "mysql":
		return m.db.Exec("SET FOREIGN_KEY_CHECKS=1;").Error
	case "postgres":
		// PostgreSQL uses session-level setting
		return nil
	case "sqlite":
		return m.db.Exec("PRAGMA foreign_keys = ON;").Error
	default:
		// Unknown database type, skip
		return nil
	}
}

// CreateMigrationsTable creates the migrations table if it doesn't exist
func (m *Migrator) CreateMigrationsTable() error {
	if m.schema.HasTable("migrations") {
		return nil
	}

	return m.schema.Create("migrations", func(table *Blueprint) {
		table.ID()
		table.String("migration", 255).NotNullable()
		table.Integer("batch").NotNullable()
		table.Timestamp("created_at").NotNullable().Default("CURRENT_TIMESTAMP")
	})
}

// GetRanMigrations returns migrations that have been run
func (m *Migrator) GetRanMigrations() ([]string, error) {
	if err := m.CreateMigrationsTable(); err != nil {
		return nil, err
	}

	var migrations []MigrationInfo
	err := m.db.Table("migrations").Order("batch ASC, migration ASC").Find(&migrations).Error
	if err != nil {
		return nil, err
	}

	var names []string
	for _, migration := range migrations {
		names = append(names, migration.Migration)
	}

	return names, nil
}

// GetPendingMigrations returns migrations that haven't been run
func (m *Migrator) GetPendingMigrations() ([]Migration, error) {
	ran, err := m.GetRanMigrations()
	if err != nil {
		return nil, err
	}

	ranMap := make(map[string]bool)
	for _, name := range ran {
		ranMap[name] = true
	}

	var pending []Migration
	allMigrations := m.registry.GetMigrations()

	for _, migration := range allMigrations {
		if !ranMap[migration.GetName()] {
			pending = append(pending, migration)
		}
	}

	// Sort by timestamp
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].GetTimestamp() < pending[j].GetTimestamp()
	})

	return pending, nil
}

// GetLastBatch returns the last batch number
func (m *Migrator) GetLastBatch() (int, error) {
	if err := m.CreateMigrationsTable(); err != nil {
		return 0, err
	}

	var batch int
	err := m.db.Table("migrations").Select("COALESCE(MAX(batch), 0) as batch").Scan(&batch).Error
	return batch, err
}

// GetMigrationsForRollback returns migrations from the last batch for rollback
func (m *Migrator) GetMigrationsForRollback() ([]Migration, error) {
	lastBatch, err := m.GetLastBatch()
	if err != nil {
		return nil, err
	}

	if lastBatch == 0 {
		return []Migration{}, nil
	}

	var migrationNames []string
	err = m.db.Table("migrations").
		Where("batch = ?", lastBatch).
		Order("migration DESC").
		Pluck("migration", &migrationNames).Error

	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, name := range migrationNames {
		migration := m.registry.GetMigrationByName(name)
		if migration != nil {
			migrations = append(migrations, migration)
		}
	}

	return migrations, nil
}

// Up runs all pending migrations
func (m *Migrator) Up() error {
	if err := m.CreateMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	pending, err := m.GetPendingMigrations()
	if err != nil {
		return fmt.Errorf("failed to get pending migrations: %w", err)
	}

	if len(pending) == 0 {
		fmt.Println("Nothing to migrate")
		return nil
	}

	lastBatch, err := m.GetLastBatch()
	if err != nil {
		return fmt.Errorf("failed to get last batch: %w", err)
	}

	newBatch := lastBatch + 1

	for _, migration := range pending {
		fmt.Printf("Migrating: %s", migration.GetFileName())

		// Start transaction
		tx := m.db.Begin()

		// Create schema with transaction
		txSchema := &Schema{db: tx, dbType: m.schema.dbType}

		// Run the migration
		if err := migration.Up(txSchema); err != nil {
			tx.Rollback()
			fmt.Printf(" ❌\n")
			return fmt.Errorf("migration %s failed: %w", migration.GetName(), err)
		}

		// Record the migration
		migrationRecord := MigrationInfo{
			Migration: migration.GetName(),
			Batch:     newBatch,
			CreatedAt: time.Now(),
		}

		if err := tx.Table("migrations").Create(&migrationRecord).Error; err != nil {
			tx.Rollback()
			fmt.Printf(" ❌\n")
			return fmt.Errorf("failed to record migration %s: %w", migration.GetName(), err)
		}

		tx.Commit()
		fmt.Printf(" DONE\n")
	}
	return nil
}

// Down rolls back the last batch of migrations
func (m *Migrator) Down() error {
	if err := m.CreateMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrations, err := m.GetMigrationsForRollback()
	if err != nil {
		return fmt.Errorf("failed to get rollback migrations: %w", err)
	}

	if len(migrations) == 0 {
		fmt.Println("Nothing to rollback")
		return nil
	}

	for _, migration := range migrations {
		fmt.Printf("Rolling back: %s", migration.GetFileName())

		// Start transaction
		tx := m.db.Begin()

		// Create schema with transaction
		txSchema := &Schema{db: tx, dbType: m.schema.dbType}

		// Run the rollback
		if err := migration.Down(txSchema); err != nil {
			tx.Rollback()
			fmt.Printf(" ❌\n")
			return fmt.Errorf("rollback of %s failed: %w", migration.GetName(), err)
		}

		// Remove the migration record
		if err := tx.Table("migrations").Where("migration = ?", migration.GetName()).Delete(&MigrationInfo{}).Error; err != nil {
			tx.Rollback()
			fmt.Printf(" ❌\n")
			return fmt.Errorf("failed to remove migration record %s: %w", migration.GetName(), err)
		}

		tx.Commit()
		fmt.Printf(" DONE\n")
	}
	return nil
}

// Status shows the migration status
func (m *Migrator) Status() error {
	if err := m.CreateMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	ran, err := m.GetRanMigrations()
	if err != nil {
		return fmt.Errorf("failed to get ran migrations: %w", err)
	}

	pending, err := m.GetPendingMigrations()
	if err != nil {
		return fmt.Errorf("failed to get pending migrations: %w", err)
	}

	ranMap := make(map[string]bool)
	for _, name := range ran {
		ranMap[name] = true
	}

	allMigrations := m.registry.GetMigrations()

	fmt.Printf("%-50s %s\n", "Migration", "Status")
	fmt.Printf("%-50s %s\n", strings.Repeat("-", 50), strings.Repeat("-", 10))

	for _, migration := range allMigrations {
		status := "Pending"
		if ranMap[migration.GetName()] {
			status = "Ran"
		}
		fmt.Printf("%-50s %s\n", migration.GetName(), status)
	}

	fmt.Printf("\nTotal migrations: %d\n", len(allMigrations))
	fmt.Printf("Ran: %d\n", len(ran))
	fmt.Printf("Pending: %d\n", len(pending))

	return nil
}

// Reset rolls back all migrations
func (m *Migrator) Reset() error {
	if err := m.CreateMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Disable foreign key checks
	fmt.Println("Disabling foreign key checks...")
	if err := m.disableForeignKeyChecks(); err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	// Ensure foreign key checks are re-enabled even if there's an error
	defer func() {
		fmt.Println("Re-enabling foreign key checks...")
		if err := m.enableForeignKeyChecks(); err != nil {
			fmt.Printf("Warning: failed to re-enable foreign key checks: %v\n", err)
		}
	}()

	// Rollback all batches
	if err := m.resetMigrations(); err != nil {
		return err
	}

	return nil
}

// resetMigrations is the internal implementation without FK handling
func (m *Migrator) resetMigrations() error {
	for {
		migrations, err := m.GetMigrationsForRollback()
		if err != nil {
			return err
		}

		if len(migrations) == 0 {
			break
		}

		if err := m.Down(); err != nil {
			return err
		}
	}

	return nil
}

// Refresh rolls back all migrations and runs them again
func (m *Migrator) Refresh() error {
	// Disable foreign key checks
	fmt.Println("Disabling foreign key checks...")
	if err := m.disableForeignKeyChecks(); err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	// Ensure foreign key checks are re-enabled even if there's an error
	defer func() {
		fmt.Println("Re-enabling foreign key checks...")
		if err := m.enableForeignKeyChecks(); err != nil {
			fmt.Printf("Warning: failed to re-enable foreign key checks: %v\n", err)
		}
	}()

	if err := m.CreateMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Reset all migrations (internal version without FK handling)
	if err := m.resetMigrations(); err != nil {
		return err
	}

	// Run all migrations
	return m.Up()
}
