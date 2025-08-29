package commands

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
	if err := c.CheckDbmate(); err != nil {
		c.PrintError(err.Error())
		c.PrintInfo("Install with: go install github.com/amacneil/dbmate@latest")
		return err
	}

	c.PrintInfo("Running pending migrations...")

	if err := c.RunDbmate("up"); err != nil {
		c.PrintError("Migrations failed")
		return err
	}

	c.PrintSuccess("Migrations completed successfully")
	return nil
}
