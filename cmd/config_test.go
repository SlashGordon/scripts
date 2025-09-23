package cmd

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	// Test with existing env var
	os.Setenv("TEST_VAR", "test_value")
	result := getEnv("TEST_VAR", "default")
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", result)
	}
	
	// Test with default value
	result = getEnv("NON_EXISTENT_VAR", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}
	
	os.Unsetenv("TEST_VAR")
}

func TestGetDDNSConfig(t *testing.T) {
	os.Setenv("CF_API_TOKEN", "test_token")
	os.Setenv("CF_ZONE_ID", "test_zone")
	os.Setenv("CF_RECORD_NAME", "test.example.com")
	
	config := getDDNSConfig()
	
	if config.APIToken != "test_token" {
		t.Errorf("Expected 'test_token', got '%s'", config.APIToken)
	}
	if config.ZoneID != "test_zone" {
		t.Errorf("Expected 'test_zone', got '%s'", config.ZoneID)
	}
	if config.RecordName != "test.example.com" {
		t.Errorf("Expected 'test.example.com', got '%s'", config.RecordName)
	}
	
	os.Unsetenv("CF_API_TOKEN")
	os.Unsetenv("CF_ZONE_ID")
	os.Unsetenv("CF_RECORD_NAME")
}

func TestGetAcmeConfig(t *testing.T) {
	os.Setenv("CF_API_TOKEN", "test_token")
	os.Setenv("ACME_DOMAIN", "test.example.com")
	os.Setenv("ACME_EMAIL", "test@example.com")
	
	config := getAcmeConfig()
	
	if config.CFToken != "test_token" {
		t.Errorf("Expected 'test_token', got '%s'", config.CFToken)
	}
	if config.Domain != "test.example.com" {
		t.Errorf("Expected 'test.example.com', got '%s'", config.Domain)
	}
	if config.Email != "test@example.com" {
		t.Errorf("Expected 'test@example.com', got '%s'", config.Email)
	}
	
	os.Unsetenv("CF_API_TOKEN")
	os.Unsetenv("ACME_DOMAIN")
	os.Unsetenv("ACME_EMAIL")
}