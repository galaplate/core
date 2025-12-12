package commands

import (
	"github.com/galaplate/core/database"
)

type DbStatusCommand struct {
	BaseCommand
}

func (c *DbStatusCommand) GetSignature() string {
	return "db:status"
}

func (c *DbStatusCommand) GetDescription() string {
	return "Show database migration status"
}

func (c *DbStatusCommand) Execute(args []string) error {
	// Initialize database connection
	database.New()

	migrator := database.NewMigrator()

	if err := migrator.Status(); err != nil {
		return err
	}

	return nil
}
