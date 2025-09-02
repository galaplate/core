package policies

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

// PolicyResult represents the result of a policy evaluation
type PolicyResult struct {
	Allowed bool
	Message string
	Code    int
}

// PolicyContext contains information needed for policy evaluation
type PolicyContext struct {
	User     any
	Request  *fiber.Ctx
	Resource string
	Action   string
	Data     map[string]any
}

// Policy interface that all policies must implement
type Policy interface {
	Name() string
	Evaluate(ctx context.Context, policyCtx *PolicyContext) PolicyResult
}
