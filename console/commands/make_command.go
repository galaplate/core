package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type MakeCommand struct {
	BaseCommand
}

func (c *MakeCommand) GetSignature() string {
	return "make:command"
}

func (c *MakeCommand) GetDescription() string {
	return "Create a new console command"
}

func (c *MakeCommand) Execute(args []string) error {
	var commandName string

	if len(args) == 0 {
		commandName = c.askForCommandName()
	} else {
		commandName = args[0]
	}

	if commandName == "" {
		return fmt.Errorf("command name cannot be empty")
	}

	return c.createCommand(commandName)
}

func (c *MakeCommand) askForCommandName() string {
	return c.AskRequired("Enter command name (e.g., SendWelcomeEmail, GenerateReport)")
}

func (c *MakeCommand) createCommand(name string) error {
	commandsDir := "console/commands"

	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return fmt.Errorf("failed to create commands directory: %v", err)
	}

	fileName := strings.ToLower(name) + "_command.go"
	filePath := filepath.Join(commandsDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("command file %s already exists", filePath)
	}

	moduleName, err := c.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %v", err)
	}

	className := formatClassName(name)
	signatureName := strings.ToLower(name)

	// Use internal stub template
	if err := c.GenerateFromStub("commands/command.go.stub", filePath, CommandTemplate{
		ClassName:   className,
		Signature:   signatureName,
		Description: fmt.Sprintf("Command description for %s", name),
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
		ModuleName:  moduleName,
	}); err != nil {
        panic(err)
	}

	fmt.Printf("✅ Command created successfully: %s\n", filePath)
	fmt.Printf("📝 Remember to register your command in console/commands.go\n")
	fmt.Printf("\nExample registration:\n")
	fmt.Printf("kernel.Register(&commands.%s{})\n", className)
	fmt.Printf("In console/kernel.go RegisterCommands function\n")

	return nil
}

func formatClassName(name string) string {
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, "-", "")

	if len(name) == 0 {
		return "Command"
	}

	return strings.ToUpper(name[:1]) + strings.ToLower(name[1:]) + "Command"
}

type CommandTemplate struct {
	ClassName   string
	Signature   string
	Description string
	Timestamp   string
	ModuleName  string
}
