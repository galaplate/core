package config

import (
	"fmt"
	"strings"
	"sync"
)

// Manager manages application configuration
type Manager struct {
	config map[string]any
	mu     sync.RWMutex
}

// NewManager creates a new config manager
func NewManager() *Manager {
	return &Manager{
		config: make(map[string]any),
	}
}

// Load loads configuration from a nested map
func (m *Manager) Load(data map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = data
}

// Set sets a configuration value using dot notation
// Example: Set("database.default", "mysql")
func (m *Manager) Set(key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setNested(m.config, key, value)
}

// Get retrieves a configuration value using dot notation
// Example: Get("database.connections.mysql.host")
// Returns nil if key doesn't exist
func (m *Manager) Get(key string) any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getNested(m.config, key)
}

// GetString retrieves a string configuration value
func (m *Manager) GetString(key string) string {
	value := m.Get(key)
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", value)
}

// GetInt retrieves an int configuration value
func (m *Manager) GetInt(key string) int {
	value := m.Get(key)
	if value == nil {
		return 0
	}
	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		var result int
		fmt.Sscanf(v, "%d", &result)
		return result
	}
	return 0
}

// GetBool retrieves a bool configuration value
func (m *Manager) GetBool(key string) bool {
	value := m.Get(key)
	if value == nil {
		return false
	}
	if b, ok := value.(bool); ok {
		return b
	}
	return false
}

// GetAll returns all configuration
func (m *Manager) GetAll() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Has checks if a configuration key exists
func (m *Manager) Has(key string) bool {
	return m.Get(key) != nil
}

// getNested retrieves a value from nested map using dot notation
func (m *Manager) getNested(data map[string]any, key string) any {
	if key == "" {
		return nil
	}

	parts := strings.Split(key, ".")
	var current any = data

	for _, part := range parts {
		switch c := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = c[part]
			if !ok {
				return nil
			}
		default:
			return nil
		}
	}

	return current
}

// setNested sets a value in nested map using dot notation
func (m *Manager) setNested(data map[string]any, key string, value any) {
	if key == "" {
		return
	}

	parts := strings.Split(key, ".")
	current := data

	for i := range len(parts) - 1 {
		part := parts[i]
		if _, ok := current[part]; !ok {
			current[part] = make(map[string]any)
		}

		switch c := current[part].(type) {
		case map[string]any:
			current = c
		default:
			newMap := make(map[string]any)
			current[part] = newMap
			current = newMap
		}
	}

	current[parts[len(parts)-1]] = value
}

// globalConfigManager is the global config manager instance
var globalConfigManager *Manager
var configMutex sync.Once

// InitializeGlobal initializes the global config manager with data
func InitializeGlobal(data map[string]any) {
	configMutex.Do(func() {
		globalConfigManager = NewManager()
		globalConfigManager.Load(data)
	})
}

// GetGlobal returns the global config manager
func GetGlobal() *Manager {
	if globalConfigManager == nil {
		globalConfigManager = NewManager()
	}
	return globalConfigManager
}
