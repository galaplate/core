package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const defaultPackageGoVersion = "1.25.0"

type MakePackageCommand struct {
	BaseCommand
}

type PackageTemplate struct {
	Name        string // kebab-case, used for route and command prefixes
	PackageName string // Go package identifier
	TablePrefix string // snake_case, for migration and table names
	StructName  string
	ModulePath  string
	GoVersion   string
	CoreVersion string
	Timestamp   string
}

// packageStubs maps stub templates to their path inside the new package.
var packageStubs = []struct{ stub, target string }{
	{"package/go.mod.stub", "go.mod"},
	{"package/provider.go.stub", "provider.go"},
	{"package/provider_test.go.stub", "provider_test.go"},
	{"package/internal/commands/hello_command.go.stub", "internal/commands/hello_command.go"},
	{"package/internal/controllers/ping_controller.go.stub", "internal/controllers/ping_controller.go"},
	{"package/internal/migrations/migrations.go.stub", "internal/migrations/migrations.go"},
	{"package/Makefile.stub", "Makefile"},
	{"package/README.md.stub", "README.md"},
}

var packageNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

func (c *MakePackageCommand) GetSignature() string {
	return "make:package"
}

func (c *MakePackageCommand) GetDescription() string {
	return "Scaffold a reusable Galaplate package (separate Go module)"
}

