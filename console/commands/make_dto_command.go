package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	// Validate DTO name format
	if err := c.ValidateName(dtoName, "DTO"); err != nil {
		return err
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

	// Use internal stub template
	if err := c.GenerateFromStub("dto/dto.go.stub", filePath, DtoTemplate{
		StructName: structName,
		ModuleName: moduleName,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}); err != nil {
		return err
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
