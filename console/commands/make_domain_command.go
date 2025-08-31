package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type MakeDomainCommand struct {
	BaseCommand
}

func (c *MakeDomainCommand) GetSignature() string {
	return "make:domain"
}

func (c *MakeDomainCommand) GetDescription() string {
	return "Create a new domain with DDD structure (entities, repositories, services, etc.)"
}

func (c *MakeDomainCommand) Execute(args []string) error {
	var domainName string
	var createAll bool

	// Parse arguments and flags
	for _, arg := range args {
		if arg == "--all" || arg == "-a" {
			createAll = true
			continue
		}
		if domainName == "" && !strings.HasPrefix(arg, "-") {
			domainName = arg
		}
	}

	if domainName == "" {
		domainName = c.askForDomainName()
	}

	if domainName == "" {
		return fmt.Errorf("domain name cannot be empty")
	}

	return c.createDomain(domainName, createAll)
}

func (c *MakeDomainCommand) askForDomainName() string {
	return c.AskRequired("Enter domain name (e.g., User, Product, Order)")
}

func (c *MakeDomainCommand) createDomain(name string, createAll bool) error {
	domainName := c.FormatStructName(name)
	domainLower := strings.ToLower(name)

	// Create domain directory structure
	domainDir := filepath.Join("domains", domainLower)

	directories := []string{
		filepath.Join(domainDir, "models"),
		filepath.Join(domainDir, "handlers"),
		filepath.Join(domainDir, "dto"),
		filepath.Join(domainDir, "validators"),
		filepath.Join(domainDir, "jobs"),
	}

	// Create directories
	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	moduleName, err := c.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %v", err)
	}

	if createAll {
		// Create all domain components
		if err := c.createModel(domainDir, domainName, moduleName); err != nil {
			return err
		}
		if err := c.createHandler(domainDir, domainName, moduleName); err != nil {
			return err
		}
		if err := c.createDTO(domainDir, domainName, moduleName); err != nil {
			return err
		}
		if err := c.createValidator(domainDir, domainName, moduleName); err != nil {
			return err
		}
	} else {
		// Interactive mode - ask what to create
		components := []string{
			"Model",
			"Handler",
			"DTO",
			"Validator",
		}

		for _, component := range components {
			if c.AskConfirmation(fmt.Sprintf("Create %s for %s domain?", component, domainName), true) {
				switch component {
				case "Model":
					if err := c.createModel(domainDir, domainName, moduleName); err != nil {
						return err
					}
				case "Handler":
					if err := c.createHandler(domainDir, domainName, moduleName); err != nil {
						return err
					}
				case "DTO":
					if err := c.createDTO(domainDir, domainName, moduleName); err != nil {
						return err
					}
				case "Validator":
					if err := c.createValidator(domainDir, domainName, moduleName); err != nil {
						return err
					}
				}
			}
		}
	}

	fmt.Printf("✅ Domain '%s' created successfully in: %s\n", domainName, domainDir)
	fmt.Printf("📁 Directory structure:\n")
	for _, dir := range directories {
		fmt.Printf("   - %s/\n", dir)
	}

	return nil
}

func (c *MakeDomainCommand) createModel(domainDir, domainName, moduleName string) error {
	fileName := strings.ToLower(domainName) + ".go"
	filePath := filepath.Join(domainDir, "models", fileName)

	templateData := DomainEntityTemplate{
		DomainName: domainName,
		ModuleName: moduleName,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl, err := template.New("model").Parse(domainCommandModelTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse domain model template: %v", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create model file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to write model template: %v", err)
	}

	fmt.Printf("✅ Model created: %s\n", filePath)
	return nil
}

func (c *MakeDomainCommand) createHandler(domainDir, domainName, moduleName string) error {
	fileName := strings.ToLower(domainName) + "_handler.go"
	filePath := filepath.Join(domainDir, "handlers", fileName)

	templateData := DomainHandlerTemplate{
		DomainName:  domainName,
		DomainLower: strings.ToLower(domainName),
		ModuleName:  moduleName,
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl, err := template.New("handler").Parse(handlerTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse handler template: %v", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create handler file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to write handler template: %v", err)
	}

	fmt.Printf("✅ Handler created: %s\n", filePath)
	return nil
}

func (c *MakeDomainCommand) createDTO(domainDir, domainName, moduleName string) error {
	fileName := strings.ToLower(domainName) + "_dto.go"
	filePath := filepath.Join(domainDir, "dto", fileName)

	templateData := DomainDTOTemplate{
		DomainName: domainName,
		ModuleName: moduleName,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl, err := template.New("dto").Parse(dtoTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse DTO template: %v", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create DTO file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to write DTO template: %v", err)
	}

	fmt.Printf("✅ DTO created: %s\n", filePath)
	return nil
}

