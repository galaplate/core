package commands

import (
	"slices"

	"github.com/galaplate/core/database"
)

type DbDownCommand struct {
	BaseCommand
}

func (c *DbDownCommand) GetSignature() string {
	return "db:down"
}

func (c *DbDownCommand) GetDescription() string {
	return "Rollback the last database migration batch"
}

func (c *DbDownCommand) Execute(args []string) error {
	// Initialize database connection
	database.New()

	skipConfirmation := slices.Contains(args, "--force")

	if !skipConfirmation {
		c.PrintWarning("This will rollback the last migration batch")
		confirmed := c.AskConfirmation("Are you sure you want to rollback?", false)
		if !confirmed {
			c.PrintInfo("Rollback cancelled")
			return nil
		}
	}

	migrator := database.NewMigrator()

	if err := migrator.Down(); err != nil {
		return err
	}

	return nil
}
