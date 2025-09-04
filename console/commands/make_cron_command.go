package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type MakeCronCommand struct {
	BaseCommand
}

func (c *MakeCronCommand) GetSignature() string {
	return "make:cron"
}

func (c *MakeCronCommand) GetDescription() string {
	return "Create a new cron scheduler"
}

func (c *MakeCronCommand) Execute(args []string) error {
	var cronName, schedule string

	if len(args) == 0 {
		cronName, schedule = c.askForCronDetails()
	} else {
		cronName = args[0]
		schedule = "@every 5s"
	}

	if cronName == "" {
		return fmt.Errorf("cron name cannot be empty")
	}

	return c.createCron(cronName, schedule)
}

func (c *MakeCronCommand) askForCronDetails() (string, string) {
	cronName := c.AskRequired("Enter cron name (e.g., DailyCleanup, WeeklyReport)")

	schedules := []string{
		"@every 5s",
		"@every 1m",
		"@every 1h",
		"@daily",
		"@weekly",
		"@monthly",
		"Custom cron expression",
	}

	choice := c.AskChoice("Choose a schedule", schedules, 0)

	if choice == "Custom cron expression" {
		custom := c.AskRequired("Enter custom cron expression")
		return cronName, custom
	}

	return cronName, choice
}

func (c *MakeCronCommand) createCron(name, schedule string) error {
	cronDir := "./pkg/scheduler"

	if err := os.MkdirAll(cronDir, 0755); err != nil {
		return fmt.Errorf("failed to create scheduler directory: %v", err)
	}

	fileName := strings.ToLower(name) + ".go"
	filePath := filepath.Join(cronDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("cron file %s already exists", filePath)
	}

	structName := c.FormatStructName(name)

	// Use internal stub template
	if err := c.GenerateFromStub("scheduler/cron.go.stub", filePath, CronTemplate{
		StructName: structName,
		CronName:   strings.ToLower(name),
		Schedule:   schedule,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}); err != nil {
		return err
	}

	// Add import to main.go if not already present
	c.HandleAutoImport("pkg/scheduler", "cron scheduler")

	fmt.Printf("✅ Cron scheduler created successfully: %s\n", filePath)
	fmt.Printf("📝 Cron struct: %s\n", structName)
	fmt.Printf("⏰ Schedule: %s\n", schedule)
	fmt.Printf("🚀 Scheduler will be auto-registered via init() function\n")

	return nil
}

type CronTemplate struct {
	StructName string
	CronName   string
	Schedule   string
	Timestamp  string
}
