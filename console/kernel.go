package console

import (
	"fmt"
	"sort"
	"strings"

	"github.com/galaplate/core/console/commands"
)

type Command interface {
	GetSignature() string
	GetDescription() string
	Execute(args []string) error
}

type Kernel struct {
	commands map[string]Command
}

func NewKernel() *Kernel {
	kernel := &Kernel{
		commands: make(map[string]Command),
	}

	kernel.registerDefaultCommands()
	return kernel
}

func (k *Kernel) registerDefaultCommands() {
	// Core make command
	k.Register(&commands.MakeCommand{})

	// Built-in generator commands
	k.Register(&commands.MakeModelCommand{})
	k.Register(&commands.MakeDtoCommand{})
	k.Register(&commands.MakeJobCommand{})
	k.Register(&commands.MakeCronCommand{})
	k.Register(&commands.MakeSeederCommand{})

	// Database migration commands
	k.Register(&commands.DbCreateCommand{})
	k.Register(&commands.DbUpCommand{})
	k.Register(&commands.DbDownCommand{})
	k.Register(&commands.DbStatusCommand{})
	k.Register(&commands.DbResetCommand{})
	k.Register(&commands.DbFreshCommand{})
	k.Register(&commands.DbSeedCommand{})

	// Register user-defined commands
	k.RegisterCommands()

	// Register ListCommand last so it includes all other commands
	commandsInterface := make(map[string]interface{})
	for signature, command := range k.commands {
		commandsInterface[signature] = command
	}
	k.Register(&commands.ListCommand{Commands: commandsInterface})
}

func (k *Kernel) Register(command Command) {
	k.commands[command.GetSignature()] = command
}

func (k *Kernel) Run(args []string) error {
	if len(args) < 3 {
		return k.showHelp()
	}

	commandName := args[2]
	commandArgs := args[3:]

	if command, exists := k.commands[commandName]; exists {
		return command.Execute(commandArgs)
	}

	fmt.Printf("Command '%s' not found.\n", commandName)
	return k.showHelp()
}

func (k *Kernel) showHelp() error {
	fmt.Println("Galaplate Console Commands")
	fmt.Println("Usage: go run main.go console <command> [arguments]")
	fmt.Println()
	fmt.Println("Available commands:")

	// Get sorted command signatures
	var signatures []string
	for signature := range k.commands {
		signatures = append(signatures, signature)
	}
	sort.Strings(signatures)

	// Display commands in alphabetical order
	for _, signature := range signatures {
		command := k.commands[signature]
		fmt.Printf("  %-20s %s\n", signature, command.GetDescription())
	}

	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run main.go console list")
	fmt.Println("  go run main.go console make:command MyCustomCommand")

	return nil
}

func (k *Kernel) GetCommands() map[string]Command {
	return k.commands
}

func FormatCommandName(name string) string {
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, "-", "")

	if len(name) == 0 {
		return ""
	}

	return strings.ToUpper(name[:1]) + strings.ToLower(name[1:])
}
