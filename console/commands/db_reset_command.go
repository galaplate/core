package commands

import (
	"slices"

	"github.com/galaplate/core/database"
)

type DbResetCommand struct {
	BaseCommand
}

func (c *DbResetCommand) GetSignature() string {
	return "db:reset"
}

func (c *DbResetCommand) GetDescription() string {
	return "Rollback all migrations"
}

func (c *DbResetCommand) Execute(args []string) error {
	// Initialize database connection
	database.New()

	skipConfirmation := slices.Contains(args, "--force")

	if !skipConfirmation {
		c.PrintWarning("⚠️  DANGER: This will rollback all migrations!")
		c.PrintWarning("All data will be permanently lost!")
		response := c.AskText("Type 'yes' to confirm reset", "")
		if response != "yes" {
			c.PrintInfo("Reset cancelled")
			return nil
		}
	}

	migrator := database.NewMigrator()

	if err := migrator.Reset(); err != nil {
		return err
	}

	return nil
}
