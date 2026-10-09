package testing

import (
	"sync"
	"testing"

	"github.com/galaplate/core/config"
	"github.com/stretchr/testify/suite"
)

func TestMockConfigString(t *testing.T) {
	mgr := config.NewManager()
	mgr.Load(map[string]any{
		"app": map[string]any{
			"name": "original",
		},
	})

	config.InitializeGlobal(map[string]any{
		"app": map[string]any{
			"name": "original",
		},
	})

	tc := &TestCase{mockedConfig: make(map[string]any)}
	tc.MockConfig("app.name", "mocked")

	if got := config.ConfigString("app.name"); got != "mocked" {
		t.Errorf("expected 'mocked', got '%s'", got)
	}
}

func TestMockConfigInt(t *testing.T) {
	config.InitializeGlobal(map[string]any{
		"app": map[string]any{
			"port": 8080,
		},
	})

	tc := &TestCase{mockedConfig: make(map[string]any)}
	tc.MockConfig("app.port", 3000)

	if got := config.ConfigInt("app.port"); got != 3000 {
		t.Errorf("expected 3000, got %d", got)
	}
}

func TestMockConfigBool(t *testing.T) {
	config.InitializeGlobal(map[string]any{
		"app": map[string]any{
			"debug": false,
		},
	})

	tc := &TestCase{mockedConfig: make(map[string]any)}
	tc.MockConfig("app.debug", true)

	if got := config.ConfigBool("app.debug"); !got {
		t.Error("expected true")
	}
}

func TestMockConfigNested(t *testing.T) {
	config.InitializeGlobal(map[string]any{
		"database": map[string]any{
			"default": "mysql",
			"connections": map[string]any{
				"sqlite": map[string]any{
					"database": ":memory:",
				},
			},
		},
	})

	tc := &TestCase{mockedConfig: make(map[string]any)}
	tc.MockConfig("database.default", "sqlite")
	tc.MockConfig("database.connections.sqlite.database", "/tmp/test.db")

	if got := config.ConfigString("database.default"); got != "sqlite" {
		t.Errorf("expected 'sqlite', got '%s'", got)
	}
	if got := config.ConfigString("database.connections.sqlite.database"); got != "/tmp/test.db" {
		t.Errorf("expected '/tmp/test.db', got '%s'", got)
	}
}

func TestMockConfigMap(t *testing.T) {
	config.InitializeGlobal(map[string]any{
		"app": map[string]any{
			"name":  "original",
			"debug": false,
			"port":  8080,
		},
	})

	tc := &TestCase{mockedConfig: make(map[string]any)}
	tc.MockConfigMap(map[string]any{
		"app.name":  "mapped",
		"app.debug": true,
		"app.port":  9000,
	})

	if got := config.ConfigString("app.name"); got != "mapped" {
		t.Errorf("expected 'mapped', got '%s'", got)
	}
	if got := config.ConfigBool("app.debug"); !got {
		t.Error("expected true")
	}
	if got := config.ConfigInt("app.port"); got != 9000 {
		t.Errorf("expected 9000, got %d", got)
	}
}

func TestMockConfigReset(t *testing.T) {
	config.InitializeGlobal(map[string]any{
		"app": map[string]any{
			"debug": false,
		},
	})

	tc := &TestCase{mockedConfig: make(map[string]any)}
	tc.MockConfig("app.debug", true)

	if !config.ConfigBool("app.debug") {
		t.Error("expected true before reset")
	}

	tc.resetMockedConfig()

	if config.ConfigBool("app.debug") {
		t.Error("expected false after reset")
	}
}

func TestMockConfigMultipleSets(t *testing.T) {
	config.InitializeGlobal(map[string]any{
		"app": map[string]any{
			"name": "original",
		},
	})

	tc := &TestCase{mockedConfig: make(map[string]any)}
	tc.MockConfig("app.name", "first")
	tc.MockConfig("app.name", "second")

	if got := config.ConfigString("app.name"); got != "second" {
		t.Errorf("expected 'second', got '%s'", got)
	}
}

type ConfigMockSuite struct {
	suite.Suite
}

func (s *ConfigMockSuite) TestMockConfigString() {
	config.InitializeGlobal(map[string]any{
		"app": map[string]any{
			"name": "original",
		},
	})

	tc := &TestCase{mockedConfig: make(map[string]any)}
	tc.MockConfig("app.name", "test-app")

	s.Equal("test-app", config.ConfigString("app.name"))
}

func (s *ConfigMockSuite) TestMockConfigNested() {
	config.InitializeGlobal(map[string]any{
		"database": map[string]any{
			"default": "mysql",
			"connections": map[string]any{
				"sqlite": map[string]any{
					"database": ":memory:",
				},
			},
		},
	})

	tc := &TestCase{mockedConfig: make(map[string]any)}
	tc.MockConfig("database.default", "sqlite")
	tc.MockConfig("database.connections.sqlite.database", "/tmp/test.db")

	s.Equal("sqlite", config.ConfigString("database.default"))
	s.Equal("/tmp/test.db", config.ConfigString("database.connections.sqlite.database"))
}

func (s *ConfigMockSuite) TestMockConfigReset() {
	config.InitializeGlobal(map[string]any{
		"app": map[string]any{
			"debug": false,
		},
	})

	tc := &TestCase{mockedConfig: make(map[string]any)}
	tc.MockConfig("app.debug", true)
	s.True(config.ConfigBool("app.debug"))

	tc.resetMockedConfig()
	s.False(config.ConfigBool("app.debug"))
}

func TestConfigMockSuite(t *testing.T) {
	suite.Run(t, new(ConfigMockSuite))
}

func TestMockConfigConcurrent(t *testing.T) {
	config.InitializeGlobal(map[string]any{
		"counter": 0,
	})

	tc := &TestCase{mockedConfig: make(map[string]any)}

	var wg sync.WaitGroup
	results := make([]int, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tc.MockConfig("counter", idx)
			results[idx] = config.ConfigInt("counter")
		}(i)
	}

	wg.Wait()

	for _, v := range results {
		if v < 0 || v >= 10 {
			t.Errorf("value %d out of expected range [0, 9]", v)
		}
	}
}
