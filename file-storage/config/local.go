package config

import "fmt"

// LocalConfig holds local filesystem configuration
type LocalConfig struct {
	Path         string   // Directory path for uploads
	MaxSize      int64    // Maximum file size in bytes
	AllowedTypes []string // Allowed MIME types
}

// Driver returns the driver name
func (lc *LocalConfig) Driver() string {
	return "local"
}

// Validate validates local configuration
func (lc *LocalConfig) Validate() error {
	if lc.Path == "" {
		return fmt.Errorf("local driver: path is required")
	}
	if lc.MaxSize <= 0 {
		return fmt.Errorf("local driver: max_size must be greater than 0")
	}
	if len(lc.AllowedTypes) == 0 {
		return fmt.Errorf("local driver: allowed_types cannot be empty")
	}
	return nil
}

// WithLocalDriver adds a local driver configuration
func (c *Config) WithLocalDriver(path string, maxSize int64, allowedTypes []string) *Config {
	if maxSize == 0 {
		maxSize = DefaultMaxSize()
	}
	if len(allowedTypes) == 0 {
		allowedTypes = DefaultAllowedTypes()
	}
	c.Drivers["local"] = &LocalConfig{
		Path:         path,
		MaxSize:      maxSize,
		AllowedTypes: allowedTypes,
	}
	return c
}
