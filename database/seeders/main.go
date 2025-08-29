package seeders

import (
	"fmt"
	"os"
	"strings"

	"gorm.io/gorm"
)

type Seeder interface {
	Seed(db *gorm.DB) error
}

type DatabaseSeeder struct {
	seeders []Seeder
}

func (ds *DatabaseSeeder) AddSeeder(seeder Seeder) {
	ds.seeders = append(ds.seeders, seeder)
}

func (ds *DatabaseSeeder) Run(db *gorm.DB) error {
	fmt.Println("Starting database seeding...")

	for _, seeder := range ds.seeders {
		seederName := fmt.Sprintf("%T", seeder)
		fmt.Printf("Running %s...\n", seederName)

		if err := seeder.Seed(db); err != nil {
			return fmt.Errorf("seeder %s failed: %w", seederName, err)
		}

		fmt.Printf("%s completed successfully\n", seederName)
	}

	fmt.Println("Database seeding completed!")
	return nil
}

func NewDatabaseSeeder(selectedCSV string) *DatabaseSeeder {
	ds := &DatabaseSeeder{}

	if selectedCSV == "" {
		for _, s := range SeederRegistry {
			ds.AddSeeder(s)
		}
		return ds
	}

	for _, name := range strings.Split(selectedCSV, ",") {
		seeder, ok := SeederRegistry[strings.ToLower(name)]
		if !ok {
			fmt.Printf("Seeder file '%s.go' not found or not registered\n", name)
			os.Exit(1)
		}
		ds.AddSeeder(seeder)
	}

	return ds
}

var SeederRegistry = map[string]Seeder{}

func RegisterSeeder(name string, seeder Seeder) {
	SeederRegistry[name] = seeder
}
