package commands

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// inTempApp runs the test from an empty directory holding a minimal app go.mod.
func inTempApp(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	goMod := "module github.com/acme/shop\n\ngo 1.25.0\n\nrequire (\n\tgithub.com/galaplate/core v0.0.44\n)\n"
	if err := os.WriteFile("go.mod", []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func parseGo(t *testing.T, path string) string {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), path, src, parser.AllErrors); err != nil {
		t.Fatalf("%s is not valid Go: %v\n%s", path, err, src)
	}
	return string(src)
}

func TestEveryStubIsEmbedded(t *testing.T) {
	inTempApp(t) // no local internal/stubs, so only the embed can serve them
	for _, stub := range []string{
		"commands/command.go.stub",
		"dto/dto.go.stub",
		"factories/factory.go.stub",
		"jobs/job.go.stub",
		"models/model.go.stub",
		"policies/policy.go.stub",
		"scheduler/cron.go.stub",
		"seeders/seeder.go.stub",
		"migrations/{{.Timestamp}}_{{.Name}}.go.stub",
		"providers/provider.go.stub",
	} {
		if _, err := (&BaseCommand{}).readStub(stub); err != nil {
			t.Errorf("%s: %v", stub, err)
		}
	}
}

func TestMakeProvider(t *testing.T) {
	tests := []struct {
		args       []string
		file       string
		structName string
		has        []string
		hasNot     []string
	}{
		{
			args:       []string{"Billing"},
			file:       "providers/billing_service_provider.go",
			structName: "BillingServiceProvider",
			hasNot:     []string{"Routes(", "Commands(", "core/fiber"},
		},
		{
			args:       []string{"UserNotificationServiceProvider", "--routes", "--commands"},
			file:       "providers/user_notification_service_provider.go",
			structName: "UserNotificationServiceProvider",
			has:        []string{"Routes(app *fiber.App)", "Commands() []provider.Command", "/api/user-notification"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.structName, func(t *testing.T) {
			inTempApp(t)
			if err := os.Mkdir("providers", 0755); err != nil {
				t.Fatal(err)
			}

			if err := (&MakeProviderCommand{}).Execute(tt.args); err != nil {
				t.Fatal(err)
			}

			src := parseGo(t, tt.file)
			for _, want := range append(tt.has, "type "+tt.structName+" struct{}", "provider.Add(&"+tt.structName+"{})") {
				if !strings.Contains(src, want) {
					t.Errorf("missing %q in\n%s", want, src)
				}
			}
			for _, unwanted := range tt.hasNot {
				if strings.Contains(src, unwanted) {
					t.Errorf("unexpected %q in\n%s", unwanted, src)
				}
			}

			if err := (&MakeProviderCommand{}).Execute(tt.args); err == nil {
				t.Error("expected an error when the provider already exists")
			}
		})
	}
}

func TestMakeProviderRejectsInvalidName(t *testing.T) {
	inTempApp(t)
	if err := (&MakeProviderCommand{}).Execute([]string{"Bad!Name"}); err == nil {
		t.Fatal("expected an error")
	}
}

func TestMakePackage(t *testing.T) {
	inTempApp(t)

	if err := (&MakePackageCommand{}).Execute([]string{"BillingReports", "--dir", "pkgs/billing"}); err != nil {
		t.Fatal(err)
	}

	goMod, err := os.ReadFile("pkgs/billing/go.mod")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"module github.com/acme/billing-reports",
		"go 1.25.0",
		"require github.com/galaplate/core v0.0.44",
	} {
		if !strings.Contains(string(goMod), want) {
			t.Errorf("go.mod missing %q:\n%s", want, goMod)
		}
	}

	provider := parseGo(t, "pkgs/billing/provider.go")
	for _, want := range []string{
		"package billingreports",
		"type BillingReportsServiceProvider struct{}",
		`"/api/billing-reports"`,
		`"github.com/acme/billing-reports/internal/commands"`,
	} {
		if !strings.Contains(provider, want) {
			t.Errorf("provider.go missing %q", want)
		}
	}

	for _, f := range []string{
		"provider_test.go",
		"internal/commands/hello_command.go",
		"internal/controllers/ping_controller.go",
		"internal/migrations/migrations.go",
	} {
		parseGo(t, filepath.Join("pkgs/billing", f))
	}
	for _, f := range []string{"Makefile", "README.md"} {
		if _, err := os.Stat(filepath.Join("pkgs/billing", f)); err != nil {
			t.Error(err)
		}
	}

	if err := (&MakePackageCommand{}).Execute([]string{"billing-reports", "--dir", "pkgs/billing"}); err == nil {
		t.Error("expected an error when the directory is not empty")
	}
}

func TestMakePackageRejectsInvalidName(t *testing.T) {
	inTempApp(t)
	for _, name := range []string{"1blog", "blog!", "-blog"} {
		if err := (&MakePackageCommand{}).Execute([]string{name}); err == nil {
			t.Errorf("%q: expected an error", name)
		}
	}
}

func TestAddBlankImport(t *testing.T) {
	inTempApp(t)
	file := filepath.Join("providers", "packages.go")

	if err := addBlankImport(file, "providers", "github.com/acme/a"); err != nil {
		t.Fatal(err)
	}
	if err := addBlankImport(file, "providers", "github.com/acme/b"); err != nil {
		t.Fatal(err)
	}
	if err := addBlankImport(file, "providers", "github.com/acme/a"); err != nil {
		t.Fatal(err)
	}

	src := parseGo(t, file)
	if strings.Count(src, `_ "github.com/acme/a"`) != 1 || strings.Count(src, `_ "github.com/acme/b"`) != 1 {
		t.Fatalf("unexpected imports:\n%s", src)
	}
}

func TestToSnakeCase(t *testing.T) {
	for in, want := range map[string]string{
		"Billing":          "billing",
		"UserNotification": "user_notification",
		"HTTPClient":       "http_client",
		"billing-reports":  "billing_reports",
		"BillingReports":   "billing_reports",
	} {
		if got := toSnakeCase(in); got != want {
			t.Errorf("toSnakeCase(%q) = %q, want %q", in, got, want)
		}
	}
}
