package commands

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
	if err := c.CheckDbmate(); err != nil {
		c.PrintError(err.Error())
		c.PrintInfo("Install with: go install github.com/amacneil/dbmate@latest")
		return err
	}

	c.PrintInfo("Migration status:")

	if err := c.RunDbmate("status"); err != nil {
		c.PrintError("Failed to get migration status")
		return err
	}

	return nil
}
