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
	return "Create a new DTO (Data Transfer Object) (use --domain=DomainName for DDD pattern)"
}

func (c *MakeDtoCommand) Execute(args []string) error {
	var dtoName string
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

	// Get DTO name from filtered arguments
	if len(filteredArgs) > 0 {
		dtoName = filteredArgs[0]
	}

	if len(args) == 0 || dtoName == "" {
		dtoName = c.askForDtoName()
	}

	if dtoName == "" {
		return fmt.Errorf("DTO name cannot be empty")
	}

	// If domain is specified, use DDD pattern
	if domainName != "" {
		return c.createDomainDto(dtoName, domainName)
	}

	// Otherwise use traditional MVC pattern
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

	tmpl, err := template.New("dto").Parse(mvcDtoTemplate)
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

func (c *MakeDtoCommand) createDomainDto(name, domainName string) error {
	domainLower := strings.ToLower(domainName)
	dtoDir := filepath.Join("domains", domainLower, "dto")

	if err := os.MkdirAll(dtoDir, 0755); err != nil {
		return fmt.Errorf("failed to create domain dto directory: %v", err)
	}

	fileName := strings.ToLower(name) + "_dto.go"
	filePath := filepath.Join(dtoDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("domain DTO file %s already exists", filePath)
	}

	moduleName, err := c.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %v", err)
	}

	structName := c.FormatStructName(name)

	templateData := DomainDtoTemplate{
		StructName:  structName,
		DomainName:  c.FormatStructName(domainName),
		DomainLower: domainLower,
		ModuleName:  moduleName,
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl, err := template.New("domainDto").Parse(domainDtoTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse domain dto template: %v", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to write template: %v", err)
	}

	fmt.Printf("✅ Domain DTO created successfully: %s\n", filePath)
	fmt.Printf("📝 DTO struct: %s (in %s domain)\n", structName, domainName)
	fmt.Printf("💡 Tip: Use 'go run main.go console make:domain %s --all' to generate complete DDD structure\n", domainName)

	return nil
}

type DtoTemplate struct {
	StructName string
	ModuleName string
	Timestamp  string
}

type DomainDtoTemplate struct {
	StructName  string
	DomainName  string
	DomainLower string
	ModuleName  string
	Timestamp   string
}

const mvcDtoTemplate = `package dto

import (
	"github.com/gofiber/fiber/v2"
	"github.com/galaplate/core/utils"
)

// {{.StructName}} - Generated on {{.Timestamp}}
type {{.StructName}} struct {
	// Add your DTO fields here
	// Example:
	// Name  string ` + "`" + `json:"name" validate:"required"` + "`" + `
	// Email string ` + "`" + `json:"email" validate:"required,email"` + "`" + `
}

func (s *{{.StructName}}) Validate(c *fiber.Ctx) (u *{{.StructName}}, err error) {
	myValidator := &utils.XValidator{}
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

const domainDtoTemplate = `package dto

import "time"

// Create{{.StructName}}Request represents the request for creating a {{.StructName}}
type Create{{.StructName}}Request struct {
	// Add your request fields here
	// Example:
	// Name        string ` + "`" + `json:"name" validate:"required,min=2,max=100"` + "`" + `
	// Description string ` + "`" + `json:"description" validate:"max=500"` + "`" + `
}

// Update{{.StructName}}Request represents the request for updating a {{.StructName}}
type Update{{.StructName}}Request struct {
	// Add your update fields here
	// Example:
	// Name        string ` + "`" + `json:"name" validate:"omitempty,min=2,max=100"` + "`" + `
	// Description string ` + "`" + `json:"description" validate:"omitempty,max=500"` + "`" + `
}

// {{.StructName}}Response represents the response for a {{.StructName}}
type {{.StructName}}Response struct {
	ID        uint      ` + "`" + `json:"id"` + "`" + `
	CreatedAt time.Time ` + "`" + `json:"created_at"` + "`" + `
	UpdatedAt time.Time ` + "`" + `json:"updated_at"` + "`" + `
	
	// Add your response fields here
	// Example:
	// Name        string ` + "`" + `json:"name"` + "`" + `
	// Description string ` + "`" + `json:"description"` + "`" + `
	// Status      string ` + "`" + `json:"status"` + "`" + `
}

// List{{.StructName}}Request represents the request for listing {{.StructName}}s
type List{{.StructName}}Request struct {
	Page  int ` + "`" + `json:"page" query:"page" validate:"min=1"` + "`" + `
	Limit int ` + "`" + `json:"limit" query:"limit" validate:"min=1,max=100"` + "`" + `
	
	// Add your filter fields here
	// Example:
	// Search string ` + "`" + `json:"search" query:"search"` + "`" + `
	// Status string ` + "`" + `json:"status" query:"status"` + "`" + `
}

// List{{.StructName}}Response represents the response for listing {{.StructName}}s
type List{{.StructName}}Response struct {
	Data       []*{{.StructName}}Response ` + "`" + `json:"data"` + "`" + `
	Total      int64                      ` + "`" + `json:"total"` + "`" + `
	Page       int                        ` + "`" + `json:"page"` + "`" + `
	Limit      int                        ` + "`" + `json:"limit"` + "`" + `
	TotalPages int64                      ` + "`" + `json:"total_pages"` + "`" + `
}
`
