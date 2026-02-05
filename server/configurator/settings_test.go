package configurator

import (
	"os"
	"testing"
)

func TestLoadSettings_RedisURL(t *testing.T) {
	// Test default value
	os.Unsetenv("REDIS_URL")
	settings := LoadSettings()
	if settings.RedisURL != "redis://localhost:6379/0" {
		t.Errorf("expected default redis URL, got %s", settings.RedisURL)
	}

	// Test custom value
	os.Setenv("REDIS_URL", "redis://custom:6380/1")
	defer os.Unsetenv("REDIS_URL")
	settings = LoadSettings()
	if settings.RedisURL != "redis://custom:6380/1" {
		t.Errorf("expected custom redis URL, got %s", settings.RedisURL)
	}
}
