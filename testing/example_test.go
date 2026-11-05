package testing_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/galaplate/core/database"
	coretesting "github.com/galaplate/core/testing"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/suite"
)

// exampleConfig returns a TestConfig pointing at the self-contained fixture in
// testdata/config (sqlite + a fixed app key), with the sqlite file and local
// uploads placed in a per-suite temp dir so the examples run without an app.
func exampleConfig(t *testing.T) *coretesting.TestConfig {
	_, file, _, _ := runtime.Caller(0)
	tmp := t.TempDir()
	t.Setenv("DB_SQLITE_PATH", filepath.Join(tmp, "test.db"))
	t.Setenv("FILESYSTEM_LOCAL_PATH", filepath.Join(tmp, "uploads"))

	cfg := coretesting.DefaultTestConfig()
	cfg.ConfigPath = filepath.Join(filepath.Dir(file), "testdata", "config")
	return cfg
}

type createUsersTable struct {
	database.BaseMigration
}

func (m *createUsersTable) Up(schema *database.Schema) error {
	return schema.Create("users", func(table *database.Blueprint) {
		table.ID()
		table.String("name")
		table.Timestamps()
	})
}

func (m *createUsersTable) Down(schema *database.Schema) error {
	return schema.DropIfExists("users")
}

func init() {
	database.Register(&createUsersTable{
		BaseMigration: database.BaseMigration{Name: "create_users_table", Timestamp: 1},
	})
}

type ExampleTestSuite struct {
	coretesting.TestCase
}

func (s *ExampleTestSuite) SetupSuite() {
	s.Config = exampleConfig(s.T())
	s.Config.SetupRoutes = func(app *fiber.App) {
		app.Get("/health", func(c fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"status": "ok",
			})
		})
	}
}

func (s *ExampleTestSuite) TestHealthEndpoint() {
	httpHelper := coretesting.NewHTTPTestHelper(&s.TestCase)
	assertHelper := coretesting.NewAssertHelper(&s.TestCase)

	resp, err := httpHelper.Get("/health")
	s.NoError(err)

	assertHelper.AssertOK(resp)
	assertHelper.AssertJSON(resp, "status", "ok")
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(ExampleTestSuite))
}

type RefreshDatabaseExampleSuite struct {
	coretesting.WithRefreshDatabase
}

func (s *RefreshDatabaseExampleSuite) SetupSuite() {
	s.Config = exampleConfig(s.T())
	s.Config.RefreshDatabase = true
}

func (s *RefreshDatabaseExampleSuite) TestDatabaseOperation() {
	dbHelper := coretesting.NewDatabaseHelper(&s.TestCase)

	dbHelper.AssertDatabaseCount("users", 0)
}

func TestRefreshDatabaseExampleSuite(t *testing.T) {
	suite.Run(t, new(RefreshDatabaseExampleSuite))
}
