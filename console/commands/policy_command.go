package commands

import (
	"fmt"
	"os"
	"strings"
)

// PolicyCommand handles policy-related operations
type PolicyCommand struct {
	BaseCommand
}

// NewPolicyCommand creates a new policy command
func NewPolicyCommand() *PolicyCommand {
	cmd := &PolicyCommand{}
	return cmd
}

// GetSignature returns the command signature
func (c *PolicyCommand) GetSignature() string {
	return "make:policy"
}

// GetDescription returns the command description
func (c *PolicyCommand) GetDescription() string {
	return "Create a new policy"
}

// Execute processes the policy command
func (c *PolicyCommand) Execute(args []string) error {
	var policyName string

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		c.ShowUsage(
			c.GetSignature(),
			c.GetDescription(),
			[]string{
				"make:policy <policy_name>    - Create a new custom policy",
			},
		)
		return nil
	}

	if len(args) == 0 {
		policyName = c.AskRequired("Enter policy name (e.g., admin_only, user_verified)")
	} else {
		policyName = args[0]
	}
	policyName = strings.ToLower(policyName)

	return c.createPolicy(policyName)
}

// createPolicy creates a new custom policy template
func (c *PolicyCommand) createPolicy(policyName string) error {
	structName := c.FormatStructName(policyName)

	c.PrintInfo(fmt.Sprintf("Creating new policy: %s", policyName))

	// Create policy file
	policyContent := c.generatePolicyTemplate(policyName, structName)
	policyPath := fmt.Sprintf("./pkg/policies/%s_policy.go", policyName)

	// Create directory if it doesn't exist
	if err := os.MkdirAll("./pkg/policies", 0755); err != nil {
		c.PrintError(fmt.Sprintf("Failed to create policies directory: %v", err))
		return err
	}

	if err := os.WriteFile(policyPath, []byte(policyContent), 0644); err != nil {
		c.PrintError(fmt.Sprintf("Failed to create policy file: %v", err))
		return err
	}

	c.PrintSuccess(fmt.Sprintf("Policy created: %s", policyPath))
	fmt.Println()

	// Add import to main.go if not already present
	c.HandleAutoImport("pkg/policies", "policy")

	// Show registration instructions
	c.PrintInfo("Next steps:")
	fmt.Println("1. Implement your policy logic in the Evaluate method")
	fmt.Println("2. Register the policy in your application:")
	fmt.Printf("   policies.GlobalPolicyManager.RegisterPolicy(New%sPolicy())\n", structName)
	fmt.Println("3. Use in routes:")
	fmt.Printf("   policies.WithPolicies(\"%s\")\n", policyName)

	return nil
}

func (c *PolicyCommand) generatePolicyTemplate(policyName, structName string) string {
	return fmt.Sprintf(`package policies

import (
	"context"

	"github.com/galaplate/core/policies"
	"github.com/gofiber/fiber/v2"
)

// %sPolicy implements %s policy
type %sPolicy struct {
	// Add your policy configuration fields here
}

func New%sPolicy() *%sPolicy {
	return &%sPolicy{
		// Initialize your policy configuration here
	}
}

func (p *%sPolicy) Name() string {
	return "%s"
}

func (p *%sPolicy) Evaluate(ctx context.Context, policyCtx *policies.PolicyContext) policies.PolicyResult {
	// TODO: Implement your policy logic here

	// Example policy logic:
	// Check user authentication
	if policyCtx.User == nil {
		return policies.PolicyResult{
			Allowed: false,
			Message: "Authentication required for %s policy",
			Code:    fiber.StatusUnauthorized,
		}
	}

	// Add your custom policy checks here
	// For example:
	// - Check user permissions
	// - Validate request data
	// - Check business rules
	// - etc.

	return policies.PolicyResult{
		Allowed: true,
		Message: "%s policy check passed",
		Code:    fiber.StatusOK,
	}
}

func init() {
    policies.GlobalPolicyManager.RegisterPolicy(New%sPolicy())
}

`, structName, policyName, structName, structName, structName, structName, structName, policyName, structName, policyName, structName, structName)
}
