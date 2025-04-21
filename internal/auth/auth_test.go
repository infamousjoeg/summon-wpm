package auth

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/infamousjoeg/summon-wpm/internal/config"
)

// Mock HTTP server for testing
func setupMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	t.Cleanup(func() {
		server.Close()
	})
	return server
}

func TestNeedsAuthentication(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.Config
		expected bool
	}{
		{
			name: "No token",
			cfg: config.Config{
				AuthToken: "",
			},
			expected: true,
		},
		{
			name: "Token expired",
			cfg: config.Config{
				AuthToken:   "test-token",
				TokenExpiry: time.Now().Add(-1 * time.Hour).Unix(),
			},
			expected: true,
		},
		{
			name: "Token valid",
			cfg: config.Config{
				AuthToken:   "test-token",
				TokenExpiry: time.Now().Add(1 * time.Hour).Unix(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NeedsAuthentication(&tt.cfg)
			if result != tt.expected {
				t.Errorf("NeedsAuthentication() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAuthenticateWithClientCredentials(t *testing.T) {
	// Create temp dir for test config
	tmpDir, err := os.MkdirTemp("", "summon-wpm-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configFile := filepath.Join(tmpDir, "test-config.json")

	// Setup mock server
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Check it's a token request
		if r.URL.Path != TokenEndpoint {
			t.Errorf("Expected request to %s, got %s", TokenEndpoint, r.URL.Path)
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		// Validate body contains expected client credentials
		if !bytes.Contains(body, []byte(`"client_id":"test-client-id"`)) {
			t.Errorf("Request body missing client_id, got: %s", string(body))
		}
		if !bytes.Contains(body, []byte(`"client_secret":"test-client-secret"`)) {
			t.Errorf("Request body missing client_secret, got: %s", string(body))
		}

		// Return successful token response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"access_token": "mock-access-token",
			"token_type": "Bearer",
			"expires_in": 3600
		}`))
	})

	// Create test config
	cfg := &config.Config{
		TenantURL:    server.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	// Call the function
	err = AuthenticateWithClientCredentials(cfg, configFile)
	if err != nil {
		t.Fatalf("AuthenticateWithClientCredentials failed: %v", err)
	}

	// Verify token was set in config
	if cfg.AuthToken != "mock-access-token" {
		t.Errorf("Expected AuthToken to be 'mock-access-token', got %s", cfg.AuthToken)
	}

	// Verify config was saved (by loading it)
	savedCfg, err := config.LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}
	if savedCfg.AuthToken != "mock-access-token" {
		t.Errorf("Saved AuthToken = %s, want 'mock-access-token'", savedCfg.AuthToken)
	}
}

func TestGetAppCredentials(t *testing.T) {
	// Setup mock server
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Check it's an app creds request
		if r.URL.Path != GetAppCredsEndpoint {
			t.Errorf("Expected request to %s, got %s", GetAppCredsEndpoint, r.URL.Path)
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		// Check auth header
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected auth header 'Bearer test-token', got %s", r.Header.Get("Authorization"))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		// Validate AppID is in request
		if !bytes.Contains(body, []byte(`"AppID":"test-app-id"`)) {
			t.Errorf("Request body missing AppID, got: %s", string(body))
		}

		// Return successful app creds response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"success": true,
			"Result": "Success",
			"AppKey": "app-key",
			"Username": "app-username",
			"Password": "app-password"
		}`))
	})

	// Create test config
	cfg := &config.Config{
		TenantURL: server.URL,
		AuthToken: "test-token",
	}

	// Call the function
	password, err := GetAppCredentials(cfg, "test-app-id")
	if err != nil {
		t.Fatalf("GetAppCredentials failed: %v", err)
	}

	// Verify password was returned
	if password != "app-password" {
		t.Errorf("Expected password to be 'app-password', got %s", password)
	}
}
