package provider_test

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/galaplate/core/provider"
	"github.com/gofiber/fiber/v3"
)

type fakeProvider struct {
	name  string
	calls *[]string
}

func (p *fakeProvider) Register() { *p.calls = append(*p.calls, p.name+":register") }
func (p *fakeProvider) Boot()     { *p.calls = append(*p.calls, p.name+":boot") }

type routedProvider struct{ fakeProvider }

func (p *routedProvider) Routes(app *fiber.App) {
	app.Get("/package/ping", provider.Middleware("auth"), func(c fiber.Ctx) error {
		return c.SendString("pong")
	})
}

func (p *routedProvider) Commands() []provider.Command {
	return []provider.Command{fakeCommand{}}
}

type fakeCommand struct{}

func (fakeCommand) GetSignature() string        { return "package:hello" }
func (fakeCommand) GetDescription() string      { return "Say hello" }
func (fakeCommand) Execute(args []string) error { return nil }

func TestBootRegistersAllBeforeBootingAnyAndRunsOnce(t *testing.T) {
	provider.Reset()
	var calls []string
	provider.Add(&fakeProvider{name: "a", calls: &calls})
	provider.Add(&fakeProvider{name: "b", calls: &calls})

	provider.Boot()
	provider.Boot()

	want := []string{"a:register", "b:register", "a:boot", "b:boot"}
	if len(calls) != len(want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("calls = %v, want %v", calls, want)
		}
	}
}

func TestMountRoutesUsesAliasedMiddleware(t *testing.T) {
	provider.Reset()
	var calls []string
	provider.Add(&routedProvider{fakeProvider{name: "pkg", calls: &calls}})
	provider.AliasMiddleware("auth", func(c fiber.Ctx) error {
		if c.Get("Authorization") == "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		return c.Next()
	})

	app := fiber.New()
	provider.MountRoutes(app)

	resp, err := app.Test(httptest.NewRequest("GET", "/package/ping", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}

	req := httptest.NewRequest("GET", "/package/ping", nil)
	req.Header.Set("Authorization", "Bearer x")
	resp, err = app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "pong" {
		t.Fatalf("body = %q, want pong", body)
	}
}

func TestCommandsCollectsProviderCommands(t *testing.T) {
	provider.Reset()
	var calls []string
	provider.Add(&fakeProvider{name: "plain", calls: &calls})
	provider.Add(&routedProvider{fakeProvider{name: "pkg", calls: &calls}})

	cmds := provider.Commands()
	if len(cmds) != 1 || cmds[0].GetSignature() != "package:hello" {
		t.Fatalf("commands = %v", cmds)
	}
}

func TestMiddlewarePanicsOnMissingAlias(t *testing.T) {
	provider.Reset()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for missing alias")
		}
	}()
	provider.Middleware("missing")
}
