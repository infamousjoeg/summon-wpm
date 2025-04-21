package auth

import (
	"errors"
	"os"
	"time"

	"golang.org/x/term"

	"github.com/infamousjoeg/summon-wpm/internal/config"
)

const (
	StartAuthEndpoint   = "/Security/StartAuthentication"
	AdvanceAuthEndpoint = "/Security/AdvanceAuthentication"
	TokenEndpoint       = "/oauth2/token"
	GetAppCredsEndpoint = "/UserMgmt/GetAppCreds"
)

// StartAuthRequest represents the request body for starting authentication
type StartAuthRequest struct {
	User    string `json:"User"`
	Version string `json:"Version"`
}

// StartAuthResponse represents the response from start authentication
type StartAuthResponse struct {
	Success    bool        `json:"success"`
	Result     string      `json:"Result"`
	SessionID  string      `json:"SessionId"`
	Challenges []Challenge `json:"Challenges"`
	ErrorID    int         `json:"ErrorId"`
	ErrorMsg   string      `json:"ErrorMsg"`
}

// Challenge represents an authentication challenge
type Challenge struct {
	Mechanisms []Mechanism `json:"Mechanisms"`
	SessionID  string      `json:"SessionId"`
	TenantID   string      `json:"TenantId"`
}

// Mechanism represents an authentication mechanism
type Mechanism struct {
	MechanismID      string `json:"MechanismId"`
	Name             string `json:"Name"`
	PromptSelectMech string `json:"PromptSelectMech"`
}

// AdvanceAuthRequest represents the request body for advancing authentication
type AdvanceAuthRequest struct {
	SessionID   string `json:"SessionId"`
	MechanismID string `json:"MechanismId"`
	Answer      string `json:"Answer"`
}

// AdvanceAuthResponse represents the response from advance authentication
type AdvanceAuthResponse struct {
	Success  bool   `json:"success"`
	Result   string `json:"Result"`
	Token    string `json:"Token"`
	ErrorID  int    `json:"ErrorId"`
	ErrorMsg string `json:"ErrorMsg"`
}

// TokenRequest represents the request body for token endpoint
type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope,omitempty"`
}

// TokenResponse represents the response from token endpoint
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// AppCredRequest represents the request body for getting app credentials
type AppCredRequest struct {
	AppID string `json:"AppID"`
}

// AppCredResponse represents the response from getting app credentials
type AppCredResponse struct {
	Success  bool   `json:"success"`
	Result   string `json:"Result"`
	ErrorID  int    `json:"ErrorId"`
	ErrorMsg string `json:"ErrorMsg"`
	AppKey   string `json:"AppKey"`
	Username string `json:"Username"`
	Password string `json:"Password"`
}

// IsInteractive checks if the program is running in an interactive terminal
func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// Authenticate handles authentication to CyberArk Identity
func Authenticate(cfg *config.Config, configFile string, forceInteractive bool) error {
	if cfg.ClientID != "" && cfg.ClientSecret != "" && !forceInteractive {
		return AuthenticateWithClientCredentials(cfg, configFile)
	}

	if !IsInteractive() {
		return errors.New("cannot perform interactive authentication in non-interactive mode")
	}

	return AuthenticateInteractive(cfg, configFile)
}

// NeedsAuthentication checks if authentication is needed
func NeedsAuthentication(cfg *config.Config) bool {
	return cfg.AuthToken == "" || (cfg.TokenExpiry > 0 && time.Now().Unix() > cfg.TokenExpiry)
}
