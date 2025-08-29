package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// BaseCommand provides common functionality for all commands
type BaseCommand struct{}

// askText prompts for text input with optional default value
func (b *BaseCommand) AskText(prompt string, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	if input == "" && defaultValue != "" {
		return defaultValue
	}
	return input
}

// AskRequired prompts for required input (won't accept empty)
func (b *BaseCommand) AskRequired(prompt string) string {
	for {
		fmt.Printf("%s: ", prompt)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if input != "" {
			return input
		}
		fmt.Println("❌ This field is required. Please try again.")
	}
}

// AskConfirmation prompts for yes/no confirmation
func (b *BaseCommand) AskConfirmation(prompt string, defaultValue bool) bool {
	defaultStr := "y/N"
	if defaultValue {
		defaultStr = "Y/n"
	}

	for {
		fmt.Printf("%s [%s]: ", prompt, defaultStr)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.ToLower(strings.TrimSpace(scanner.Text()))

		switch input {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		case "":
			return defaultValue
		default:
			fmt.Println("❌ Please answer yes (y) or no (n)")
		}
	}
}

// AskChoice prompts user to choose from a list of options
func (b *BaseCommand) AskChoice(prompt string, choices []string, defaultIndex int) string {
	fmt.Printf("%s\n", prompt)
	for i, choice := range choices {
		marker := " "
		if i == defaultIndex {
			marker = "*"
		}
		fmt.Printf("  %s %d) %s\n", marker, i+1, choice)
	}

	for {
		fmt.Printf("Choose [1-%d] (default: %d): ", len(choices), defaultIndex+1)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			return choices[defaultIndex]
		}

		if index, err := strconv.Atoi(input); err == nil {
			if index >= 1 && index <= len(choices) {
				return choices[index-1]
			}
		}

		fmt.Printf("❌ Please choose a number between 1 and %d\n", len(choices))
	}
}

// AskNumber prompts for numeric input
func (b *BaseCommand) AskNumber(prompt string, defaultValue int) int {
	for {
		defaultStr := fmt.Sprintf("%d", defaultValue)
		fmt.Printf("%s [%s]: ", prompt, defaultStr)

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			return defaultValue
		}

		if number, err := strconv.Atoi(input); err == nil {
			return number
		}

		fmt.Println("❌ Please enter a valid number")
	}
}

// AskPassword prompts for password input (attempts to hide input)
func (b *BaseCommand) AskPassword(prompt string) string {
	fmt.Printf("%s: ", prompt)

	// Simple password input - in production you'd use golang.org/x/term
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

// AskList prompts for comma-separated list input
func (b *BaseCommand) AskList(prompt string, defaultValues []string) []string {
	defaultStr := strings.Join(defaultValues, ", ")
	fmt.Printf("%s [%s]: ", prompt, defaultStr)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	if input == "" {
		return defaultValues
	}

	var result []string
	for _, item := range strings.Split(input, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}

// GetModuleName reads module name from go.mod
func (b *BaseCommand) GetModuleName() (string, error) {
	file, err := os.Open("go.mod")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("module not found in go.mod")
}

func (b *BaseCommand) FormatStructName(name string) string {
	re := regexp.MustCompile(`[_\-\s]+`)
	words := re.Split(name, -1)

	var result string
	for _, word := range words {
		if len(word) > 0 {
			result += strings.ToUpper(word[:1]) + word[1:]
		}
	}

	if result == "" {
		return "Dto"
	}

	return result
}

// GetDbConnection reads DB_CONNECTION from .env
func (b *BaseCommand) GetDbConnection() (string, error) {
	file, err := os.Open(".env")
	if err != nil {
		// Default to mysql if .env doesn't exist
		return "mysql", nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "DB_CONNECTION=") {
			return strings.TrimPrefix(line, "DB_CONNECTION="), nil
		}
	}

	// Default to mysql if not found
	return "mysql", nil
}

// LoadEnvVariables loads environment variables from .env file
func (b *BaseCommand) LoadEnvVariables() error {
	file, err := os.Open(".env")
	if err != nil {
		return fmt.Errorf(".env file not found: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// BuildDatabaseURL builds database URL from environment variables
func (b *BaseCommand) BuildDatabaseURL() (string, error) {
	if err := b.LoadEnvVariables(); err != nil {
		return "", err
	}

	dbConnection := os.Getenv("DB_CONNECTION")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbDatabase := os.Getenv("DB_DATABASE")
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")

	if dbConnection == "" || dbHost == "" || dbPort == "" || dbDatabase == "" {
		return "", fmt.Errorf("missing database configuration in .env file")
	}

	switch dbConnection {
	case "mysql":
		return fmt.Sprintf("mysql://%s:%s@%s:%s/%s?parseTime=true",
			dbUsername, dbPassword, dbHost, dbPort, dbDatabase), nil
	case "postgres", "postgresql":
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbUsername, dbPassword, dbHost, dbPort, dbDatabase), nil
	default:
		return "", fmt.Errorf("unsupported database driver: %s", dbConnection)
	}
}

// CheckDbmate checks if dbmate is installed
func (b *BaseCommand) CheckDbmate() error {
	_, err := exec.LookPath("dbmate")
	if err != nil {
		return fmt.Errorf("dbmate is not installed. Install with: go install github.com/amacneil/dbmate@latest")
	}
	return nil
}

// RunDbmate executes dbmate command with given arguments
func (b *BaseCommand) RunDbmate(args ...string) error {
	dbURL, err := b.BuildDatabaseURL()
	if err != nil {
		return err
	}

	migrationDir := "./db/migrations"
	if err := os.MkdirAll(migrationDir, 0755); err != nil {
		return fmt.Errorf("failed to create migration directory: %v", err)
	}

	cmdArgs := append([]string{"-d", migrationDir, "--url", dbURL}, args...)
	cmd := exec.Command("dbmate", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// PrintSuccess prints a success message with checkmark
func (b *BaseCommand) PrintSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

// PrintError prints an error message with X mark
func (b *BaseCommand) PrintError(message string) {
	fmt.Printf("❌ %s\n", message)
}

// PrintWarning prints a warning message with warning symbol
func (b *BaseCommand) PrintWarning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// PrintInfo prints an info message with info symbol
func (b *BaseCommand) PrintInfo(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}
