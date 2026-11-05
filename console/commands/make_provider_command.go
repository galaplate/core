package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type MakeProviderCommand struct {
	BaseCommand
}

type ProviderTemplate struct {
	StructName   string
	RoutePrefix  string
	WithRoutes   bool
	WithCommands bool
	Timestamp    string
}

func (c *MakeProviderCommand) GetSignature() string {
	return "make:provider"
}

func (c *MakeProviderCommand) GetDescription() string {
	return "Create a new service provider in providers/"
}

func (c *MakeProviderCommand) Execute(args []string) error {
	var name string
	var withRoutes, withCommands bool

	for _, arg := range args {
		switch arg {
		case "--routes":
			withRoutes = true
		case "--commands":
			withCommands = true
		case "-h", "--help":
			c.showUsage()
			return nil
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown option %s", arg)
			}
			if name != "" {
				return fmt.Errorf("unexpected argument %s", arg)
			}
			name = arg
		}
	}

	if name == "" {
		name = c.AskRequired("Enter provider name (e.g., Billing, Notification)")
		withRoutes = c.AskConfirmation("Does it mount routes?", false)
		withCommands = c.AskConfirmation("Does it register console commands?", false)
	}

	return c.createProvider(name, withRoutes, withCommands)
}

func (c *MakeProviderCommand) showUsage() {
	c.ShowUsage("make:provider", c.GetDescription(), []string{
		"make:provider Billing                       # providers/billing_service_provider.go",
		"make:provider Billing --routes --commands   # also stub Routes() and Commands()",
	})
}

func (c *MakeProviderCommand) createProvider(name string, withRoutes, withCommands bool) error {
	base := providerBaseName(name)
	if err := c.ValidateName(base, "Provider"); err != nil {
		return err
	}

	providersDir := "providers"
	_, statErr := os.Stat(providersDir)
	providersDirIsNew := os.IsNotExist(statErr)

	filePath := filepath.Join(providersDir, toSnakeCase(base)+"_service_provider.go")
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("provider file %s already exists", filePath)
	}

	structName := base + "ServiceProvider"
	if err := c.GenerateFromStub("providers/provider.go.stub", filePath, ProviderTemplate{
		StructName:   structName,
		RoutePrefix:  strings.ReplaceAll(toSnakeCase(base), "_", "-"),
		WithRoutes:   withRoutes,
		WithCommands: withCommands,
		Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
	}); err != nil {
		return err
	}

	// Existing apps already import providers from main.go and the test case;
	// only a brand new providers package needs wiring up.
	if providersDirIsNew {
		c.HandleAutoImport("providers", "providers package")
		c.PrintWarning("Also add the import to tests/test_case.go so tests load your providers.")
	}

	c.PrintSuccess(fmt.Sprintf("Provider created successfully: %s", filePath))
	fmt.Printf("📝 Provider struct: %s\n", structName)
	fmt.Printf("🚀 It registers itself via init(); no further wiring needed\n")

	return nil
}

// providerBaseName strips a trailing "ServiceProvider" or "Provider" so that
// Billing, BillingProvider and BillingServiceProvider all generate
// BillingServiceProvider.
func providerBaseName(name string) string {
	base := toCamelCase(name)
	for _, suffix := range []string{"ServiceProvider", "Provider"} {
		if trimmed := strings.TrimSuffix(base, suffix); trimmed != base && trimmed != "" {
			return trimmed
		}
	}
	return base
}

// toCamelCase turns snake/kebab/space separated names into CamelCase while
// leaving already CamelCase names untouched.
func toCamelCase(name string) string {
	return (&BaseCommand{}).FormatStructName(strings.TrimSpace(name))
}

// toSnakeCase converts CamelCase (or kebab-case) to snake_case.
func toSnakeCase(name string) string {
	var b strings.Builder
	runes := []rune(name)
	for i, r := range runes {
		switch {
		case r == '-' || r == ' ':
			b.WriteRune('_')
		case r >= 'A' && r <= 'Z':
			prevLower := i > 0 && runes[i-1] >= 'a' && runes[i-1] <= 'z'
			nextLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
			prevUpper := i > 0 && runes[i-1] >= 'A' && runes[i-1] <= 'Z'
			if i > 0 && (prevLower || (prevUpper && nextLower)) {
				b.WriteRune('_')
			}
			b.WriteRune(r + ('a' - 'A'))
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
