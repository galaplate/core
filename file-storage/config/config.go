package config

import (
	"fmt"
	"strings"
)

// Config holds all filesystem configuration
type Config struct {
	Default string                  // Default driver ("local", "s3", "gcs")
	Drivers map[string]DriverConfig // Driver-specific configurations
}

// DriverConfig is the base configuration for any driver
type DriverConfig interface {
	Driver() string
	Validate() error
}

// New creates a new Config instance
func New() *Config {
	return &Config{
		Drivers: make(map[string]DriverConfig),
	}
}

// SetDefault sets the default driver
func (c *Config) SetDefault(driver string) *Config {
	c.Default = driver
	return c
}

// GetDriver returns driver configuration by name
func (c *Config) GetDriver(name string) (DriverConfig, error) {
	driver, exists := c.Drivers[name]
	if !exists {
		return nil, fmt.Errorf("driver not found: %s", name)
	}
	return driver, nil
}

// GetDefaultDriver returns the default driver configuration
func (c *Config) GetDefaultDriver() (DriverConfig, error) {
	if c.Default == "" {
		return nil, fmt.Errorf("no default driver configured")
	}
	return c.GetDriver(c.Default)
}

// Validate validates all driver configurations
func (c *Config) Validate() error {
	if c.Default == "" {
		return fmt.Errorf("no default driver configured")
	}

	if len(c.Drivers) == 0 {
		return fmt.Errorf("no drivers configured")
	}

	for name, driver := range c.Drivers {
		if c.Default == name {
			if err := driver.Validate(); err != nil {
				return fmt.Errorf("driver %s validation failed: %w", name, err)
			}
		}
	}

	// Verify default driver exists
	if _, exists := c.Drivers[c.Default]; !exists {
		return fmt.Errorf("default driver %s not configured", c.Default)
	}

	return nil
}

// String returns a string representation of the configuration
func (c *Config) String() string {
	var sb strings.Builder
	sb.WriteString("Filesystems Configuration:\n")
	fmt.Fprintf(&sb, "Default Driver: %s\n", c.Default)
	sb.WriteString("Configured Drivers:\n")

	for name, driver := range c.Drivers {
		fmt.Fprintf(&sb, "  - %s (%s)\n", name, driver.Driver())
	}

	return sb.String()
}
