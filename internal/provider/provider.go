package provider

import (
	"fmt"
	"os"
	"strings"

	"github.com/infamousjoeg/summon-wpm/internal/auth"
	"github.com/infamousjoeg/summon-wpm/internal/config"
)

// Provider represents the Summon provider for CyberArk Identity
type Provider struct {
	verbose bool
}

// NewProvider creates a new provider instance
func NewProvider(verbose bool) *Provider {
	return &Provider{
		verbose: verbose,
	}
}

// GetCredential retrieves a credential from CyberArk Identity
func (p *Provider) GetCredential(appID string) (string, error) {
	configFile := config.GetConfigFilePath()

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no configuration found. Run with --config to set up")
		}
		return "", fmt.Errorf("error loading config: %s", err)
	}

	// Check if we need to authenticate or refresh token
	needAuth := auth.NeedsAuthentication(cfg)
	interactive := auth.IsInteractive()

	if needAuth {
		if p.verbose {
			fmt.Fprintln(os.Stderr, "Authentication required, authenticating...")
		}

		if cfg.ClientID != "" && cfg.ClientSecret != "" {
			// Non-interactive service user auth
			if err := auth.AuthenticateWithClientCredentials(cfg, configFile); err != nil {
				if p.verbose {
					fmt.Fprintf(os.Stderr, "Service user authentication failed: %s\n", err)
				}

				if interactive {
					// Fallback to interactive if running in terminal
					if err := auth.AuthenticateInteractive(cfg, configFile); err != nil {
						return "", fmt.Errorf("interactive authentication failed: %s", err)
					}
				} else {
					return "", fmt.Errorf("service user authentication failed: %s", err)
				}
			}
		} else if interactive {
			// Interactive user auth
			if err := auth.AuthenticateInteractive(cfg, configFile); err != nil {
				return "", fmt.Errorf("authentication failed: %s", err)
			}
		} else {
			return "", fmt.Errorf("authentication required but running in non-interactive mode with no service credentials")
		}
	}

	// Get app credentials
	credential, err := auth.GetAppCredentials(cfg, appID)
	if err != nil {
		// If we get an auth error, try to re-authenticate once
		if strings.Contains(err.Error(), "authentication") || strings.Contains(err.Error(), "401") {
			if p.verbose {
				fmt.Fprintln(os.Stderr, "Authentication token expired or invalid, re-authenticating...")
			}

			if cfg.ClientID != "" && cfg.ClientSecret != "" {
				if err := auth.AuthenticateWithClientCredentials(cfg, configFile); err != nil {
					return "", fmt.Errorf("re-authentication failed: %s", err)
				}
			} else if interactive {
				if err := auth.AuthenticateInteractive(cfg, configFile); err != nil {
					return "", fmt.Errorf("re-authentication failed: %s", err)
				}
			} else {
				return "", fmt.Errorf("re-authentication required but running in non-interactive mode")
			}

			// Try again with new token
			return auth.GetAppCredentials(cfg, appID)
		}

		return "", err
	}

	return credential, nil
}
