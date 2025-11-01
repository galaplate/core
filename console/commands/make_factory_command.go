package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type MakeFactoryCommand struct {
	BaseCommand
}

func (c *MakeFactoryCommand) GetSignature() string {
	return "make:factory"
}

func (c *MakeFactoryCommand) GetDescription() string {
	return "Create a new factory"
}

func (c *MakeFactoryCommand) Execute(args []string) error {
	var modelName string

	if len(args) == 0 {
		modelName = c.askForModelName()
	} else {
		modelName = args[0]
	}

	if modelName == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	return c.createFactory(modelName)
}

func (c *MakeFactoryCommand) askForModelName() string {
	return c.AskRequired("Enter model name (e.g., User, Product)")
}

func (c *MakeFactoryCommand) createFactory(name string) error {
	factoryDir := "./db/factories"

	if err := os.MkdirAll(factoryDir, 0755); err != nil {
		return fmt.Errorf("failed to create factories directory: %v", err)
	}

	fileName := strings.ToLower(name) + "_factory.go"
	filePath := filepath.Join(factoryDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("factory file %s already exists", filePath)
	}

	moduleName, err := c.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %v", err)
	}

	structName := c.FormatStructName(name)
	factoryName := structName + "Factory"
	factoryConstructor := "New" + factoryName

	// Use internal stub template
	if err := c.GenerateFromStub("factories/factory.go.stub", filePath, FactoryTemplate{
		FactoryName:        factoryName,
		FactoryConstructor: factoryConstructor,
		ModelName:          structName,
		Timestamp:          time.Now().Format("2006-01-02 15:04:05"),
		ModuleName:         moduleName,
	}); err != nil {
		return err
	}

	fmt.Printf("✅ Factory created successfully: %s\n", filePath)
	fmt.Printf("📝 Factory struct: %s\n", factoryName)
	fmt.Printf("📝 Model: %s\n", structName)

	return nil
}

type FactoryTemplate struct {
	FactoryName        string
	FactoryConstructor string
	ModelName          string
	Timestamp          string
	ModuleName         string
}
