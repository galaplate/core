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
	return "Create a new queue job handler (use --domain=DomainName for DDD pattern)"
}

func (c *MakeJobCommand) Execute(args []string) error {
	var jobName string
	var domainName string

	// Parse arguments and flags
	filteredArgs := []string{}
	for i, arg := range args {
		if strings.HasPrefix(arg, "--domain=") {
			domainName = strings.TrimPrefix(arg, "--domain=")
		} else if arg == "--domain" && i+1 < len(args) {
			domainName = args[i+1]
			i++ // Skip next arg as it's the domain value
		} else if !strings.HasPrefix(arg, "-") {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	// Get job name from filtered arguments
	if len(filteredArgs) > 0 {
		jobName = filteredArgs[0]
	}

	if len(args) == 0 || jobName == "" {
		jobName = c.askForJobName()
	}

	if jobName == "" {
		return fmt.Errorf("job name cannot be empty")
	}

	// If domain is specified, use DDD pattern
	if domainName != "" {
		return c.createDomainJob(jobName, domainName)
	}

	// Otherwise use traditional MVC pattern
	return c.createJob(jobName)
}

func (c *MakeJobCommand) askForJobName() string {
	return c.AskRequired("Enter job name (e.g., ProcessPayment, SendEmail)")
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

	// Add import to main.go if not already present
	c.HandleAutoImport("pkg/jobs", "job")

	fmt.Printf("✅ Job created successfully: %s\n", filePath)
	fmt.Printf("📝 Job struct: %s\n", structName)
	fmt.Printf("🚀 Job will be auto-registered via init() function\n")

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

func (c *MakeJobCommand) createDomainJob(name, domainName string) error {
	domainLower := strings.ToLower(domainName)
	jobDir := filepath.Join("domains", domainLower, "jobs")

	if err := os.MkdirAll(jobDir, 0755); err != nil {
		return fmt.Errorf("failed to create domain jobs directory: %v", err)
	}

	fileName := strings.ToLower(name) + ".go"
	filePath := filepath.Join(jobDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("domain job file %s already exists", filePath)
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

	templateData := DomainJobTemplate{
		StructName:  structName,
		JobName:     strings.ToLower(name),
		DomainName:  c.FormatStructName(domainName),
		DomainLower: domainLower,
		ModuleName:  moduleName,
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl, err := template.New("domainJob").Parse(domainJobTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse domain job template: %v", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to write template: %v", err)
	}

	// Add import to main.go if not already present
	c.HandleAutoImport(fmt.Sprintf("domains/%s/jobs", domainLower), "domain job")

	fmt.Printf("✅ Domain job created successfully: %s\n", filePath)
	fmt.Printf("📝 Job struct: %s (in %s domain)\n", structName, domainName)
	fmt.Printf("🚀 Job will be auto-registered via init() function\n")
	fmt.Printf("💡 Tip: Use 'go run main.go console make:domain %s --all' to generate complete DDD structure\n", domainName)

	if err := c.createMigrationIfNeeded(dbConnection); err != nil {
		fmt.Printf("⚠️  Warning: Could not create migration: %v\n", err)
	}

	return nil
}

type JobTemplate struct {
	StructName string
	JobName    string
	ModuleName string
	Timestamp  string
}

type DomainJobTemplate struct {
	StructName  string
	JobName     string
	DomainName  string
	DomainLower string
	ModuleName  string
	Timestamp   string
}

const jobTemplate = `package jobs

import (
	"encoding/json"
	"time"

	"github.com/galaplate/core/queue"
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

const domainJobTemplate = `package jobs

import (
	"encoding/json"
	"time"

	"github.com/galaplate/core/queue"
)

// {{.StructName}} - Generated on {{.Timestamp}}
// Domain job for {{.DomainName}} domain
type {{.StructName}} struct {
	// Add your job payload fields here
	// Example:
	// UserID int ` + "`" + `json:"user_id"` + "`" + `
	// Data   map[string]interface{} ` + "`" + `json:"data"` + "`" + `
}

func (j {{.StructName}}) MaxAttempts() int {
	return 3
}

func (j {{.StructName}}) RetryAfter() time.Duration {
	return 2 * time.Minute
}

func ({{.StructName}}) Type() string {
	return "{{.DomainLower}}.{{.JobName}}"
}

func (j {{.StructName}}) Handle(payload json.RawMessage) error {
	// Unmarshal payload into struct
	// var data {{.StructName}}
	// if err := json.Unmarshal(payload, &data); err != nil {
	//     return err
	// }

	// TODO: Add your domain-specific job logic here
	// You can use domain services, repositories, etc.

	return nil
}

func init() {
	queue.RegisterJob({{.StructName}}{})
}
`
