package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// BwLogin executes bw login command and returns success status and error message
func BwLogin(email, password, serverURL string) (bool, string) {
	// Check if bw command exists
	_, err := exec.LookPath("bw")
	if err != nil {
		return false, "bw command is not installed"
	}

	// Start spinner
	StartSpinner("Logging in...")
	defer StopSpinner()

	// If self-hosted, configure server first
	// But first check if we need to logout
	if serverURL != "" {
		// Check current server config
		checkCmd := exec.Command("bw", "config", "server")
		checkOutput, _ := checkCmd.CombinedOutput()
		currentServer := strings.TrimSpace(string(checkOutput))

		// If server URL is different, logout first
		if currentServer != "" && currentServer != serverURL {
			logoutCmd := exec.Command("bw", "logout")
			logoutCmd.Run() // Ignore errors, just try to logout
		}

		configCmd := exec.Command("bw", "config", "server", serverURL)
		configOutput, err := configCmd.CombinedOutput()
		if err != nil {
			errorMsg := strings.TrimSpace(string(configOutput))
			if errorMsg == "" {
				errorMsg = err.Error()
			}
			// If error is about logout required, try logout and retry
			if strings.Contains(errorMsg, "Logout required") {
				logoutCmd := exec.Command("bw", "logout")
				logoutCmd.Run() // Ignore errors
				// Retry config
				configOutput, err = configCmd.CombinedOutput()
				if err != nil {
					errorMsg = strings.TrimSpace(string(configOutput))
					if errorMsg == "" {
						errorMsg = err.Error()
					}
					return false, fmt.Sprintf("Failed to configure server: %s", errorMsg)
				}
			} else {
				return false, fmt.Sprintf("Failed to configure server: %s", errorMsg)
			}
		}
	}

	// Prefer --raw so Docker / non-interactive runs get a session key.
	args := []string{"login", email, password, "--raw"}

	cmd := exec.Command("bw", args...)
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		errorMsg := outputStr
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		// Already logged in: try to unlock for a session instead of failing hard.
		if strings.Contains(strings.ToLower(errorMsg), "already logged in") {
			if session, unlockErr := unlockRaw(password); unlockErr == nil && session != "" {
				os.Setenv("BW_SESSION", session)
				return true, ""
			}
		}
		return false, errorMsg
	}

	// --raw returns the session key on success
	if outputStr != "" &&
		!strings.Contains(outputStr, " ") &&
		!strings.Contains(strings.ToLower(outputStr), "error") {
		os.Setenv("BW_SESSION", outputStr)
		return true, ""
	}

	if strings.Contains(outputStr, "You are logged in") ||
		strings.Contains(outputStr, "You're logged in") ||
		strings.Contains(outputStr, "logged in!") {
		if session, unlockErr := unlockRaw(password); unlockErr == nil && session != "" {
			os.Setenv("BW_SESSION", session)
		}
		return true, ""
	}

	return false, outputStr
}

func unlockRaw(password string) (string, error) {
	cmd := exec.Command("bw", "unlock", "--raw", "--passwordenv", "BW_PASSWORD")
	cmd.Env = append(os.Environ(), "BW_PASSWORD="+password)
	out, err := cmd.CombinedOutput()
	session := strings.TrimSpace(string(out))
	if err != nil {
		return "", fmt.Errorf("%s", session)
	}
	return session, nil
}

// CheckBwCommand checks if bw command is installed
func CheckBwCommand() (bool, string) {
	path, err := exec.LookPath("bw")
	if err != nil {
		return false, ""
	}
	return true, path
}
