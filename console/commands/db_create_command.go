package commands

import (
	"fmt"
)

type DbCreateCommand struct {
	BaseCommand
}

func (c *DbCreateCommand) GetSignature() string {
	return "db:create"
}

func (c *DbCreateCommand) GetDescription() string {
	return "Create a new database migration"
}

func (c *DbCreateCommand) Execute(args []string) error {
	if err := c.CheckDbmate(); err != nil {
		c.PrintError(err.Error())
		c.PrintInfo("Install with: go install github.com/amacneil/dbmate@latest")
		return err
	}

	var migrationName string

	if len(args) == 0 {
		migrationName = c.AskRequired("Enter migration name (e.g., create_users_table)")
	} else {
		migrationName = args[0]
	}

	if migrationName == "" {
		return fmt.Errorf("migration name cannot be empty")
	}

	c.PrintInfo(fmt.Sprintf("Creating migration: %s", migrationName))

	if err := c.RunDbmate("new", migrationName); err != nil {
		c.PrintError(fmt.Sprintf("Failed to create migration: %v", err))
		return err
	}

	c.PrintSuccess("Migration created successfully")
	return nil
}
