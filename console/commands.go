package console

import "github.com/galaplate/core/console/commands"

// RegisterCommands registers all available console commands
// Users can add their custom commands here
func (k *Kernel) RegisterCommands() {
	// Example command (you can remove this)
	k.Register(&commands.ExampleCommand{})

	// Interactive demo command
	k.Register(&commands.InteractiveCommand{})

	// Database commands
	k.Register(&commands.DbCreateCommand{})
	k.Register(&commands.DbUpCommand{})
	k.Register(&commands.DbDownCommand{})
	k.Register(&commands.DbStatusCommand{})
	k.Register(&commands.DbResetCommand{})
	k.Register(&commands.DbFreshCommand{})
	k.Register(&commands.DbSeedCommand{})
	k.Register(&commands.DbDumpCommand{})

	// Make commands
	k.Register(&commands.MakeCommand{})
	k.Register(&commands.MakeModelCommand{})
	k.Register(&commands.MakeSeederCommand{})
	k.Register(&commands.MakeJobCommand{})
	k.Register(&commands.MakeDtoCommand{})
	k.Register(&commands.MakeCronCommand{})

	// Template commands
	k.Register(&commands.TemplateCommand{})
	k.Register(&commands.TemplateGenerateCommand{})

	// Other commands
	k.Register(&commands.ListCommand{})
	k.Register(&commands.PolicyCommand{})

	// Register your custom commands here
	// Example: k.Register(&commands.YourCustomCommand{})
}
