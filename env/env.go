package env

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func Get(key string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}

	err := godotenv.Load(".env")
	if err != nil {
		fmt.Print("Error loading .env file")
	}

	return os.Getenv(key)
}
