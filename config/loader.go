package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/galaplate/core/env"
	"gopkg.in/yaml.v3"
)

// Loader loads configuration files from a directory
type Loader struct {
	configPath string
}

// NewLoader creates a new config loader
func NewLoader(configPath string) *Loader {
	return &Loader{
		configPath: configPath,
	}
}

// Load loads all configuration files from the config directory
func (l *Loader) Load() (map[string]any, error) {
	config := make(map[string]any)

	if _, err := os.Stat(l.configPath); os.IsNotExist(err) {
		return config, fmt.Errorf("config directory does not exist: %s", l.configPath)
	}

	files, err := os.ReadDir(l.configPath)
	if err != nil {
		return config, fmt.Errorf("failed to read config directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}

		filename := filepath.Join(l.configPath, file.Name())
		fileConfig, err := l.loadFile(filename)
		if err != nil {
			return config, fmt.Errorf("failed to load config file %s: %w", filename, err)
		}

		configName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		config[configName] = fileConfig
	}

	return config, nil
}

// loadFile loads a single YAML file and processes env variables
func (l *Loader) loadFile(filename string) (any, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	contentStr := string(content)
	contentStr = l.processEnvVariables(contentStr)

	var data any
	err = yaml.Unmarshal([]byte(contentStr), &data)
	if err != nil {
		return nil, err
	}

	return l.convertToProperTypes(data), nil
}

// processEnvVariables replaces ${VAR_NAME} or ${VAR_NAME:default} with env values
func (l *Loader) processEnvVariables(content string) string {
	// Pattern: ${ENV_VAR} or ${ENV_VAR:default_value}
	result := content
	start := 0

	for {
		idx := strings.Index(result[start:], "${")
		if idx == -1 {
			break
		}
		idx += start

		endIdx := strings.Index(result[idx:], "}")
		if endIdx == -1 {
			break
		}
		endIdx += idx

		varPart := result[idx+2 : endIdx]
		var varName, defaultValue string

		if before, after, ok := strings.Cut(varPart, ":"); ok {
			varName = before
			defaultValue = after
		} else {
			varName = varPart
		}

		value := env.Get(varName)
		if value == "" {
			value = defaultValue
		}

		result = result[:idx] + value + result[endIdx+1:]
		start = idx + len(value)
	}

	return result
}

// convertToProperTypes converts YAML any types to proper types
func (l *Loader) convertToProperTypes(data any) any {
	switch v := data.(type) {
	case map[any]any:
		result := make(map[string]any)
		for key, val := range v {
			keyStr := fmt.Sprintf("%v", key)
			result[keyStr] = l.convertToProperTypes(val)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = l.convertToProperTypes(val)
		}
		return result
	default:
		return v
	}
}
