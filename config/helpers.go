package config

// Config retrieves a configuration value using dot notation
// This is the main helper function to be used throughout the application
// Example: config.Config("database.connections.mysql.host")
func Config(key string) any {
	return GetGlobal().Get(key)
}

// ConfigString retrieves a string configuration value
// Example: config.ConfigString("database.driver")
func ConfigString(key string) string {
	return GetGlobal().GetString(key)
}

// ConfigInt retrieves an int configuration value
// Example: config.ConfigInt("database.port")
func ConfigInt(key string) int {
	return GetGlobal().GetInt(key)
}

// ConfigBool retrieves a bool configuration value
// Example: config.ConfigBool("database.strict")
func ConfigBool(key string) bool {
	return GetGlobal().GetBool(key)
}
