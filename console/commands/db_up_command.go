package commands

import (
	"github.com/galaplate/core/database"
)

type DbUpCommand struct {
	BaseCommand
}

func (c *DbUpCommand) GetSignature() string {
	return "db:up"
}

func (c *DbUpCommand) GetDescription() string {
	return "Run pending database migrations"
}

func (c *DbUpCommand) Execute(args []string) error {
	// Initialize database connection
	database.New()

	migrator := database.NewMigrator()

	if err := migrator.Up(); err != nil {
		return err
	}

	return nil
}
