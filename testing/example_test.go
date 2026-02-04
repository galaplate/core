package testing_test

import (
	"testing"

	coretesting "github.com/galaplate/core/testing"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/suite"
)

type ExampleTestSuite struct {
	coretesting.TestCase
}

func (s *ExampleTestSuite) SetupSuite() {
	s.Config = coretesting.DefaultTestConfig()
	s.Config.SetupRoutes = func(app *fiber.App) {
		app.Get("/health", func(c *fiber.Ctx) error {
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
	s.Config = coretesting.DefaultTestConfig()
	s.Config.RefreshDatabase = true
}

func (s *RefreshDatabaseExampleSuite) TestDatabaseOperation() {
	dbHelper := coretesting.NewDatabaseHelper(&s.TestCase)

	dbHelper.AssertDatabaseCount("users", 0)
}

func TestRefreshDatabaseExampleSuite(t *testing.T) {
	suite.Run(t, new(RefreshDatabaseExampleSuite))
}
