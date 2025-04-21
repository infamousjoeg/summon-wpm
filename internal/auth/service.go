package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/infamousjoeg/summon-wpm/internal/api"
	"github.com/infamousjoeg/summon-wpm/internal/config"
)

// AuthenticateWithClientCredentials performs non-interactive authentication using client credentials
func AuthenticateWithClientCredentials(cfg *config.Config, configFile string) error {
	tokenReq := TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
	}

	tokenBody, err := json.Marshal(tokenReq)
	if err != nil {
		return fmt.Errorf("error marshaling token request: %s", err)
	}

	tokenResp, err := api.MakeRequest(cfg, "POST", TokenEndpoint, bytes.NewBuffer(tokenBody))
	if err != nil {
		return fmt.Errorf("token request failed: %s", err)
	}

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
	appCredReq := AppCredRequest{
		AppID: appID,
	}

	appCredBody, err := json.Marshal(appCredReq)
	if err != nil {
		return "", fmt.Errorf("error marshaling app cred request: %s", err)
	}

	appCredResp, err := api.MakeAuthenticatedRequest(cfg, "POST", GetAppCredsEndpoint, bytes.NewBuffer(appCredBody))
	if err != nil {
		return "", fmt.Errorf("app credentials request failed: %s", err)
	}

	var appCredResponse AppCredResponse
	if err := json.Unmarshal(appCredResp, &appCredResponse); err != nil {
		return "", fmt.Errorf("error parsing app cred response: %s", err)
	}

	if !appCredResponse.Success {
		return "", fmt.Errorf("get app credentials failed: %s", appCredResponse.ErrorMsg)
	}

	// Return the password
	return appCredResponse.Password, nil
}