func (c *MakeDomainCommand) createValidator(domainDir, domainName, moduleName string) error {
	fileName := strings.ToLower(domainName) + "_validator.go"
	filePath := filepath.Join(domainDir, "validators", fileName)

	templateData := DomainValidatorTemplate{
		DomainName: domainName,
		ModuleName: moduleName,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl, err := template.New("validator").Parse(validatorTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse validator template: %v", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create validator file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to write validator template: %v", err)
	}

	fmt.Printf("✅ Validator created: %s\n", filePath)
	return nil
}

// Template structs
type DomainEntityTemplate struct {
	DomainName string
	ModuleName string
	Timestamp  string
}

type DomainHandlerTemplate struct {
	DomainName  string
	DomainLower string
	ModuleName  string
	Timestamp   string
}

type DomainDTOTemplate struct {
	DomainName string
	ModuleName string
	Timestamp  string
}

type DomainValidatorTemplate struct {
	DomainName string
	ModuleName string
	Timestamp  string
}

// Templates
const domainCommandModelTemplate = `package models

import (
	"time"

	"gorm.io/gorm"
)

// {{.DomainName}} - Generated on {{.Timestamp}}
// Domain model for {{.DomainName}} domain
type {{.DomainName}} struct {
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

// Add your domain business logic methods here
// Example:
// func (u *{{.DomainName}}) IsActive() bool {
//     return u.Status == "active"
// }
`



const handlerTemplate = `package handlers

import (
	"strconv"

	"{{.ModuleName}}/domains/{{.DomainLower}}/dto"
	"{{.ModuleName}}/domains/{{.DomainLower}}/services"
	
	"github.com/galaplate/core/logger"
	"github.com/gofiber/fiber/v2"
)

// {{.DomainName}}Handler handles HTTP requests for {{.DomainName}} domain
type {{.DomainName}}Handler struct {
	service services.{{.DomainName}}ServiceInterface
	logger  *logger.LogRequest
}

// New{{.DomainName}}Handler creates a new {{.DomainName}} handler
func New{{.DomainName}}Handler(service services.{{.DomainName}}ServiceInterface) *{{.DomainName}}Handler {
	return &{{.DomainName}}Handler{
		service: service,
		logger:  logger.NewLogRequestWithUUID(logger.WithField("handler", "{{.DomainName}}Handler"), "{{.DomainLower}}-handler"),
	}
}

// Create handles POST /{{.DomainLower}}s
func (h *{{.DomainName}}Handler) Create(c *fiber.Ctx) error {
	var req dto.Create{{.DomainName}}Request
	if err := c.BodyParser(&req); err != nil {
		h.logger.Logger.Error(map[string]any{"error": err.Error(), "action": "parse_create_{{.DomainLower}}_request"})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result, err := h.service.Create(&req)
	if err != nil {
		h.logger.Logger.Error(map[string]any{"error": err.Error(), "action": "create_{{.DomainLower}}_handler"})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "{{.DomainName}} created successfully",
		"data":    result,
	})
}

// GetByID handles GET /{{.DomainLower}}s/:id
func (h *{{.DomainName}}Handler) GetByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	result, err := h.service.GetByID(uint(id))
	if err != nil {
		h.logger.Logger.Error(map[string]any{"error": err.Error(), "{{.DomainLower}}_id": id, "action": "get_{{.DomainLower}}_by_id_handler"})
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "{{.DomainName}} not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": result,
	})
}

// GetAll handles GET /{{.DomainLower}}s
func (h *{{.DomainName}}Handler) GetAll(c *fiber.Ctx) error {
	req := dto.List{{.DomainName}}Request{
		Page:  c.QueryInt("page", 1),
		Limit: c.QueryInt("limit", 10),
	}

	result, err := h.service.GetAll(&req)
	if err != nil {
		h.logger.Logger.Error(map[string]any{"error": err.Error(), "action": "get_all_{{.DomainLower}}s_handler"})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":        result.Data,
		"total":       result.Total,
		"page":        result.Page,
		"limit":       result.Limit,
		"total_pages": result.TotalPages,
	})
}

// Update handles PUT /{{.DomainLower}}s/:id
func (h *{{.DomainName}}Handler) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	var req dto.Update{{.DomainName}}Request
	if err := c.BodyParser(&req); err != nil {
		h.logger.Logger.Error(map[string]any{"error": err.Error(), "{{.DomainLower}}_id": id, "action": "parse_update_{{.DomainLower}}_request"})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result, err := h.service.Update(uint(id), &req)
	if err != nil {
		h.logger.Logger.Error(map[string]any{"error": err.Error(), "{{.DomainLower}}_id": id, "action": "update_{{.DomainLower}}_handler"})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "{{.DomainName}} updated successfully",
		"data":    result,
	})
}

// Delete handles DELETE /{{.DomainLower}}s/:id
func (h *{{.DomainName}}Handler) Delete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	if err := h.service.Delete(uint(id)); err != nil {
		h.logger.Logger.Error(map[string]any{"error": err.Error(), "{{.DomainLower}}_id": id, "action": "delete_{{.DomainLower}}_handler"})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "{{.DomainName}} deleted successfully",
	})
}
`

