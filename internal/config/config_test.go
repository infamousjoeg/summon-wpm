package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadConfig(t *testing.T) {
	// Create temp dir for test
	tmpDir, err := os.MkdirTemp("", "summon-wpm-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test config file path
	configFile := filepath.Join(tmpDir, defaultConfigFileName)

	// Test config
	cfg := &Config{
		TenantURL:    "https://example.cyberark.cloud",
		Username:     "testuser",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		AuthToken:    "test-token",
		TokenExpiry:  1234567890,
	}

	// Save config
	err = SaveConfig(cfg, configFile)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loadedCfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded config matches original
	if loadedCfg.TenantURL != cfg.TenantURL {
		t.Errorf("TenantURL mismatch: got %s, want %s", loadedCfg.TenantURL, cfg.TenantURL)
	}
	if loadedCfg.Username != cfg.Username {
		t.Errorf("Username mismatch: got %s, want %s", loadedCfg.Username, cfg.Username)
	}
	if loadedCfg.ClientID != cfg.ClientID {
		t.Errorf("ClientID mismatch: got %s, want %s", loadedCfg.ClientID, cfg.ClientID)
	}
	if loadedCfg.ClientSecret != cfg.ClientSecret {
		t.Errorf("ClientSecret mismatch: got %s, want %s", loadedCfg.ClientSecret, cfg.ClientSecret)
	}
	if loadedCfg.AuthToken != cfg.AuthToken {
		t.Errorf("AuthToken mismatch: got %s, want %s", loadedCfg.AuthToken, cfg.AuthToken)
	}
	if loadedCfg.TokenExpiry != cfg.TokenExpiry {
		t.Errorf("TokenExpiry mismatch: got %d, want %d", loadedCfg.TokenExpiry, cfg.TokenExpiry)
	}
}

func TestGetConfigFilePath(t *testing.T) {
	// Test with environment variable
	testDir := "/tmp/test-config-dir"
	origEnv := os.Getenv("SUMMON_WPM_CONFIG_DIR")
	defer os.Setenv("SUMMON_WPM_CONFIG_DIR", origEnv)

	os.Setenv("SUMMON_WPM_CONFIG_DIR", testDir)
	path := GetConfigFilePath()
	expected := filepath.Join(testDir, defaultConfigFileName)

	if path != expected {
		t.Errorf("Config path mismatch: got %s, want %s", path, expected)
	}
}

func TestMaskString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"password", "********"},
		{"test", "********"},
	}

	for _, tt := range tests {
		result := maskString(tt.input)
		if result != tt.expected {
			t.Errorf("maskString(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
