package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
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

	return c.createSeeder(seederName)
}

func (c *MakeSeederCommand) askForSeederName() string {
	return c.AskRequired("Enter seeder name (e.g., UserSeeder, ProductSeeder)")
}

func (c *MakeSeederCommand) createSeeder(name string) error {
	seederDir := "./db/seeders"

	if err := os.MkdirAll(seederDir, 0755); err != nil {
		return fmt.Errorf("failed to create seeders directory: %v", err)
	}

	fileName := strings.ToLower(name) + ".go"
	filePath := filepath.Join(seederDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("seeder file %s already exists", filePath)
	}

	structName := c.FormatStructName(name)

	templateData := SeederTemplate{
		StructName: structName,
		SeederName: strings.ToLower(name),
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl, err := template.New("seeder").Parse(seederTemplate)
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

const seederTemplate = `package seeders

import (
	"gorm.io/gorm"
    "github.com/galaplate/core/database/seeders"
)

// {{.StructName}} - Generated on {{.Timestamp}}
type {{.StructName}} struct{}

func (s {{.StructName}}) Seed(db *gorm.DB) error {
	// TODO: Add your seeding logic here

	// Example: Create sample users
	// users := []models.User{
	//     {Name: "John Doe", Email: "john@example.com"},
	//     {Name: "Jane Smith", Email: "jane@example.com"},
	// }
	//
	// for _, user := range users {
	//     if err := db.FirstOrCreate(&user, models.User{Email: user.Email}).Error; err != nil {
	//         return err
	//     }
	// }

	return nil
}

func init() {
	seeders.RegisterSeeder("{{.SeederName}}", {{.StructName}}{})
}
`
