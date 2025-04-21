package auth

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/infamousjoeg/summon-wpm/internal/api"
	"github.com/infamousjoeg/summon-wpm/internal/config"
)

// AuthenticateInteractive performs interactive authentication with user input
func AuthenticateInteractive(cfg *config.Config, configFile string) error {
	// Start authentication
	startAuthReq := StartAuthRequest{
		User:    cfg.Username,
		Version: "1.0",
	}

	startAuthBody, err := json.Marshal(startAuthReq)
	if err != nil {
		return fmt.Errorf("error marshaling start auth request: %s", err)
	}

	startAuthResp, err := api.MakeRequest(cfg, "POST", StartAuthEndpoint, bytes.NewBuffer(startAuthBody))
	if err != nil {
		return fmt.Errorf("start authentication request failed: %s", err)
	}

	var startAuthResponse StartAuthResponse
	if err := json.Unmarshal(startAuthResp, &startAuthResponse); err != nil {
		return fmt.Errorf("error parsing start auth response: %s", err)
	}

	if !startAuthResponse.Success {
		return fmt.Errorf("start authentication failed: %s", startAuthResponse.ErrorMsg)
	}

	// Handle authentication challenges
	if len(startAuthResponse.Challenges) == 0 {
		return errors.New("no authentication challenges received")
	}

	challenge := startAuthResponse.Challenges[0]
	if len(challenge.Mechanisms) == 0 {
		return errors.New("no authentication mechanisms available")
	}

	// Select mechanism
	fmt.Println("Available authentication mechanisms:")
	for i, mechanism := range challenge.Mechanisms {
		fmt.Printf("%d. %s\n", i+1, mechanism.Name)
	}

	fmt.Print("Select mechanism (1-" + fmt.Sprintf("%d", len(challenge.Mechanisms)) + "): ")
	reader := bufio.NewReader(os.Stdin)
	mechIndexStr, _ := reader.ReadString('\n')
	mechIndexStr = strings.TrimSpace(mechIndexStr)

	var mechIndex int
	if _, err := fmt.Sscanf(mechIndexStr, "%d", &mechIndex); err != nil || mechIndex < 1 || mechIndex > len(challenge.Mechanisms) {
		return errors.New("invalid selection")
	}

	mechanism := challenge.Mechanisms[mechIndex-1]

	// Get authentication answer
	fmt.Print("Enter your response: ")
	var answer string

	// If this is a password mechanism, don't echo input
	if strings.Contains(strings.ToLower(mechanism.Name), "password") {
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("error reading password: %s", err)
		}
		answer = string(passwordBytes)
		fmt.Println() // Add newline after password input
	} else {
		answer, _ = reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
	}

	// Advance authentication
	advanceAuthReq := AdvanceAuthRequest{
		SessionID:   startAuthResponse.SessionID,
		MechanismID: mechanism.MechanismID,
		Answer:      answer,
	}

	advanceAuthBody, err := json.Marshal(advanceAuthReq)
	if err != nil {
		return fmt.Errorf("error marshaling advance auth request: %s", err)
	}

	advanceAuthResp, err := api.MakeRequest(cfg, "POST", AdvanceAuthEndpoint, bytes.NewBuffer(advanceAuthBody))
	if err != nil {
		return fmt.Errorf("advance authentication request failed: %s", err)
	}

	var advanceAuthResponse AdvanceAuthResponse
	if err := json.Unmarshal(advanceAuthResp, &advanceAuthResponse); err != nil {
		return fmt.Errorf("error parsing advance auth response: %s", err)
	}

	if !advanceAuthResponse.Success {
		return fmt.Errorf("advance authentication failed: %s", advanceAuthResponse.ErrorMsg)
	}

	// Save token to config
	cfg.AuthToken = advanceAuthResponse.Token
	cfg.TokenExpiry = time.Now().Add(1 * time.Hour).Unix() // Assuming token valid for 1 hour

	return config.SaveConfig(cfg, configFile)
}
