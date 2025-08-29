package commands

import "slices"

type DbFreshCommand struct {
	BaseCommand
}

func (c *DbFreshCommand) GetSignature() string {
	return "db:fresh"
}

func (c *DbFreshCommand) GetDescription() string {
	return "Drop database, recreate and run all migrations"
}

func (c *DbFreshCommand) Execute(args []string) error {
	if err := c.CheckDbmate(); err != nil {
		c.PrintError(err.Error())
		c.PrintInfo("Install with: go install github.com/amacneil/dbmate@latest")
		return err
	}

	c.PrintWarning("⚠️  DANGER: This will drop the database and run all migrations!")
	c.PrintWarning("All data will be permanently lost!")

	skipConfirmation := slices.Contains(args, "--force")

	if !skipConfirmation {
		response := c.AskText("Type 'yes' to confirm fresh migration", "")
		if response != "yes" {
			c.PrintInfo("Fresh migration cancelled")
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

	// Run all migrations
	c.PrintInfo("Running all migrations...")
	if err := c.RunDbmate("up"); err != nil {
		c.PrintError("Migrations failed")
		return err
	}

	c.PrintSuccess("Fresh migration completed successfully")
	return nil
}
