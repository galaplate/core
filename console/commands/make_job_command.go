package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type MakeJobCommand struct {
	BaseCommand
}

func (c *MakeJobCommand) GetSignature() string {
	return "make:job"
}

func (c *MakeJobCommand) GetDescription() string {
	return "Create a new queue job handler"
}

func (c *MakeJobCommand) Execute(args []string) error {
	var jobName string

	if len(args) == 0 {
		jobName = c.askForJobName()
	} else {
		jobName = args[0]
	}

	if jobName == "" {
		return fmt.Errorf("job name cannot be empty")
	}

	// Validate job name format
	if err := c.ValidateName(jobName, "Job"); err != nil {
		return err
	}

	return c.createJob(jobName)
}

func (c *MakeJobCommand) askForJobName() string {
	name := c.AskRequired("Enter job name (e.g., ProcessPayment, SendEmail)")
	return strings.TrimSpace(name)
}

func (c *MakeJobCommand) createJob(name string) error {
	jobDir := "./pkg/jobs"

	if err := os.MkdirAll(jobDir, 0755); err != nil {
		return fmt.Errorf("failed to create jobs directory: %v", err)
	}

	fileName := strings.ToLower(name) + ".go"
	filePath := filepath.Join(jobDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("job file %s already exists", filePath)
	}

	moduleName, err := c.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %v", err)
	}

	structName := c.FormatStructName(name)

	// Ask for additional configuration
	maxAttempts := c.askMaxAttempts()
	retryMinutes := c.askRetryAfter()

	// Use internal stub template
	if err := c.GenerateFromStub("jobs/job.go.stub", filePath, JobTemplate{
		StructName:   structName,
		JobName:      strings.ToLower(name),
		ModuleName:   moduleName,
		Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		MaxAttempts:  maxAttempts,
		RetryMinutes: retryMinutes,
	}); err != nil {
		return err
	}

	// Add import to main.go if not already present
	c.HandleAutoImport("pkg/jobs", "job")

	c.PrintSuccess(fmt.Sprintf("Job created successfully: %s", filePath))
	fmt.Printf("📝 Job struct:    %s\n", structName)
	fmt.Printf("🔄 Max attempts: %d\n", maxAttempts)
	fmt.Printf("⏳ Retry after:  %d minutes\n", retryMinutes)
	fmt.Printf("🚀 Job will be auto-registered via init() function\n")

	// Create migration for jobs table if needed
	if err := c.createMigrationIfNeeded(); err != nil {
		fmt.Printf("⚠️  Warning: Could not create migration: %v\n", err)
	}

	// Display usage example
	c.displayJobUsageExample(structName)

	return nil
}

// askMaxAttempts prompts user for maximum retry attempts
func (c *MakeJobCommand) askMaxAttempts() int {
	for {
		input := c.AskText("Maximum retry attempts", "3")
		attempts, err := strconv.Atoi(input)
		if err != nil || attempts < 1 {
			fmt.Println("❌ Please enter a valid number (minimum 1)")
			continue
		}
		return attempts
	}
}

// askRetryAfter prompts user for retry delay in minutes
func (c *MakeJobCommand) askRetryAfter() int {
	for {
		input := c.AskText("Retry delay in minutes", "2")
		minutes, err := strconv.Atoi(input)
		if err != nil || minutes < 0 {
			fmt.Println("❌ Please enter a valid number (minimum 0)")
			continue
		}
		return minutes
	}
}

