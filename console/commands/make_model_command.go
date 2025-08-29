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
	return "Create a new model"
}

func (c *MakeModelCommand) Execute(args []string) error {
	var modelName string

	if len(args) == 0 {
		modelName = c.askForModelName()
	} else {
		modelName = args[0]
	}

	if modelName == "" {
		return fmt.Errorf("model name cannot be empty")
	}

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

type ModelTemplate struct {
	StructName string
	Timestamp  string
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
