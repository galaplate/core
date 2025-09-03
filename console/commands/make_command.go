package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
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

	templateData := CommandTemplate{
		ClassName:   className,
		Signature:   signatureName,
		Description: fmt.Sprintf("Command description for %s", name),
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
		ModuleName:  moduleName,
	}

	tmpl, err := template.New("command").Parse(commandTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to write template: %v", err)
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

const commandTemplate = `package commands

import (
	"fmt"
	// "{{.ModuleName}}/db"
    "github.com/galaplate/core/logger"
)

// {{.ClassName}} - Generated on {{.Timestamp}}
// Add your custom command logic here
type {{.ClassName}} struct{}

func (c *{{.ClassName}}) GetSignature() string {
	return "{{.Signature}}"
}

func (c *{{.ClassName}}) GetDescription() string {
	return "{{.Description}}"
}

func (c *{{.ClassName}}) Execute(args []string) error {
	logger := logger.NewLogRequestWithUUID(logger.WithField("console", "{{.ClassName}}@Execute"), "console-command")

	fmt.Println("🚀 Executing {{.Signature}} command...")

	// Database connection is available via database.Connect
	// Logger is available for structured logging

	// TODO: Add your command logic here

	// Example database query:
	// var count int64
	// if err := database.Connect.Table("users").Count(&count).Error; err != nil {
	//     logger.Logger.Error(map[string]any{"error": err.Error(), "action": "count_users_failed"})
	//     return fmt.Errorf("failed to count users: %v", err)
	// }

	// Example with arguments:
	// if len(args) > 0 {
	//     fmt.Printf("First argument: %s\n", args[0])
	// }

	fmt.Println("✅ {{.Signature}} command completed successfully")
	logger.Logger.Info(map[string]any{"action": "command_completed", "command": "{{.Signature}}"})

	return nil
}
`
