package api

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/infamousjoeg/summon-wpm/internal/config"
)

// MakeRequest makes a non-authenticated request to the CyberArk Identity API
func MakeRequest(cfg *config.Config, method, endpoint string, body io.Reader) ([]byte, error) {
	// Ensure baseURL doesn't end with slash
	baseURL := strings.TrimRight(cfg.TenantURL, "/")

	// Create request
	req, err := http.NewRequest(method, baseURL+endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("request failed with status: " + resp.Status)
	}

	// Read response body
	return io.ReadAll(resp.Body)
}

// MakeAuthenticatedRequest makes an authenticated request to the CyberArk Identity API
func MakeAuthenticatedRequest(cfg *config.Config, method, endpoint string, body io.Reader) ([]byte, error) {
	// Ensure baseURL doesn't end with slash
	baseURL := strings.TrimRight(cfg.TenantURL, "/")

	// Create request
	req, err := http.NewRequest(method, baseURL+endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.AuthToken)

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, errors.New("authentication failed: token expired or invalid")
		}
		return nil, errors.New("request failed with status: " + resp.Status)
	}

	// Read response body
	return io.ReadAll(resp.Body)
}
