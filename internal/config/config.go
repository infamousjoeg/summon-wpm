package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const defaultConfigFileName = "cyberark-wpm.json"

// Config stores the configuration for the provider
type Config struct {
	TenantURL    string `json:"tenant_url"`
	Username     string `json:"username"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	AuthToken    string `json:"auth_token,omitempty"`
	TokenExpiry  int64  `json:"token_expiry,omitempty"`
}

// GetConfigFilePathFunc defines the function signature for getting config file path
type GetConfigFilePathFunc func() string

// getConfigFilePathImpl is the actual implementation
func getConfigFilePathImpl() string {
	// Determine config directory based on OS
	var configDir string

	if os.Getenv("SUMMON_WPM_CONFIG_DIR") != "" {
		configDir = os.Getenv("SUMMON_WPM_CONFIG_DIR")
	} else if runtime.GOOS == "windows" {
		configDir = filepath.Join(os.Getenv("APPDATA"), "summon-wpm")
	} else {
		// Unix-like systems
		if os.Getenv("XDG_CONFIG_HOME") != "" {
			configDir = filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "summon-wpm")
		} else {
			configDir = filepath.Join(os.Getenv("HOME"), ".config", "summon-wpm")
		}
	}

	return filepath.Join(configDir, defaultConfigFileName)
}

// GetConfigFilePath is the function variable that can be replaced in tests
var GetConfigFilePath GetConfigFilePathFunc = getConfigFilePathImpl

// RunConfigWizard runs the configuration wizard
func RunConfigWizard(configFile string) {
	fmt.Println("CyberArk Workload Password Management (WPM) Configuration")
	fmt.Println("===================================================")

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configFile)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create config directory: %s\n", err)
		os.Exit(1)
	}

	// Load existing config if possible
	config := &Config{}
	existingConfig, err := LoadConfig(configFile)
	if err == nil {
		config = existingConfig
		fmt.Println("Loaded existing configuration. Press Enter to keep current values.")
	}

	reader := bufio.NewReader(os.Stdin)

	// Get tenant URL
	fmt.Printf("Tenant URL [%s]: ", config.TenantURL)
	tenantURL, _ := reader.ReadString('\n')
	tenantURL = strings.TrimSpace(tenantURL)
	if tenantURL != "" {
		config.TenantURL = tenantURL
	}

	// Get username
	fmt.Printf("Username [%s]: ", config.Username)
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username != "" {
		config.Username = username
	}

	// Ask if they want to use service account
	fmt.Print("Do you want to configure a service account (client credentials)? (y/n): ")
	useService, _ := reader.ReadString('\n')
	useService = strings.TrimSpace(strings.ToLower(useService))

	if useService == "y" || useService == "yes" {
		fmt.Printf("Client ID [%s]: ", config.ClientID)
		clientID, _ := reader.ReadString('\n')
		clientID = strings.TrimSpace(clientID)
		if clientID != "" {
			config.ClientID = clientID
		}

		fmt.Printf("Client Secret [%s]: ", maskString(config.ClientSecret))
		clientSecret, _ := reader.ReadString('\n')
		clientSecret = strings.TrimSpace(clientSecret)
		if clientSecret != "" {
			config.ClientSecret = clientSecret
		}
	}

	// Save config
	if err := SaveConfig(config, configFile); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save config: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration saved to:", configFile)
	fmt.Println("Run with --login to authenticate now")
}

// maskString masks a string with asterisks
func maskString(s string) string {
	if s == "" {
		return ""
	}
	return "********"
}

// LoadConfig loads the configuration from the config file
func LoadConfig(configFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("invalid config file format: %s", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(config *Config, configFile string) error {
	// Ensure directory exists
	configDir := filepath.Dir(configFile)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0600)
}
