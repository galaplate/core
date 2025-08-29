package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
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

	return c.createJob(jobName)
}

func (c *MakeJobCommand) askForJobName() string {
	return c.AskRequired("Enter job name (e.g., ProcessPayment, SendEmail)")
}

func (c *MakeJobCommand) createJob(name string) error {
	jobDir := "./pkg/queue/jobs"

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

	dbConnection, err := c.GetDbConnection()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %v", err)
	}

	structName := c.FormatStructName(name)

	templateData := JobTemplate{
		StructName: structName,
		JobName:    strings.ToLower(name),
		ModuleName: moduleName,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl, err := template.New("job").Parse(jobTemplate)
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

	fmt.Printf("✅ Job created successfully: %s\n", filePath)
	fmt.Printf("📝 Job struct: %s\n", structName)

	if err := c.createMigrationIfNeeded(dbConnection); err != nil {
		fmt.Printf("⚠️  Warning: Could not create migration: %v\n", err)
	}

	return nil
}

func (c *MakeJobCommand) createMigrationIfNeeded(dbConnection string) error {
	migrationsDir := "./db/migrations"

	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(migrationsDir, 0755); err != nil {
			return err
		}
	}

	// Check if jobs table migration already exists
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*_create_jobs_table*.sql"))
	if err != nil {
		return err
	}
	if len(files) > 0 {
		// Migration already exists, don't print anything
		return nil
	}

	stubDir := "./internal/stubs/migrations"
	var stubSuffix string
	switch dbConnection {
	case "mysql":
		stubSuffix = "mysql"
	case "postgres", "postgresql":
		stubSuffix = "pgsql"
	default:
		return fmt.Errorf("unsupported DB connection: %s", dbConnection)
	}

	stubFile := filepath.Join(stubDir, fmt.Sprintf("20250609004425_create_jobs_table.%s.sql.stub", stubSuffix))
	targetFile := filepath.Join(migrationsDir, "20250609004425_create_jobs_table.sql")

	stubContent, err := os.ReadFile(stubFile)
	if err != nil {
		return err
	}

	if err := os.WriteFile(targetFile, stubContent, 0644); err != nil {
		return err
	}

	// Only print this message when migration is actually created
	fmt.Printf("📄 Migration created: %s\n", targetFile)
	return nil
}

type JobTemplate struct {
	StructName string
	JobName    string
	ModuleName string
	Timestamp  string
}

const jobTemplate = `package jobs

import (
	"encoding/json"
	"time"

	"{{.ModuleName}}/pkg/queue"
)

// {{.StructName}} - Generated on {{.Timestamp}}
type {{.StructName}} struct {
	// Add your job payload fields here
	// Example:
	// UserID int ` + "`" + `json:"user_id"` + "`" + `
}

func (j {{.StructName}}) MaxAttempts() int {
	return 3
}

func (j {{.StructName}}) RetryAfter() time.Duration {
	return 2 * time.Minute
}

func ({{.StructName}}) Type() string {
	return "{{.JobName}}"
}

func (j {{.StructName}}) Handle(payload json.RawMessage) error {
	// Unmarshal payload into struct
	// var data {{.StructName}}
	// if err := json.Unmarshal(payload, &data); err != nil {
	//     return err
	// }

	// TODO: Add your job logic here

	return nil
}

func init() {
	queue.RegisterJob({{.StructName}}{})
}
`
