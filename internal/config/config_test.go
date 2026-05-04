package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	if val := getEnv("TEST_KEY", "fallback"); val != "test_value" {
		t.Errorf("Expected test_value, got %s", val)
	}

	if val := getEnv("MISSING_KEY", "fallback"); val != "fallback" {
		t.Errorf("Expected fallback, got %s", val)
	}
}

func TestGetEnvAsInt(t *testing.T) {
	os.Setenv("TEST_INT", "123")
	defer os.Unsetenv("TEST_INT")

	if val := getEnvAsInt("TEST_INT", 0); val != 123 {
		t.Errorf("Expected 123, got %d", val)
	}

	if val := getEnvAsInt("MISSING_INT", 42); val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}

	os.Setenv("BAD_INT", "abc")
	defer os.Unsetenv("BAD_INT")
	if val := getEnvAsInt("BAD_INT", 99); val != 99 {
		t.Errorf("Expected 99, got %d", val)
	}
}
