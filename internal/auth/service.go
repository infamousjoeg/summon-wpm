package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/infamousjoeg/summon-wpm/internal/api"
	"github.com/infamousjoeg/summon-wpm/internal/config"
)

// AuthenticateWithClientCredentials performs non-interactive authentication using client credentials
func AuthenticateWithClientCredentials(cfg *config.Config, configFile string) error {
	// Create form data
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", cfg.ClientID)
	data.Set("client_secret", cfg.ClientSecret)

	// Create request
	req, err := http.NewRequest("POST",
		strings.TrimRight(cfg.TenantURL, "/")+TokenEndpoint,
		strings.NewReader(data.Encode()))

	if err != nil {
		return fmt.Errorf("error creating request: %s", err)
	}

	// Set proper headers for form data
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("token request failed: %s", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %s", resp.Status)
	}

	// Read response
	tokenResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %s", err)
	}

	// Parse token response
	var tokenResponse TokenResponse
	if err := json.Unmarshal(tokenResp, &tokenResponse); err != nil {
		return fmt.Errorf("error parsing token response: %s", err)
	}

	if tokenResponse.AccessToken == "" {
		return errors.New("no access token received")
	}

	// Save token to config
	cfg.AuthToken = tokenResponse.AccessToken
	expiryDuration := time.Duration(tokenResponse.ExpiresIn) * time.Second
	cfg.TokenExpiry = time.Now().Add(expiryDuration).Unix()

	return config.SaveConfig(cfg, configFile)
}

// GetAppCredentials retrieves application credentials from CyberArk Identity
func GetAppCredentials(cfg *config.Config, appID string) (string, error) {
	// Create the endpoint URL with query parameter
	endpoint := fmt.Sprintf("%s?appkey=%s", GetAppCredsEndpoint, url.QueryEscape(appID))

	// Enable verbose debugging
	fmt.Fprintf(os.Stderr, "Making request to: %s%s\n", cfg.TenantURL, endpoint)

	// Make request with empty body since we're using query parameters
	appCredResp, err := api.MakeAuthenticatedRequest(cfg, "POST", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("app credentials request failed: %s", err)
	}

	// Print full response for debugging
	fmt.Fprintf(os.Stderr, "Raw response: %s\n", string(appCredResp))

	// Parse the response
	var appCredResponse AppCredResponse
	if err := json.Unmarshal(appCredResp, &appCredResponse); err != nil {
		return "", fmt.Errorf("error parsing app cred response: %s", err)
	}

	// Check for errors
	if appCredResponse.Error != nil {
		return "", fmt.Errorf("get app credentials failed: %v", appCredResponse.Error)
	}

	// Check if Result contains data
	if appCredResponse.Result == nil || len(appCredResponse.Result) == 0 {
		return "", errors.New("empty result from API - credential not found or access denied")
	}

	// Extract the password from the Result map
	for _, possibleKey := range []string{"Password", "password", "secret", "value", "credential"} {
		if val, ok := appCredResponse.Result[possibleKey].(string); ok {
			return val, nil
		}
	}

	// If we can't find any password field, dump the contents for debugging
	resultBytes, _ := json.Marshal(appCredResponse.Result)
	return "", fmt.Errorf("password not found in result: %s", string(resultBytes))
}