func (c *MakeJobCommand) createMigrationIfNeeded() error {
	// Check if jobs table already exists in the database
	if c.TableExists("jobs") {
		fmt.Printf("ℹ️  Jobs table already exists in database\n")
		return nil
	}

	migrationsDir := "db/migrations"

	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(migrationsDir, 0755); err != nil {
			return err
		}
	}

	// Check if jobs table migration already exists (Go-based migrations)
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*_create_jobs_table.go"))
	if err != nil {
		return err
	}
	if len(files) > 0 {
		fmt.Printf("ℹ️  Migration for jobs table already exists\n")
		return nil
	}

	// Generate unique timestamp for migration
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	targetFile := filepath.Join(migrationsDir, fmt.Sprintf("%s_create_jobs_table.go", timestamp))

	// Generate the jobs table migration using the Go stub
	if err := c.GenerateFromStub("migrations/{{.Timestamp}}_{{.Name}}.go.stub", targetFile, MigrationTemplate{
		Timestamp: timestamp,
		Name:      "create_jobs_table",
	}); err != nil {
		return err
	}

	// Update the migration content to include jobs table schema
	return c.updateJobsTableMigration(targetFile, timestamp)
}

// updateJobsTableMigration updates the generated migration to create the jobs table
func (c *MakeJobCommand) updateJobsTableMigration(filePath, timestamp string) error {
	jobsTableSchema := `func (m *Migration{{.Timestamp}}) Up(schema *database.Schema) error {
	return schema.Create("jobs", func(table *database.Blueprint) {
		table.ID()
		table.String("type").NotNullable()
		table.JSON("payload").Nullable()
		table.String("state", 16).NotNullable().Default("pending")
		table.Text("error_msg").Nullable()
		table.Integer("attempts").Nullable()
		table.DateTime("available_at").Nullable()
		table.DateTime("created_at").NotNullable().Default("CURRENT_TIMESTAMP")
		table.DateTime("started_at").Nullable()
		table.DateTime("finished_at").Nullable()

		// Indexes
		table.Index([]string{"state"})
		table.Index([]string{"created_at"})
		table.Index([]string{"available_at"})
	})
}

func (m *Migration{{.Timestamp}}) Down(schema *database.Schema) error {
	return schema.DropIfExists("jobs")
}`

	// Read the generated file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// Replace placeholder methods with actual implementation
	// Find and replace the Up and Down methods
	updatedContent := strings.ReplaceAll(contentStr, `func (m *Migration{{.Timestamp}}) Up(schema *database.Schema) error {
	// Add your migration code here
	// Example:
	// return schema.Create("users", func(table *database.Blueprint) {
	//     table.ID()
	//     table.String("name").NotNullable()
	//     table.String("email").NotNullable().Unique()
	//     table.Timestamps()
	// })
	return nil
}

func (m *Migration{{.Timestamp}}) Down(schema *database.Schema) error {
	// Add your rollback code here
	// Example:
	// return schema.DropIfExists("users")
	return nil
}`, jobsTableSchema)

	// Restore template variable
	updatedContent = strings.ReplaceAll(updatedContent, "{{.Timestamp}}", timestamp)

	return os.WriteFile(filePath, []byte(updatedContent), 0644)
}

// displayJobUsageExample shows how to use the job
func (c *MakeJobCommand) displayJobUsageExample(structName string) {
	fmt.Println()
	fmt.Println("📌 Quick start guide:")
	fmt.Println()
	fmt.Printf("1. Add your job payload fields:\n")
	fmt.Printf("   type %s struct {\n", structName)
	fmt.Printf("       UserID int    `json:\"user_id\"`\n")
	fmt.Printf("       Email  string `json:\"email\"`\n")
	fmt.Printf("   }\n")
	fmt.Println()
	fmt.Printf("2. Implement the Handle method with your job logic\n")
	fmt.Println()
	fmt.Printf("3. Dispatch the job from your controller or service:\n")
	fmt.Printf("   queue.Dispatch(%s{}, userID, email)\n", structName)
	fmt.Println()
	fmt.Printf("4. Run migrations to create the jobs table:\n")
	fmt.Printf("   go run main.go console db:up\n")
	fmt.Println()
}

type JobTemplate struct {
	StructName   string
	JobName      string
	ModuleName   string
	Timestamp    string
	MaxAttempts  int
	RetryMinutes int
}

type MigrationTemplate struct {
	Timestamp string
	Name      string
}
