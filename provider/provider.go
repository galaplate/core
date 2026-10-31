package provider

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v3"
)

// Provider is the Laravel-style service provider contract. Register runs for
// every provider before any Boot, so Boot may rely on what others registered.
type Provider interface {
	Register()
	Boot()
}

// RouteProvider is implemented by providers that own HTTP routes. Routes are
// mounted after the application's own routes.
type RouteProvider interface {
	Routes(app *fiber.App)
}

// Command mirrors console.Command so providers can ship console commands
// without importing the console package.
type Command interface {
	GetSignature() string
	GetDescription() string
	Execute(args []string) error
}

// CommandProvider is implemented by providers that ship console commands.
type CommandProvider interface {
	Commands() []Command
}

var (
	mu         sync.RWMutex
	providers  []Provider
	middleware = map[string]fiber.Handler{}
	booted     bool
)

// Add registers a provider, typically from the package's init function.
func Add(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	providers = append(providers, p)
}

// All returns the registered providers in registration order.
func All() []Provider {
	mu.RLock()
	defer mu.RUnlock()
	return append([]Provider(nil), providers...)
}

// Boot calls Register on every provider, then Boot on every provider. It only
// runs once.
func Boot() {
	mu.Lock()
	if booted {
		mu.Unlock()
		return
	}
	booted = true
	mu.Unlock()

	all := All()
	for _, p := range all {
		p.Register()
	}
	for _, p := range all {
		p.Boot()
	}
}

// MountRoutes mounts the routes of every RouteProvider onto app.
func MountRoutes(app *fiber.App) {
	for _, p := range All() {
		if rp, ok := p.(RouteProvider); ok {
			rp.Routes(app)
		}
	}
}

// Commands collects the console commands of every CommandProvider.
func Commands() []Command {
	var cmds []Command
	for _, p := range All() {
		if cp, ok := p.(CommandProvider); ok {
			cmds = append(cmds, cp.Commands()...)
		}
	}
	return cmds
}

// AliasMiddleware names a middleware so provider routes can use it without
// importing the application.
func AliasMiddleware(name string, h fiber.Handler) {
	mu.Lock()
	defer mu.Unlock()
	middleware[name] = h
}

// Middleware returns the middleware aliased as name. It panics when the alias
// is missing so misconfiguration fails at startup instead of on a request.
func Middleware(name string) fiber.Handler {
	mu.RLock()
	defer mu.RUnlock()
	h, ok := middleware[name]
	if !ok {
		panic(fmt.Sprintf("provider: middleware %q is not aliased", name))
	}
	return h
}

// Reset clears all providers and middleware aliases. Intended for tests.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	providers = nil
	middleware = map[string]fiber.Handler{}
	booted = false
}
