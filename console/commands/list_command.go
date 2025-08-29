package commands

import (
	"fmt"
	"sort"
)

type ListCommand struct {
	Commands map[string]interface{}
}

func (c *ListCommand) GetSignature() string {
	return "list"
}

func (c *ListCommand) GetDescription() string {
	return "List all available console commands"
}

func (c *ListCommand) Execute(args []string) error {
	fmt.Println("📋 Available Console Commands:")
	fmt.Println()

	if len(c.Commands) == 0 {
		fmt.Println("No commands registered.")
		return nil
	}

	// Get sorted command signatures
	var signatures []string
	for signature := range c.Commands {
		signatures = append(signatures, signature)
	}
	sort.Strings(signatures)

	// Display commands in alphabetical order
	for _, signature := range signatures {
		if command, ok := c.Commands[signature].(Command); ok {
			fmt.Printf("  %-20s %s\n", signature, command.GetDescription())
		}
	}

	fmt.Println()
	fmt.Println("💡 Usage: go run main.go console <command> [arguments]")

	return nil
}

type Command interface {
	GetSignature() string
	GetDescription() string
	Execute(args []string) error
}