func (c *MakePackageCommand) Execute(args []string) error {
	var name, module, dir string
	var link bool

	for i := 0; i < len(args); i++ {
		arg := args[i]
		key, value, hasValue := strings.Cut(arg, "=")
		switch key {
		case "--module", "--dir":
			if !hasValue {
				if i+1 >= len(args) {
					return fmt.Errorf("%s needs a value", key)
				}
				i++
				value = args[i]
			}
			if key == "--module" {
				module = value
			} else {
				dir = value
			}
		case "--link":
			link = true
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

	interactive := name == ""
	if interactive {
		name = c.AskRequired("Enter package name (e.g., blog, billing-reports)")
	}

	name = strings.ReplaceAll(toSnakeCase(strings.TrimSpace(name)), "_", "-")
	if !packageNamePattern.MatchString(name) {
		return fmt.Errorf("package name %q must be lowercase letters, digits and dashes, starting with a letter (e.g., blog, billing-reports)", name)
	}

	appModule, _ := c.GetModuleName()
	if module == "" {
		module = defaultPackageModule(appModule, name)
		if interactive {
			module = c.AskText("Go module path", module)
		}
	}
	if dir == "" {
		dir = filepath.Join("..", name)
		if interactive {
			dir = c.AskText("Directory", dir)
		}
	}
	if interactive && !link && appModule != "" {
		link = c.AskConfirmation("Link it into this application for local development?", true)
	}

	if err := c.createPackage(name, module, dir); err != nil {
		return err
	}

	if link {
		if err := c.linkPackage(module, dir); err != nil {
			c.PrintWarning(fmt.Sprintf("Could not link the package: %v", err))
			link = false
		}
	}

	c.printNextSteps(module, dir, link)
	return nil
}

func (c *MakePackageCommand) showUsage() {
	c.ShowUsage("make:package", c.GetDescription(), []string{
		"make:package blog                                   # ../blog",
		"make:package blog --module github.com/acme/blog     # custom module path",
		"make:package blog --dir ./packages/blog             # custom directory",
		"make:package blog --link                            # also replace + import it in this app",
	})
}

func (c *MakePackageCommand) createPackage(name, module, dir string) error {
	if entries, err := os.ReadDir(dir); err == nil && len(entries) > 0 {
		return fmt.Errorf("directory %s already exists and is not empty", dir)
	}

	goVersion, coreVersion := readAppGoMod("go.mod")
	if goVersion == "" {
		goVersion = defaultPackageGoVersion
	}

	data := PackageTemplate{
		Name:        name,
		PackageName: strings.ReplaceAll(name, "-", ""),
		TablePrefix: strings.ReplaceAll(name, "-", "_"),
		StructName:  toCamelCase(name) + "ServiceProvider",
		ModulePath:  module,
		GoVersion:   goVersion,
		CoreVersion: coreVersion,
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
	}

	for _, f := range packageStubs {
		target := filepath.Join(dir, filepath.FromSlash(f.target))
		if err := c.GenerateFromStub(f.stub, target, data); err != nil {
			return err
		}
	}

	c.PrintSuccess(fmt.Sprintf("Package created successfully: %s", dir))
	fmt.Printf("📦 Module:   %s\n", module)
	fmt.Printf("📝 Provider: %s.%s\n", data.PackageName, data.StructName)
	if coreVersion == "" {
		c.PrintWarning("Couldn't read the github.com/galaplate/core version from ./go.mod; run `go get github.com/galaplate/core` in the package.")
	}

	return nil
}

// linkPackage points this application at the local package and blank-imports
// it from providers/packages.go, which main.go and the tests already load.
func (c *MakePackageCommand) linkPackage(module, dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	replacePath, err := filepath.Rel(cwd, absDir)
	if err != nil {
		replacePath = absDir
	}
	replacePath = filepath.ToSlash(replacePath)
	if !strings.HasPrefix(replacePath, ".") && !filepath.IsAbs(replacePath) {
		replacePath = "./" + replacePath
	}

	cmd := exec.Command("go", "mod", "edit",
		"-require="+module+"@v0.0.0",
		"-replace="+module+"="+replacePath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod edit: %v: %s", err, strings.TrimSpace(string(out)))
	}

	if err := addBlankImport(filepath.Join("providers", "packages.go"), "providers", module); err != nil {
		return err
	}

	c.PrintSuccess(fmt.Sprintf("Linked into this application (replace %s => %s, imported in providers/packages.go)", module, replacePath))
	return nil
}

func (c *MakePackageCommand) printNextSteps(module, dir string, linked bool) {
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s && go mod tidy && make test\n", dir)
	if linked {
		fmt.Println("  go run main.go console list        # shows the package's commands")
		fmt.Println("  go run main.go console route:list  # shows the package's routes")
		fmt.Println("  Remove the replace directive from go.mod before releasing the application.")
		return
	}
	fmt.Println()
	fmt.Println("To use it in an application:")
	fmt.Printf("  go get %s\n", module)
	fmt.Printf("  and add  _ \"%s\"  to providers/packages.go\n", module)
}

// defaultPackageModule derives a module path next to the application's, e.g.
// github.com/acme/shop -> github.com/acme/blog.
func defaultPackageModule(appModule, name string) string {
	if i := strings.LastIndex(appModule, "/"); i > 0 {
		return appModule[:i] + "/" + name
	}
	return name
}

// readAppGoMod returns the go directive and the github.com/galaplate/core
// version from a go.mod file, or empty strings when they can't be read.
func readAppGoMod(path string) (goVersion, coreVersion string) {
	file, err := os.Open(path)
	if err != nil {
		return "", ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		switch {
		case len(fields) >= 2 && fields[0] == "go":
			goVersion = fields[1]
		case len(fields) >= 2 && fields[0] == "github.com/galaplate/core" && strings.HasPrefix(fields[1], "v"):
			coreVersion = fields[1]
		case len(fields) >= 3 && fields[0] == "require" && fields[1] == "github.com/galaplate/core":
			coreVersion = fields[2]
		}
	}
	return goVersion, coreVersion
}

// addBlankImport adds `_ "importPath"` to the import block of file, creating
// the file (in package pkg) when it doesn't exist.
func addBlankImport(file, pkg, importPath string) error {
	line := fmt.Sprintf("\t_ %q", importPath)

	content, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			return err
		}
		src := fmt.Sprintf("// Packages installed into this application. Each blank import registers\n// the package's service provider.\npackage %s\n\nimport (\n%s\n)\n", pkg, line)
		return os.WriteFile(file, []byte(src), 0644)
	}
	if err != nil {
		return err
	}

	src := string(content)
	if strings.Contains(src, fmt.Sprintf("%q", importPath)) {
		return nil
	}

	start := strings.Index(src, "import (")
	if start < 0 {
		return fmt.Errorf("no import block in %s; add %s manually", file, strings.TrimSpace(line))
	}
	end := strings.Index(src[start:], "\n)")
	if end < 0 {
		return fmt.Errorf("unterminated import block in %s", file)
	}
	end += start

	src = src[:end] + "\n" + line + src[end:]
	return os.WriteFile(file, []byte(src), 0644)
}
