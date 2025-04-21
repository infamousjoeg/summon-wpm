package provider

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/infamousjoeg/summon-wpm/internal/config"
)

func TestGetCredential(t *testing.T) {
	// Create temp dir for test config
	tmpDir, err := os.MkdirTemp("", "summon-wpm-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configFile := filepath.Join(tmpDir, "test-config.json")

	// Save original function and restore after test
	origGetConfigFilePath := config.GetConfigFilePath
	defer func() {
		config.GetConfigFilePath = origGetConfigFilePath
	}()

	// Mock the function to return our test path
	config.GetConfigFilePath = func() string {
		return configFile
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			// Client credentials request
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "new-test-token",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))
		case "/UserMgmt/GetAppCreds":
			// App credentials request
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"success": true,
				"Result": "Success",
				"AppKey": "app-key",
				"Username": "app-username",
				"Password": "test-credential"
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create and save test config
	cfg := &config.Config{
		TenantURL:    server.URL,
		Username:     "test-user",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		AuthToken:    "valid-token",
		TokenExpiry:  time.Now().Add(1 * time.Hour).Unix(),
	}
	if err := config.SaveConfig(cfg, configFile); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Create provider
	p := NewProvider(true)

	// Test with valid token
	credential, err := p.GetCredential("test-app-id")
	if err != nil {
		t.Fatalf("GetCredential failed: %v", err)
	}
	if credential != "test-credential" {
		t.Errorf("Expected credential 'test-credential', got %s", credential)
	}

	// Test with expired token
	cfg.TokenExpiry = time.Now().Add(-1 * time.Hour).Unix()
	if err := config.SaveConfig(cfg, configFile); err != nil {
		t.Fatalf("Failed to save config with expired token: %v", err)
	}

	// Should re-authenticate and still work
	credential, err = p.GetCredential("test-app-id")
	if err != nil {
		t.Fatalf("GetCredential with expired token failed: %v", err)
	}
	if credential != "test-credential" {
		t.Errorf("Expected credential 'test-credential', got %s", credential)
	}

	// Verify token was refreshed in the config
	updatedCfg, err := config.LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load updated config: %v", err)
	}
	if updatedCfg.AuthToken != "new-test-token" {
		t.Errorf("Expected AuthToken to be refreshed to 'new-test-token', got %s", updatedCfg.AuthToken)
	}
}

func TestGetCredentialWithNoConfig(t *testing.T) {
	// Save original function and restore after test
	origGetConfigFilePath := config.GetConfigFilePath
	defer func() {
		config.GetConfigFilePath = origGetConfigFilePath
	}()

	// Mock the function to return a non-existent path
	config.GetConfigFilePath = func() string {
		return "/non/existent/path/config.json"
	}

	p := NewProvider(false)
	_, err := p.GetCredential("test-app-id")
	if err == nil {
		t.Fatal("Expected error for non-existent config, got nil")
	}
	if os.IsNotExist(err) {
		t.Errorf("Expected non-existence error message, got: %v", err)
	}
}