const dtoTemplate = `package dto

import "time"

// Create{{.DomainName}}Request represents the request for creating a {{.DomainName}}
type Create{{.DomainName}}Request struct {
	// Add your request fields here
	// Example:
	// Name        string ` + "`" + `json:"name" validate:"required,min=2,max=100"` + "`" + `
	// Description string ` + "`" + `json:"description" validate:"max=500"` + "`" + `
}

// Update{{.DomainName}}Request represents the request for updating a {{.DomainName}}
type Update{{.DomainName}}Request struct {
	// Add your update fields here
	// Example:
	// Name        string ` + "`" + `json:"name" validate:"omitempty,min=2,max=100"` + "`" + `
	// Description string ` + "`" + `json:"description" validate:"omitempty,max=500"` + "`" + `
}

// {{.DomainName}}Response represents the response for a {{.DomainName}}
type {{.DomainName}}Response struct {
	ID        uint      ` + "`" + `json:"id"` + "`" + `
	CreatedAt time.Time ` + "`" + `json:"created_at"` + "`" + `
	UpdatedAt time.Time ` + "`" + `json:"updated_at"` + "`" + `
	
	// Add your response fields here
	// Example:
	// Name        string ` + "`" + `json:"name"` + "`" + `
	// Description string ` + "`" + `json:"description"` + "`" + `
	// Status      string ` + "`" + `json:"status"` + "`" + `
}

// List{{.DomainName}}Request represents the request for listing {{.DomainName}}s
type List{{.DomainName}}Request struct {
	Page  int ` + "`" + `json:"page" query:"page" validate:"min=1"` + "`" + `
	Limit int ` + "`" + `json:"limit" query:"limit" validate:"min=1,max=100"` + "`" + `
	
	// Add your filter fields here
	// Example:
	// Search string ` + "`" + `json:"search" query:"search"` + "`" + `
	// Status string ` + "`" + `json:"status" query:"status"` + "`" + `
}

// List{{.DomainName}}Response represents the response for listing {{.DomainName}}s
type List{{.DomainName}}Response struct {
	Data       []*{{.DomainName}}Response ` + "`" + `json:"data"` + "`" + `
	Total      int64                      ` + "`" + `json:"total"` + "`" + `
	Page       int                        ` + "`" + `json:"page"` + "`" + `
	Limit      int                        ` + "`" + `json:"limit"` + "`" + `
	TotalPages int64                      ` + "`" + `json:"total_pages"` + "`" + `
}
`

const validatorTemplate = `package validators

import (
	"{{.ModuleName}}/domains/{{.DomainLower}}/dto"
	
	"github.com/galaplate/core/utils"
)

// {{.DomainName}}Validator handles validation for {{.DomainName}} domain
type {{.DomainName}}Validator struct {
	validator *utils.Validator
}

// New{{.DomainName}}Validator creates a new {{.DomainName}} validator
func New{{.DomainName}}Validator() *{{.DomainName}}Validator {
	return &{{.DomainName}}Validator{
		validator: &utils.Validator{},
	}
}

// ValidateCreate validates Create{{.DomainName}}Request
func (v *{{.DomainName}}Validator) ValidateCreate(req *dto.Create{{.DomainName}}Request) error {
	return v.validator.Validate(req)
}

// ValidateUpdate validates Update{{.DomainName}}Request  
func (v *{{.DomainName}}Validator) ValidateUpdate(req *dto.Update{{.DomainName}}Request) error {
	return v.validator.Validate(req)
}

// Custom validation methods can be added here
// Example:
// func (v *{{.DomainName}}Validator) ValidateBusinessRule({{.DomainLower}} *entities.{{.DomainName}}) error {
//     // Add custom business logic validation
//     return nil
// }
`
