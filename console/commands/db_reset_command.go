package commands

import "slices"

type DbResetCommand struct {
	BaseCommand
}

func (c *DbResetCommand) GetSignature() string {
	return "db:reset"
}

func (c *DbResetCommand) GetDescription() string {
	return "Drop and recreate the database"
}

func (c *DbResetCommand) Execute(args []string) error {
	if err := c.CheckDbmate(); err != nil {
		c.PrintError(err.Error())
		c.PrintInfo("Install with: go install github.com/amacneil/dbmate@latest")
		return err
	}

	c.PrintWarning("⚠️  DANGER: This will drop and recreate the entire database!")
	c.PrintWarning("All data will be permanently lost!")

	skipConfirmation := slices.Contains(args, "--force")

	if !skipConfirmation {
		response := c.AskText("Type 'yes' to confirm database reset", "")
		if response != "yes" {
			c.PrintInfo("Database reset cancelled")
			return nil
		}
	}

	c.PrintInfo("Dropping database...")
	if err := c.RunDbmate("drop"); err != nil {
		c.PrintError("Failed to drop database")
		return err
	}

	c.PrintInfo("Creating database...")
	if err := c.RunDbmate("create"); err != nil {
		c.PrintError("Failed to create database")
		return err
	}

	c.PrintSuccess("Database reset completed successfully")
	return nil
}
