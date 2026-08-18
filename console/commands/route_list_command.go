package commands

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/galaplate/core/fiber"
	"github.com/galaplate/core/logger"
	"slices"
)

type RouteListCommand struct {
	SetupRoutes func(*fiber.App)
}

type RouteInfo struct {
	Method string `json:"method"`
	Path   string `json:"uri"`
}

func (c *RouteListCommand) GetSignature() string {
	return "route:list"
}

func (c *RouteListCommand) GetDescription() string {
	return "Display all registered routes in table or JSON format (use --json flag for JSON output)"
}

func (c *RouteListCommand) Execute(args []string) error {
	if c.SetupRoutes == nil {
		fmt.Println("❌ No router configured. Set SetupRoutes before registering this command.")
		return nil
	}

	outputJSON := c.hasFlag(args, "--json")

	if !outputJSON {
		fmt.Println("🔍 Collecting routes...")
	}

	app := fiber.New(fiber.Config{})

	c.SetupRoutes(app)

	routes := c.extractRoutes(app)

	if len(routes) == 0 {
		if outputJSON {
			c.printJSON([]RouteInfo{})
		} else {
			fmt.Println("❌ No routes found")
		}
		return nil
	}

	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	if outputJSON {
		c.printJSON(routes)
	} else {
		c.printRoutesTable(routes)
	}

	logger.Info("RouteListCommand@Execute", map[string]any{
		"action":      "routes_listed",
		"route_count": len(routes),
		"format":      c.getFormat(outputJSON),
	})

	return nil
}

func (c *RouteListCommand) extractRoutes(app *fiber.App) []RouteInfo {
	routeMap := make(map[string]RouteInfo)

	for _, route := range app.GetRoutes() {
		if route.Method == "HEAD" {
			continue
		}

		if route.Path == "/" && isNonStandardMethod(route.Method) {
			continue
		}

		key := route.Method + ":" + route.Path
		routeMap[key] = RouteInfo{
			Method: route.Method,
			Path:   route.Path,
		}
	}

	var routes []RouteInfo
	for _, route := range routeMap {
		routes = append(routes, route)
	}

	return routes
}

func (c *RouteListCommand) printRoutesTable(routes []RouteInfo) {
	methodWidth := 6
	pathWidth := 4

	for _, route := range routes {
		if len(route.Method) > methodWidth {
			methodWidth = len(route.Method)
		}
		if len(route.Path) > pathWidth {
			pathWidth = len(route.Path)
		}
	}

	methodWidth += 2
	pathWidth += 2

	fmt.Printf("\n")
	fmt.Printf("┌%s┬%s┐\n",
		strings.Repeat("─", methodWidth),
		strings.Repeat("─", pathWidth))

	fmt.Printf("│%s│%s│\n",
		c.centerText("METHOD", methodWidth),
		c.centerText("PATH", pathWidth))

	fmt.Printf("├%s┼%s┤\n",
		strings.Repeat("─", methodWidth),
		strings.Repeat("─", pathWidth))

	for _, route := range routes {
		methodColor := c.getMethodColor(route.Method)
		fmt.Printf("│%s%s\033[0m │ %-*s │\n",
			methodColor,
			c.padText(route.Method, methodWidth-1),
			pathWidth-2,
			route.Path)
	}

	fmt.Printf("└%s┴%s┘\n",
		strings.Repeat("─", methodWidth),
		strings.Repeat("─", pathWidth))

	fmt.Printf("\n📊 Total routes: %d\n\n", len(routes))
}

func (c *RouteListCommand) centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}
	padding := width - len(text)
	leftPad := padding / 2
	rightPad := padding - leftPad
	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}

func (c *RouteListCommand) padText(text string, width int) string {
	if len(text) >= width {
		return text
	}
	return text + strings.Repeat(" ", width-len(text))
}

func (c *RouteListCommand) getMethodColor(method string) string {
	switch method {
	case "GET":
		return "\033[32m"
	case "POST":
		return "\033[33m"
	case "PUT":
		return "\033[34m"
	case "PATCH":
		return "\033[35m"
	case "DELETE":
		return "\033[31m"
	default:
		return "\033[37m"
	}
}

func (c *RouteListCommand) hasFlag(args []string, flag string) bool {
	return slices.Contains(args, flag)
}

func (c *RouteListCommand) printJSON(output []RouteInfo) {
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

func (c *RouteListCommand) getFormat(isJSON bool) string {
	if isJSON {
		return "json"
	}
	return "table"
}

func isNonStandardMethod(method string) bool {
	switch method {
	case "CONNECT", "DELETE", "OPTIONS", "PATCH", "PUT", "TRACE":
		return true
	default:
		return false
	}
}
