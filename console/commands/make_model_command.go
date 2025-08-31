package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type MakeModelCommand struct {
	BaseCommand
}

func (c *MakeModelCommand) GetSignature() string {
	return "make:model"
}

func (c *MakeModelCommand) GetDescription() string {
	return "Create a new model (use --domain=DomainName for DDD pattern)"
}

func (c *MakeModelCommand) Execute(args []string) error {
	var modelName string
	var domainName string

	// Parse arguments and flags
	filteredArgs := []string{}
	skipNext := false
	for i, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}

		if strings.HasPrefix(arg, "--domain=") {
			domainName = strings.TrimPrefix(arg, "--domain=")
		} else if arg == "--domain" && i+1 < len(args) {
			domainName = args[i+1]
			skipNext = true
		} else if !strings.HasPrefix(arg, "-") {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	// Get model name from filtered arguments
	if len(filteredArgs) > 0 {
		modelName = filteredArgs[0]
	} else if len(args) == 0 || (len(args) > 0 && len(filteredArgs) == 0) {
		// Only prompt if no args provided OR args were provided but were all flags
		modelName = c.askForModelName()
	}

	if modelName == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	// If domain is specified, use DDD pattern
	if domainName != "" {
		return c.createDomainModel(modelName, domainName)
	}

	// Otherwise use traditional MVC pattern
	return c.createModel(modelName)
}

func (c *MakeModelCommand) askForModelName() string {
	return c.AskRequired("Enter model name (e.g., User, Product)")
}

func (c *MakeModelCommand) createModel(name string) error {
	modelDir := "./pkg/models"

	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return fmt.Errorf("failed to create models directory: %v", err)
	}

	fileName := strings.ToLower(name) + ".go"
	filePath := filepath.Join(modelDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("model file %s already exists", filePath)
	}

	structName := c.FormatStructName(name)

	templateData := ModelTemplate{
		StructName: structName,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl, err := template.New("model").Parse(modelTemplate)
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

	fmt.Printf("✅ Model created successfully: %s\n", filePath)
	fmt.Printf("📝 Model struct: %s\n", structName)

	return nil
}

func (c *MakeModelCommand) createDomainModel(name, domainName string) error {
	domainLower := strings.ToLower(domainName)
	modelDir := filepath.Join("domains", domainLower, "models")

	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return fmt.Errorf("failed to create domain models directory: %v", err)
	}

	fileName := strings.ToLower(name) + ".go"
	filePath := filepath.Join(modelDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("domain model file %s already exists", filePath)
	}

	structName := c.FormatStructName(name)

	templateData := DomainModelTemplate{
		StructName:  structName,
		DomainName:  c.FormatStructName(domainName),
		DomainLower: domainLower,
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl, err := template.New("domainModel").Parse(domainModelTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse domain model template: %v", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to write template: %v", err)
	}

	fmt.Printf("✅ Domain model created successfully: %s\n", filePath)
	fmt.Printf("📝 Model struct: %s (in %s domain)\n", structName, domainName)
	fmt.Printf("💡 Tip: Use 'go run main.go console make:domain %s --all' to generate complete DDD structure\n", domainName)

	return nil
}

type ModelTemplate struct {
	StructName string
	Timestamp  string
}

type DomainModelTemplate struct {
	StructName  string
	DomainName  string
	DomainLower string
	Timestamp   string
}

const modelTemplate = `package models

import (
	"time"

	"gorm.io/gorm"
)

// {{.StructName}} - Generated on {{.Timestamp}}
type {{.StructName}} struct {
	ID        uint           ` + "`" + `gorm:"primaryKey" json:"id"` + "`" + `
	CreatedAt time.Time      ` + "`" + `json:"created_at"` + "`" + `
	UpdatedAt time.Time      ` + "`" + `json:"updated_at"` + "`" + `
	DeletedAt gorm.DeletedAt ` + "`" + `gorm:"index" json:"deleted_at"` + "`" + `

	// Add your model fields here
}
`

const domainModelTemplate = `package models

import (
	"time"

	"gorm.io/gorm"
)

// {{.StructName}} - Generated on {{.Timestamp}}
// Domain model for {{.DomainName}} domain
type {{.StructName}} struct {
	ID        uint           ` + "`" + `gorm:"primaryKey" json:"id"` + "`" + `
	CreatedAt time.Time      ` + "`" + `json:"created_at"` + "`" + `
	UpdatedAt time.Time      ` + "`" + `json:"updated_at"` + "`" + `
	DeletedAt gorm.DeletedAt ` + "`" + `gorm:"index" json:"deleted_at"` + "`" + `

	// Add your domain-specific fields here
	// Example:
	// Name        string ` + "`" + `json:"name" gorm:"not null"` + "`" + `
	// Description string ` + "`" + `json:"description"` + "`" + `
	// Status      string ` + "`" + `json:"status" gorm:"default:'active'"` + "`" + `
}

// TableName specifies the table name for GORM
func ({{.StructName}}) TableName() string {
	return "{{.DomainLower}}_{{.StructName}}"
}

// Add your domain business logic methods here
// Example:
// func (u *{{.StructName}}) IsActive() bool {
//     return u.Status == "active"
// }
`
