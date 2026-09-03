package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// looksLikeSessionKey reports whether s looks like a bw --raw session key.
// Session keys are base64 of 64 random bytes (typically 88 chars).
func looksLikeSessionKey(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < 40 || strings.ContainsAny(s, " \t\n\r") {
		return false
	}
	for _, r := range s {
		ok := (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '+' || r == '/' || r == '=' || r == '-' || r == '_'
		if !ok {
			return false
		}
	}
	return true
}

// extractSessionKey picks the first line that looks like a session key from mixed output.
func extractSessionKey(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if looksLikeSessionKey(line) {
			return line
		}
	}
	return ""
}

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
			if session, unlockErr := unlockRaw(password); unlockErr == nil {
				os.Setenv("BW_SESSION", session)
				return true, ""
			}
			return false, fmt.Sprintf("%s (unlock after login failed)", errorMsg)
		}
		return false, errorMsg
	}

	// --raw returns the session key on success
	if session := extractSessionKey(outputStr); session != "" {
		os.Setenv("BW_SESSION", session)
		return true, ""
	}

	if strings.Contains(outputStr, "You are logged in") ||
		strings.Contains(outputStr, "You're logged in") ||
		strings.Contains(outputStr, "logged in!") {
		session, unlockErr := unlockRaw(password)
		if unlockErr != nil {
			return false, fmt.Sprintf("logged in but failed to unlock: %v", unlockErr)
		}
		os.Setenv("BW_SESSION", session)
		return true, ""
	}

	return false, outputStr
}

func unlockRaw(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("master password is empty")
	}

	cmd := exec.Command("bw", "unlock", "--raw", "--passwordenv", "BW_PASSWORD")
	cmd.Env = append(environWithout("BW_PASSWORD"), "BW_PASSWORD="+password)
	out, err := cmd.CombinedOutput()
	combined := strings.TrimSpace(string(out))
	session := extractSessionKey(combined)
	if err != nil {
		if combined == "" {
			return "", err
		}
		return "", fmt.Errorf("%s", combined)
	}
	if session == "" {
		return "", fmt.Errorf("unlock returned no session key (%s)", truncateForErr(combined))
	}
	return session, nil
}

// environWithout returns os.Environ() without keys in drop (case-sensitive key match).
func environWithout(drop ...string) []string {
	deny := make(map[string]struct{}, len(drop))
	for _, k := range drop {
		deny[k] = struct{}{}
	}
	env := os.Environ()
	out := make([]string, 0, len(env))
	for _, kv := range env {
		key, _, _ := strings.Cut(kv, "=")
		if _, skip := deny[key]; skip {
			continue
		}
		out = append(out, kv)
	}
	return out
}

func truncateForErr(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "empty output"
	}
	if len(s) > 120 {
		return s[:120] + "..."
	}
	return s
}

// CheckBwCommand checks if bw command is installed
func CheckBwCommand() (bool, string) {
	path, err := exec.LookPath("bw")
	if err != nil {
		return false, ""
	}
	return true, path
}
