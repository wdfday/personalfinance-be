package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// GetConfigValue returns a configuration value by key with optional default
func GetConfigValue(key string, defaultValue ...interface{}) interface{} {
	if viper.IsSet(key) {
		return viper.Get(key)
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return nil
}

// GetStringConfig returns a string configuration value
func GetStringConfig(key string, defaultValue ...string) string {
	if viper.IsSet(key) {
		return viper.GetString(key)
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// GetIntConfig returns an integer configuration value
func GetIntConfig(key string, defaultValue ...int) int {
	if viper.IsSet(key) {
		return viper.GetInt(key)
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetBoolConfig returns a boolean configuration value
func GetBoolConfig(key string, defaultValue ...bool) bool {
	if viper.IsSet(key) {
		return viper.GetBool(key)
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return false
}

// GetStringSliceConfig returns a string slice configuration value
func GetStringSliceConfig(key string, defaultValue ...[]string) []string {
	if viper.IsSet(key) {
		return viper.GetStringSlice(key)
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return []string{}
}

// ValidateConfig validates required configuration values
func ValidateConfig() error {
	requiredKeys := []string{
		"JWT_SECRET",
		"DB_HOST",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
	}

	var missingKeys []string
	for _, key := range requiredKeys {
		if !viper.IsSet(key) || viper.GetString(key) == "" {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) > 0 {
		return fmt.Errorf("missing required configuration keys: %s", strings.Join(missingKeys, ", "))
	}

	return nil
}

// PrintConfig prints current configuration (excluding sensitive data)
func PrintConfig() {
	log.Println("=== Configuration ===")

	// Server config
	log.Printf("Server: %s:%s", GetStringConfig("HOST"), GetStringConfig("PORT"))
	log.Printf("Gin Mode: %s", GetStringConfig("GIN_MODE"))

	// Database config
	log.Printf("Database: %s:%d", GetStringConfig("DB_HOST"), GetIntConfig("DB_PORT"))
	log.Printf("Database Name: %s", GetStringConfig("DB_NAME"))
	log.Printf("Database User: %s", GetStringConfig("DB_USER"))

	// Email config
	log.Printf("SMTP Host: %s:%d", GetStringConfig("SMTP_HOST"), GetIntConfig("SMTP_PORT"))
	log.Printf("From Email: %s", GetStringConfig("FROM_EMAIL"))

	// CORS config
	corsOrigins := GetStringSliceConfig("CORS_ORIGINS")
	log.Printf("CORS Origins: %v", corsOrigins)

	// Redis config
	log.Printf("Redis URL: %s", GetStringConfig("REDIS_URL"))

	// Logging config
	log.Printf("Log Level: %s", GetStringConfig("LOG_LEVEL"))
	log.Printf("Log Format: %s", GetStringConfig("LOG_FORMAT"))

	log.Println("=====================")
}

// IsDevelopment returns true if running in development mode
func IsDevelopment() bool {
	return GetStringConfig("GIN_MODE") == "debug"
}

// IsProduction returns true if running in production mode
func IsProduction() bool {
	return GetStringConfig("GIN_MODE") == "release"
}

// GetDatabaseURL returns the complete database URL
func GetDatabaseURL() string {
	if url := GetStringConfig("DATABASE_URL"); url != "" {
		return url
	}

	// Build URL from components
	host := GetStringConfig("DB_HOST", "localhost")
	port := GetIntConfig("DB_PORT", 5432)
	user := GetStringConfig("DB_USER", "finance_user")
	password := GetStringConfig("DB_PASSWORD", "finance_password")
	name := GetStringConfig("DB_NAME", "finance_dss")

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, name)
}
