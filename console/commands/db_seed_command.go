package commands

import (
	"fmt"

	"github.com/galaplate/core/database"
	"github.com/galaplate/core/database/seeders"
)

type DbSeedCommand struct {
	BaseCommand
}

func (c *DbSeedCommand) GetSignature() string {
	return "db:seed"
}

func (c *DbSeedCommand) GetDescription() string {
	return "Run database seeders"
}

func (c *DbSeedCommand) Execute(args []string) error {
	var seederFile string

	if len(args) == 0 {
		choices := []string{"All seeders", "Specific seeder"}
		choice := c.AskChoice("What would you like to seed?", choices, 0)

		if choice == "Specific seeder" {
			seederFile = c.AskRequired("Enter seeder filename (without .go extension)")
		}
	} else {
		seederFile = args[0]
	}

	c.PrintInfo("Running database seeders...")

	database.New()

	seeder := seeders.NewDatabaseSeeder(seederFile)

	if err := seeder.Run(database.Connect); err != nil {
		c.PrintError(fmt.Sprintf("Seeding failed: %v", err))
		return err
	}

	c.PrintSuccess("Database seeding completed successfully")
	return nil
}
