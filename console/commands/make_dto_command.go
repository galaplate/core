package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type MakeDtoCommand struct {
	BaseCommand
}

func (c *MakeDtoCommand) GetSignature() string {
	return "make:dto"
}

func (c *MakeDtoCommand) GetDescription() string {
	return "Create a new DTO (Data Transfer Object)"
}

func (c *MakeDtoCommand) Execute(args []string) error {
	var dtoName string

	if len(args) == 0 {
		dtoName = c.askForDtoName()
	} else {
		dtoName = args[0]
	}

	if dtoName == "" {
		return fmt.Errorf("DTO name cannot be empty")
	}

	return c.createDto(dtoName)
}

func (c *MakeDtoCommand) askForDtoName() string {
	return c.AskRequired("Enter DTO name (e.g., UserCreate, ProductUpdate)")
}

func (c *MakeDtoCommand) createDto(name string) error {
	dtoDir := "./pkg/dto"

	if err := os.MkdirAll(dtoDir, 0755); err != nil {
		return fmt.Errorf("failed to create dto directory: %v", err)
	}

	fileName := strings.ToLower(name) + ".go"
	filePath := filepath.Join(dtoDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("DTO file %s already exists", filePath)
	}

	moduleName, err := c.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %v", err)
	}

	structName := c.FormatStructName(name)

	templateData := DtoTemplate{
		StructName: structName,
		ModuleName: moduleName,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl, err := template.New("dto").Parse(dtoTemplate)
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

	fmt.Printf("✅ DTO created successfully: %s\n", filePath)
	fmt.Printf("📝 DTO struct: %s\n", structName)

	return nil
}

type DtoTemplate struct {
	StructName string
	ModuleName string
	Timestamp  string
}

const dtoTemplate = `package dto

import (
	"github.com/gofiber/fiber/v2"
	"github.com/galaplate/core/supports"
)

// {{.StructName}} - Generated on {{.Timestamp}}
type {{.StructName}} struct {
	// Add your DTO fields here
	// Example:
	// Name  string ` + "`" + `json:"name" validate:"required"` + "`" + `
	// Email string ` + "`" + `json:"email" validate:"required,email"` + "`" + `
}

func (s *{{.StructName}}) Validate(c *fiber.Ctx) (u *{{.StructName}}, err error) {
	myValidator := &supports.XValidator{}
	if err := c.BodyParser(s); err != nil {
		return nil, err
	}

	if err := myValidator.Validate(s); err != nil {
		return nil, &fiber.Error{
			Code:    fiber.ErrUnprocessableEntity.Code,
			Message: err.Error(),
		}
	}

	return s, nil
}
`
