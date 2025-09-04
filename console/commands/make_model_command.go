package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	// Use internal stub template
	if err := c.GenerateFromStub("models/model.go.stub", filePath, ModelTemplate{
		StructName: structName,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}); err != nil {
		return err
	}

	fmt.Printf("✅ Model created successfully: %s\n", filePath)
	fmt.Printf("📝 Model struct: %s\n", structName)

	return nil
}

type ModelTemplate struct {
	StructName string
	Timestamp  string
}
