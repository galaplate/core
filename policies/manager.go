package policies

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// PolicyManager manages and executes policies
type PolicyManager struct {
	policies map[string]Policy
}

// NewPolicyManager creates a new policy manager
func NewPolicyManager() *PolicyManager {
	return &PolicyManager{
		policies: make(map[string]Policy),
	}
}

// RegisterPolicy registers a new policy
func (pm *PolicyManager) RegisterPolicy(policy Policy) {
	pm.policies[policy.Name()] = policy
}

// GetPolicy retrieves a policy by name
func (pm *PolicyManager) GetPolicy(name string) (Policy, bool) {
	policy, exists := pm.policies[name]
	return policy, exists
}

// EvaluatePolicy evaluates a single policy
func (pm *PolicyManager) EvaluatePolicy(ctx context.Context, policyName string, policyCtx *PolicyContext) PolicyResult {
	policy, exists := pm.policies[policyName]
	if !exists {
		return PolicyResult{
			Allowed: false,
			Message: fmt.Sprintf("Policy '%s' not found", policyName),
			Code:    fiber.StatusInternalServerError,
		}
	}

	return policy.Evaluate(ctx, policyCtx)
}

// EvaluatePolicies evaluates multiple policies (all must pass)
func (pm *PolicyManager) EvaluatePolicies(ctx context.Context, policyNames []string, policyCtx *PolicyContext) PolicyResult {
	for _, policyName := range policyNames {
		result := pm.EvaluatePolicy(ctx, policyName, policyCtx)
		if !result.Allowed {
			return result
		}
	}

	return PolicyResult{
		Allowed: true,
		Message: "All policies passed",
		Code:    fiber.StatusOK,
	}
}

// PolicyMiddleware creates a Fiber middleware that enforces policies
func (pm *PolicyManager) PolicyMiddleware(policyNames ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		policyCtx := &PolicyContext{
			Request:  c,
			Resource: c.Route().Path,
			Action:   c.Method(),
			Data:     make(map[string]any),
		}

		// Try to get user from context (set by auth middleware)
		if user := c.Locals("user"); user != nil {
			policyCtx.User = user
		}

		result := pm.EvaluatePolicies(ctx, policyNames, policyCtx)
		if !result.Allowed {
			return c.Status(result.Code).JSON(fiber.Map{
				"success": false,
				"message": result.Message,
			})
		}

		return c.Next()
	}
}

// Global policy manager instance
var GlobalPolicyManager = NewPolicyManager()

// WithPolicies creates policy middleware using the global manager
func WithPolicies(policyNames ...string) fiber.Handler {
	return GlobalPolicyManager.PolicyMiddleware(policyNames...)
}
