package config_test

import (
	"http/v1/dev/internal/config"
	"testing"
)

func TestLoadWithEnvs(t *testing.T) {
	mockEnvs := map[string]string{
		"DB_HOST":     "localhost",
		"DB_PORT":     "5424",
		"DB_DATABASE": "http",
		"DB_USER":     "test",
		"DB_PASSWORD": "test",
	}
	expectedDbString := "postgres://test:test@localhost:5424/http?sslmode=disable"

	for key, value := range mockEnvs {
		t.Setenv(key, value)
	}

	c, err := config.Load()
	if err != nil {
		t.Fatalf("error when mocking the load: %q", err)
	}

	if c.DatabaseURL != expectedDbString {
		t.Fatalf("error when parsing the db url. Expected %q, got %q", expectedDbString, c.DatabaseURL)
	}
}

func TestLoadWithoutEnv(t *testing.T) {
	mockEnvs := map[string]string{
		"DB_HOST":     "localhost",
		"DB_PORT":     "5424",
		"DB_DATABASE": "http",
		"DB_USER":     "test",
		"DB_PASSWORD": "",
	}

	for key, value := range mockEnvs {
		t.Setenv(key, value)
	}

	_, err := config.Load()
	if err == nil {
		t.Fatalf("Expected error for DB_PASSWORD, got %v", nil)
	}
}
