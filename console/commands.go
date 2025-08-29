package console

import "github.com/galaplate/core/console/commands"

// RegisterCommands registers all available console commands
// Users can add their custom commands here
func (k *Kernel) RegisterCommands() {
	// Example command (you can remove this)
	k.Register(&commands.ExampleCommand{})

	// Interactive demo command
	k.Register(&commands.InteractiveCommand{})

	// Register your custom commands here
	// Example: k.Register(&commands.YourCustomCommand{})
}
