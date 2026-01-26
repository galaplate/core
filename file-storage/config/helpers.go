package config

import (
	"os"

	"github.com/galaplate/core/config"
)

// DefaultAllowedTypes returns the default list of allowed MIME types
func DefaultAllowedTypes() []string {
	return []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}
}

// DefaultMaxSize returns the default maximum file size (10MB)
func DefaultMaxSize() int64 {
	return 10 * 1024 * 1024
}

// getEnvValue is a helper that tries to get from Galaplate config first, then falls back to os.Getenv
func getEnvValue(configPath, envVar string) string {
	val := tryGetConfigValue(configPath)
	if val != "" {
		return val
	}

	return os.Getenv(envVar)
}

func tryGetConfigValue(path string) string {
	return config.ConfigString(path)
}
