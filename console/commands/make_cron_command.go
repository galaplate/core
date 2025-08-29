package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
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

	templateData := CronTemplate{
		StructName: structName,
		CronName:   strings.ToLower(name),
		Schedule:   schedule,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl, err := template.New("cron").Parse(cronTemplate)
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

	fmt.Printf("✅ Cron scheduler created successfully: %s\n", filePath)
	fmt.Printf("📝 Cron struct: %s\n", structName)
	fmt.Printf("⏰ Schedule: %s\n", schedule)

	return nil
}

type CronTemplate struct {
	StructName string
	CronName   string
	Schedule   string
	Timestamp  string
}

const cronTemplate = `package scheduler

// {{.StructName}} - Generated on {{.Timestamp}}
type {{.StructName}} struct{}

func ({{.StructName}}) Handle() (string, func()) {
	// Cron expression examples:
	// "@every 5s"     - every 5 seconds
	// "@every 1m"     - every minute
	// "@every 1h"     - every hour
	// "@daily"        - once a day (midnight)
	// "@weekly"       - once a week (Sunday midnight)
	// "@monthly"      - once a month (first day of month, midnight)
	// "0 30 * * * *"  - every 30 seconds
	// "0 0 12 * * *"  - every day at noon
	// "0 15 10 * * *" - 10:15 AM every day

	return "{{.Schedule}}", func() {
		// TODO: Add your scheduled task logic here

		// Example: Log a message
		// logger.Info("{{.StructName}} scheduler executed")

		// Example: Database operation
		// database.Connect.Model(&models.User{}).Where("active = ?", true).Count(&count)

		// Example: Queue a job
		// queue.Dispatch(jobs.SomeJob{UserID: 1})
	}
}

func init() {
	RegisterScheduler("{{.CronName}}", {{.StructName}}{}.Handle)
}
`
