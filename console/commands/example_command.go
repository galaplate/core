package commands

import (
	"fmt"

	"github.com/galaplate/core/database"
	"github.com/galaplate/core/logger"
)

// ExampleCommand - Shows how to create a custom command
type ExampleCommand struct{}

func (c *ExampleCommand) GetSignature() string {
	return "example:demo"
}

func (c *ExampleCommand) GetDescription() string {
	return "Example command showing basic usage patterns"
}

func (c *ExampleCommand) Execute(args []string) error {
	logger := logger.NewLogRequestWithUUID(logger.WithField("console", "ExampleCommand@Execute"), "console-command")

	fmt.Println("🚀 Running example command...")

	// Example: Working with arguments
	if len(args) > 0 {
		fmt.Printf("📝 First argument provided: %s\n", args[0])
	} else {
		fmt.Println("📝 No arguments provided")
	}

	// Example: Database query
	var userCount int64
	if err := database.Connect.Table("users").Count(&userCount).Error; err != nil {
		logger.Logger.Error(map[string]any{"error": err.Error(), "action": "count_users_failed"})
		return fmt.Errorf("failed to count users: %v", err)
	}

	fmt.Printf("👥 Total users in database: %d\n", userCount)

	// Example: Structured logging
	logger.Logger.Info(map[string]any{
		"action":     "command_executed",
		"command":    "example:demo",
		"user_count": userCount,
		"args_count": len(args),
	})

	fmt.Println("✅ Example command completed successfully!")

	return nil
}
