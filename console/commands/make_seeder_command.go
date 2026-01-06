package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type MakeSeederCommand struct {
	BaseCommand
}

func (c *MakeSeederCommand) GetSignature() string {
	return "make:seeder"
}

func (c *MakeSeederCommand) GetDescription() string {
	return "Create a new database seeder"
}

func (c *MakeSeederCommand) Execute(args []string) error {
	var seederName string

	if len(args) == 0 {
		seederName = c.askForSeederName()
	} else {
		seederName = args[0]
	}

	if seederName == "" {
		return fmt.Errorf("seeder name cannot be empty")
	}

	// Validate seeder name format
	if err := c.ValidateName(seederName, "Seeder"); err != nil {
		return err
	}

	return c.createSeeder(seederName)
}

func (c *MakeSeederCommand) askForSeederName() string {
	return c.AskRequired("Enter seeder name (e.g., UserSeeder, ProductSeeder)")
}

func (c *MakeSeederCommand) createSeeder(name string) error {
	seederDir := "db/seeders"

	if err := os.MkdirAll(seederDir, 0755); err != nil {
		return fmt.Errorf("failed to create seeders directory: %v", err)
	}

	fileName := strings.ToLower(name) + ".go"
	filePath := filepath.Join(seederDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("seeder file %s already exists", filePath)
	}

	structName := c.FormatStructName(name)

	// Use internal stub template
	if err := c.GenerateFromStub("seeders/seeder.go.stub", filePath, SeederTemplate{
		StructName: structName,
		SeederName: strings.ToLower(name),
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}); err != nil {
		return err
	}

	// Add import to main.go if not already present
	c.HandleAutoImport("db/seeders", "seeder")

	fmt.Printf("✅ Seeder created successfully: %s\n", filePath)
	fmt.Printf("📝 Seeder struct: %s\n", structName)
	fmt.Printf("🌱 Run with: go run main.go console db:seed %s\n", strings.ToLower(name))

	return nil
}

type SeederTemplate struct {
	StructName string
	SeederName string
	Timestamp  string
}
