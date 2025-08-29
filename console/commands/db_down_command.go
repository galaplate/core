package commands

import "slices"

type DbDownCommand struct {
	BaseCommand
}

func (c *DbDownCommand) GetSignature() string {
	return "db:down"
}

func (c *DbDownCommand) GetDescription() string {
	return "Rollback the last database migration"
}

func (c *DbDownCommand) Execute(args []string) error {
	if err := c.CheckDbmate(); err != nil {
		c.PrintError(err.Error())
		c.PrintInfo("Install with: go install github.com/amacneil/dbmate@latest")
		return err
	}

	c.PrintWarning("This will rollback the last migration")

	skipConfirmation := slices.Contains(args, "--force")

	if !skipConfirmation {
		confirmed := c.AskConfirmation("Are you sure you want to rollback?", false)
		if !confirmed {
			c.PrintInfo("Rollback cancelled")
			return nil
		}
	}

	c.PrintInfo("Rolling back last migration...")

	if err := c.RunDbmate("down"); err != nil {
		c.PrintError("Rollback failed")
		return err
	}

	c.PrintSuccess("Rollback completed successfully")
	return nil
}
