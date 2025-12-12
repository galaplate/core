package commands

import (
	"slices"

	"github.com/galaplate/core/database"
)

type DbFreshCommand struct {
	BaseCommand
}

func (c *DbFreshCommand) GetSignature() string {
	return "db:fresh"
}

func (c *DbFreshCommand) GetDescription() string {
	return "Rollback all migrations and run them again"
}

func (c *DbFreshCommand) Execute(args []string) error {
	// Initialize database connection
	database.New()

	skipConfirmation := slices.Contains(args, "--force")

	if !skipConfirmation {
		c.PrintWarning("⚠️  DANGER: This will rollback and re-run all migrations!")
		c.PrintWarning("All data will be permanently lost!")
		response := c.AskText("Type 'yes' to confirm fresh migration", "")
		if response != "yes" {
			c.PrintInfo("Fresh migration cancelled")
			return nil
		}
	}

	migrator := database.NewMigrator()

	if err := migrator.Refresh(); err != nil {
		return err
	}

	return nil
}
