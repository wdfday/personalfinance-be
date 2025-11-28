package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set some test environment variables
	err := os.Setenv("PORT", "9000")
	if err != nil {
		return
	}
	os.Setenv("DB_HOST", "test-host")
	defer os.Unsetenv("PORT")
	defer os.Unsetenv("DB_HOST")

	// Load configuration
	cfg := Load()

	// Test that environment variables are properly loaded
	if cfg.Server.Port != "9000" {
		t.Errorf("Expected PORT to be '9000', got '%s'", cfg.Server.Port)
	}

	if cfg.Database.Host != "test-host" {
		t.Errorf("Expected DB_HOST to be 'test-host', got '%s'", cfg.Database.Host)
	}

	// Test that defaults are applied
	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected default HOST to be 'localhost', got '%s'", cfg.Server.Host)
	}

	if cfg.Database.Port != 5434 {
		t.Errorf("Expected default DB_PORT to be 5432, got %d", cfg.Database.Port)
	}
}

func TestGetStringConfig(t *testing.T) {
	// Test with environment variable
	os.Setenv("TEST_VAR", "test-value")
	defer os.Unsetenv("TEST_VAR")

	value := GetStringConfig("TEST_VAR", "default-value")
	if value != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", value)
	}

	// Test with default value
	value = GetStringConfig("NONEXISTENT_VAR", "default-value")
	if value != "default-value" {
		t.Errorf("Expected 'default-value', got '%s'", value)
	}
}

func TestGetIntConfig(t *testing.T) {
	// Test with environment variable
	os.Setenv("TEST_INT", "123")
	defer os.Unsetenv("TEST_INT")

	value := GetIntConfig("TEST_INT", 456)
	if value != 123 {
		t.Errorf("Expected 123, got %d", value)
	}

	// Test with default value
	value = GetIntConfig("NONEXISTENT_INT", 456)
	if value != 456 {
		t.Errorf("Expected 456, got %d", value)
	}
}

func TestValidateConfig(t *testing.T) {
	// Test with missing required config
	err := ValidateConfig()
	if err == nil {
		t.Error("Expected validation error for missing JWT_SECRET")
	}

	// Set required config
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "test-user")
	os.Setenv("DB_PASSWORD", "test-password")
	os.Setenv("DB_NAME", "test-db")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
	}()

	err = ValidateConfig()
	if err != nil {
		t.Errorf("Expected no validation error, got: %v", err)
	}
}

func TestIsDevelopment(t *testing.T) {
	// Test development mode
	os.Setenv("GIN_MODE", "debug")
	defer os.Unsetenv("GIN_MODE")

	if !IsDevelopment() {
		t.Error("Expected IsDevelopment() to return true for debug mode")
	}

	// Test production mode
	os.Setenv("GIN_MODE", "release")
	if IsDevelopment() {
		t.Error("Expected IsDevelopment() to return false for release mode")
	}
}

func TestGetDatabaseURL(t *testing.T) {
	// Set database configuration
	os.Setenv("DB_HOST", "test-host")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_USER", "test-user")
	os.Setenv("DB_PASSWORD", "test-password")
	os.Setenv("DB_NAME", "test-db")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
	}()

	expected := "postgres://test-user:test-password@test-host:3306/test-db"
	url := GetDatabaseURL()
	if url != expected {
		t.Errorf("Expected '%s', got '%s'", expected, url)
	}
}
