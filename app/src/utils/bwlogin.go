package utils

import (
	"fmt"
	"os"
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
	if err := requireBwInstalled(); err != nil {
		return false, err.Error()
	}

	StartSpinner("Logging in...")
	defer StopSpinner()

	if serverURL != "" {
		checkRes := runBw("bw", []string{"config", "server"}, bwRunOptions{})
		currentServer := strings.TrimSpace(string(checkRes.Output))

		if currentServer != "" && currentServer != serverURL {
			_ = runBw("bw", []string{"logout"}, bwRunOptions{})
		}

		configRes := runBw("bw", []string{"config", "server", serverURL}, bwRunOptions{})
		if configRes.Err != nil {
			errorMsg := strings.TrimSpace(string(configRes.Output))
			if errorMsg == "" {
				errorMsg = configRes.Err.Error()
			}
			if strings.Contains(errorMsg, "Logout required") {
				_ = runBw("bw", []string{"logout"}, bwRunOptions{})
				configRes = runBw("bw", []string{"config", "server", serverURL}, bwRunOptions{})
				if configRes.Err != nil {
					errorMsg = strings.TrimSpace(string(configRes.Output))
					if errorMsg == "" {
						errorMsg = configRes.Err.Error()
					}
					return false, fmt.Sprintf("Failed to configure server: %s", errorMsg)
				}
			} else {
				return false, fmt.Sprintf("Failed to configure server: %s", errorMsg)
			}
		}
	}

	args := []string{"login", email, password, "--raw"}
	res := runBw("bw", args, bwRunOptions{})
	outputStr := strings.TrimSpace(string(res.Output))

	if res.Err != nil {
		errorMsg := outputStr
		if errorMsg == "" {
			errorMsg = res.Err.Error()
		}
		if strings.Contains(strings.ToLower(errorMsg), "already logged in") {
			if session, unlockErr := unlockRaw(password); unlockErr == nil {
				os.Setenv("BW_SESSION", session)
				return true, ""
			}
			return false, fmt.Sprintf("%s (unlock after login failed)", errorMsg)
		}
		return false, errorMsg
	}

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

	res := runBw("bw", []string{"unlock", "--raw", "--passwordenv", "BW_PASSWORD"}, bwRunOptions{
		Env: append(environWithout("BW_PASSWORD", "BW_SESSION"), "BW_PASSWORD="+password),
	})
	combined := strings.TrimSpace(string(res.Output))
	session := extractSessionKey(combined)
	if res.Err != nil {
		if combined == "" {
			return "", res.Err
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
	path, err := lookPath("bw")
	if err != nil {
		return false, ""
	}
	return true, path
}
