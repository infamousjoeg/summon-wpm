package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/infamousjoeg/summon-wpm/internal/config"
)

func TestMakeRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check method
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept header application/json, got %s", r.Header.Get("Accept"))
		}

		// Check path
		if r.URL.Path != "/test-endpoint" {
			t.Errorf("Expected path /test-endpoint, got %s", r.URL.Path)
		}

		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		if string(body) != `{"test":"data"}` {
			t.Errorf("Expected body {\"test\":\"data\"}, got %s", string(body))
		}

		// Return response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	// Create test config
	cfg := &config.Config{
		TenantURL: server.URL,
	}

	// Make request
	response, err := MakeRequest(cfg, "POST", "/test-endpoint", strings.NewReader(`{"test":"data"}`))
	if err != nil {
		t.Fatalf("MakeRequest failed: %v", err)
	}

	// Verify response
	if string(response) != `{"success":true}` {
		t.Errorf("Expected response {\"success\":true}, got %s", string(response))
	}
}

func TestMakeAuthenticatedRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got %s", r.Header.Get("Authorization"))
		}

		// Return response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	// Create test config
	cfg := &config.Config{
		TenantURL: server.URL,
		AuthToken: "test-token",
	}

	// Make authenticated request
	response, err := MakeAuthenticatedRequest(cfg, "GET", "/authenticated", nil)
	if err != nil {
		t.Fatalf("MakeAuthenticatedRequest failed: %v", err)
	}

	// Verify response
	if string(response) != `{"success":true}` {
		t.Errorf("Expected response {\"success\":true}, got %s", string(response))
	}
}

func TestMakeAuthenticatedRequestUnauthorized(t *testing.T) {
	// Create a test server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer server.Close()

	// Create test config
	cfg := &config.Config{
		TenantURL: server.URL,
		AuthToken: "invalid-token",
	}

	// Make authenticated request
	_, err := MakeAuthenticatedRequest(cfg, "GET", "/authenticated", nil)

	// Verify error
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("Expected error to contain 'authentication failed', got %s", err.Error())
	}
}
