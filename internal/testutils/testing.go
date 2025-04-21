package testutils

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/infamousjoeg/summon-wpm/internal/config"
)

// MockServer represents a mock HTTP server for testing
type MockServer struct {
	Server *httptest.Server
	Calls  []MockServerCall
}

// MockServerCall represents a recorded call to the mock server
type MockServerCall struct {
	Method  string
	Path    string
	Body    string
	Headers http.Header
}

// NewMockServer creates a new mock server that records all requests
func NewMockServer(t *testing.T, handler http.HandlerFunc) *MockServer {
	mock := &MockServer{
		Calls: []MockServerCall{},
	}

	// Create a wrapper handler that records calls
	wrapperHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Record the call
		body, err := io.ReadAll(r.Body)
		if err == nil {
			r.Body.Close()
			// Recreate the body for the original handler
			r.Body = io.NopCloser(bytes.NewReader(body))
		}

		mock.Calls = append(mock.Calls, MockServerCall{
			Method:  r.Method,
			Path:    r.URL.Path,
			Body:    string(body),
			Headers: r.Header,
		})

		// Call the original handler
		handler(w, r)
	})

	// Create the server
	mock.Server = httptest.NewServer(wrapperHandler)

	// Cleanup when the test finishes
	t.Cleanup(func() {
		mock.Server.Close()
	})

	return mock
}

// TempConfig creates a temporary config file for testing
func TempConfig(t *testing.T) (string, func()) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "summon-wpm-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create config file path
	configFile := filepath.Join(tmpDir, "test-config.json")

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return configFile, cleanup
}

// CreateTestConfig creates and saves a test config
func CreateTestConfig(t *testing.T, configFile string, cfg *config.Config) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configFile), 0700); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Save config
	if err := config.SaveConfig(cfg, configFile); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
}

// MockConfigFilePath mocks the config.GetConfigFilePath function to return a custom path
func MockConfigFilePath(t *testing.T, path string) func() {
	origGetConfigFilePath := config.GetConfigFilePath
	config.GetConfigFilePath = func() string {
		return path
	}

	return func() {
		config.GetConfigFilePath = origGetConfigFilePath
	}
}
