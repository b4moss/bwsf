package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"bwsf/src/config"
)

// ErrBitwardenLocked is returned when Bitwarden CLI is locked
var ErrBitwardenLocked = errors.New("Bitwarden CLI is locked")

// Folder represents a Bitwarden folder
type Folder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Item represents a Bitwarden item
type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// resolveConfiguredFolderName loads folder_name from config (default: "dotenvs").
func resolveConfiguredFolderName() string {
	cfg, err := config.LoadConfig()
	if err != nil {
		return config.DefaultFolderName
	}
	return config.ResolveFolderName(cfg)
}

// GetDotenvsFolderID retrieves the folder ID for the configured Bitwarden folder
// (folder_name in config, default "dotenvs"). Signature kept for BwClient compatibility.
func GetDotenvsFolderID() (string, error) {
	return GetFolderID(resolveConfiguredFolderName())
}

// GetFolderID retrieves the folder ID for the given Bitwarden folder name.
func GetFolderID(folderName string) (string, error) {
	if folderName == "" {
		folderName = config.DefaultFolderName
	}

	// Check if bw command exists
	_, err := exec.LookPath("bw")
	if err != nil {
		return "", fmt.Errorf("bw command is not installed")
	}

	// Start spinner
	StartSpinner("Fetching folders...")
	defer StopSpinner()

	// Execute bw list folders command
	cmd := exec.Command("bw", "list", "folders")
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorMsg := strings.TrimSpace(string(output))
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		return "", fmt.Errorf("failed to list folders: %s", errorMsg)
	}

	// Parse JSON output
	var folders []Folder
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return "", fmt.Errorf("no output from bw list folders command")
	}

	// Check if Bitwarden CLI is locked (requires master password)
	if strings.Contains(outputStr, "Master password") || strings.Contains(outputStr, "master password") {
		return "", ErrBitwardenLocked
	}

	// Check if output looks like JSON (starts with '[' or '{')
	if !strings.HasPrefix(outputStr, "[") && !strings.HasPrefix(outputStr, "{") {
		return "", fmt.Errorf("unexpected output from bw list folders (not JSON): %s", outputStr)
	}

	if err := json.Unmarshal([]byte(outputStr), &folders); err != nil {
		return "", fmt.Errorf("failed to parse folders JSON (output: %s): %w", outputStr, err)
	}

	for _, folder := range folders {
		if folder.Name == folderName {
			return folder.ID, nil
		}
	}

	return "", fmt.Errorf("%s folder not found", folderName)
}

// ErrDotenvsFolderNotFound is returned when the configured folder does not exist
var ErrDotenvsFolderNotFound = errors.New("dotenvs folder not found")

// DotenvsFolderExists checks if the configured folder exists in Bitwarden.
func DotenvsFolderExists() (bool, error) {
	return FolderExists(resolveConfiguredFolderName())
}

// FolderExists checks if the given folder exists in Bitwarden.
func FolderExists(folderName string) (bool, error) {
	_, err := GetFolderID(folderName)
	if err != nil {
		if strings.Contains(err.Error(), "folder not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CreateDotenvsFolder creates the configured folder in Bitwarden.
func CreateDotenvsFolder() error {
	return CreateFolder(resolveConfiguredFolderName())
}

// CreateFolder creates a Bitwarden folder with the given name.
func CreateFolder(folderName string) error {
	if folderName == "" {
		folderName = config.DefaultFolderName
	}

	// Check if bw command exists
	_, err := exec.LookPath("bw")
	if err != nil {
		return fmt.Errorf("bw command is not installed")
	}

	// Start spinner
	StartSpinner(fmt.Sprintf("Creating %s folder...", folderName))
	defer StopSpinner()

	// Create folder JSON object
	folderData := map[string]string{
		"name": folderName,
	}
	jsonData, err := json.Marshal(folderData)
	if err != nil {
		return fmt.Errorf("failed to marshal folder data: %w", err)
	}

	// Base64 encode the JSON (required by bw create folder)
	encodedData := base64.StdEncoding.EncodeToString(jsonData)

	// Execute bw create folder command
	cmd := exec.Command("bw", "create", "folder", encodedData)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorMsg := strings.TrimSpace(string(output))
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		return fmt.Errorf("failed to create folder: %s", errorMsg)
	}

	// Sync to ensure the folder is available
	syncCmd := exec.Command("bw", "sync")
	syncCmd.CombinedOutput() // Ignore errors from sync

	return nil
}

// ListItemsInFolder retrieves all items in the specified folder
func ListItemsInFolder(folderID string) ([]Item, error) {
	// Check if bw command exists
	_, err := exec.LookPath("bw")
	if err != nil {
		return nil, fmt.Errorf("bw command is not installed")
	}

	// Start spinner
	StartSpinner("Listing items...")
	defer StopSpinner()

	// Execute bw list items command with folder filter
	cmd := exec.Command("bw", "list", "items", "--folderid", folderID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorMsg := strings.TrimSpace(string(output))
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		return nil, fmt.Errorf("failed to list items: %s", errorMsg)
	}

	// Parse JSON output
	var items []Item
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return nil, fmt.Errorf("no output from bw list items command")
	}

	// Check if Bitwarden CLI is locked (requires master password)
	if strings.Contains(outputStr, "Master password") || strings.Contains(outputStr, "master password") {
		return nil, ErrBitwardenLocked
	}

	// Check if output looks like JSON (starts with '[' or '{')
	if !strings.HasPrefix(outputStr, "[") && !strings.HasPrefix(outputStr, "{") {
		return nil, fmt.Errorf("unexpected output from bw list items (not JSON): %s", outputStr)
	}

	if err := json.Unmarshal([]byte(outputStr), &items); err != nil {
		return nil, fmt.Errorf("failed to parse items JSON (output: %s): %w", outputStr, err)
	}

	return items, nil
}

// BwUnlock executes bw unlock command with the provided master password
// Returns true if successful, and sets BW_SESSION environment variable if session key is returned
// Note: Due to issues with bw unlock in Docker containers, this function may not work as expected
// If unlock fails, consider using BwLoginWithPassword instead
func BwUnlock(masterPassword string) (bool, string) {
	// Check if bw command exists
	_, err := exec.LookPath("bw")
	if err != nil {
		return false, "bw command is not installed"
	}

	// Start spinner
	StartSpinner("Unlocking vault...")
	defer StopSpinner()

	// Declare variables at function scope
	var output []byte
	var stderr []byte
	var stderrStr string

	// Try method 1: Use --passwordfile option (most reliable for non-interactive use)
	// Create a temporary file with the password
	tmpFile, tmpErr := os.CreateTemp("", "bwsf-password-*")
	if tmpErr == nil {
		defer os.Remove(tmpFile.Name()) // Clean up temp file
		tmpFile.WriteString(masterPassword)
		tmpFile.Close()

		// Try with --raw first to get session key directly
		cmd := exec.Command("bw", "unlock", "--raw", "--passwordfile", tmpFile.Name())
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf

		err = cmd.Run()
		output = stdoutBuf.Bytes()
		stderr = stderrBuf.Bytes()

		// Check if stderr contains password prompt
		stderrStr = strings.TrimSpace(string(stderr))
		outputStr := strings.TrimSpace(string(output))

		// If --raw didn't work, try without --raw to get export command
		if err == nil && outputStr == "" && !strings.Contains(stderrStr, "Master password") && !strings.Contains(stderrStr, "master password") {
			cmd = exec.Command("bw", "unlock", "--passwordfile", tmpFile.Name())
			stdoutBuf.Reset()
			stderrBuf.Reset()
			cmd.Stdout = &stdoutBuf
			cmd.Stderr = &stderrBuf
			err = cmd.Run()
			output = stdoutBuf.Bytes()
			stderr = stderrBuf.Bytes()
			outputStr = strings.TrimSpace(string(output))
			stderrStr = strings.TrimSpace(string(stderr))

			// Parse export BW_SESSION="..." command
			if strings.Contains(outputStr, "export BW_SESSION=") {
				// Extract session key from export command
				// Format: export BW_SESSION="session_key_here"
				startIdx := strings.Index(outputStr, "BW_SESSION=\"")
				if startIdx != -1 {
					startIdx += len("BW_SESSION=\"")
					endIdx := strings.Index(outputStr[startIdx:], "\"")
					if endIdx != -1 {
						sessionKey := outputStr[startIdx : startIdx+endIdx]
						os.Setenv("BW_SESSION", sessionKey)
						outputStr = sessionKey // Use extracted session key
					}
				}
			}
		}

		// Check if we got a session key (either from --raw or from export command)
		if err == nil && !strings.Contains(stderrStr, "Master password") && !strings.Contains(stderrStr, "master password") {
			if outputStr != "" {
				os.Setenv("BW_SESSION", outputStr)
				// Verify unlock succeeded
				statusCmd := exec.Command("bw", "status")
				statusOutput, statusErr := statusCmd.CombinedOutput()
				if statusErr == nil {
					statusStr := string(statusOutput)
					if strings.Contains(statusStr, "\"status\":\"unlocked\"") {
						return true, ""
					}
				}
			}
		}
		// If we get here, method 1 failed, continue to next method
		if err != nil || strings.Contains(stderrStr, "Master password") || strings.Contains(stderrStr, "master password") || outputStr == "" {
			err = fmt.Errorf("method 1 failed")
		}
	} else {
		err = tmpErr
	}

	// Try method 2: Use --passwordenv option with environment variable
	if err != nil {
		cmd := exec.Command("bw", "unlock", "--raw", "--passwordenv", "BW_PASSWORD")
		cmd.Env = append(os.Environ(), "BW_PASSWORD="+masterPassword)
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf

		err = cmd.Run()
		output = stdoutBuf.Bytes()
		stderr = stderrBuf.Bytes()

		// Check if stderr contains password prompt
		stderrStr = strings.TrimSpace(string(stderr))
		if err == nil && !strings.Contains(stderrStr, "Master password") && !strings.Contains(stderrStr, "master password") {
			// Success! Check if we got a session key
			outputStr := strings.TrimSpace(string(output))
			if outputStr != "" {
				os.Setenv("BW_SESSION", outputStr)
				return true, ""
			}
		}
	}

	// If that fails, try method 3: Pass password as argument with --raw
	if err != nil {
		cmd := exec.Command("bw", "unlock", "--raw", masterPassword)
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		err = cmd.Run()
		output = stdoutBuf.Bytes()
		stderr = stderrBuf.Bytes()

		// Check if stderr contains password prompt
		stderrStr = strings.TrimSpace(string(stderr))
		if err == nil && !strings.Contains(stderrStr, "Master password") && !strings.Contains(stderrStr, "master password") {
			// Success! Check if we got a session key
			outputStr := strings.TrimSpace(string(output))
			if outputStr != "" {
				os.Setenv("BW_SESSION", outputStr)
				return true, ""
			}
		}
	}

	// If that fails, try method 4: Pass password as argument without --raw
	if err != nil {
		cmd := exec.Command("bw", "unlock", masterPassword)
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		err = cmd.Run()
		output = stdoutBuf.Bytes()
		stderr = stderrBuf.Bytes()

		// Check if stderr contains password prompt
		stderrStr = strings.TrimSpace(string(stderr))
		if err == nil && !strings.Contains(stderrStr, "Master password") && !strings.Contains(stderrStr, "master password") {
			// Success! Check if we got a session key
			outputStr := strings.TrimSpace(string(output))
			if outputStr != "" {
				os.Setenv("BW_SESSION", outputStr)
				return true, ""
			}
		}
	}

	// If all methods failed, return error
	if err != nil {
		// Combine stdout and stderr for error message
		allOutput := strings.TrimSpace(string(output))
		if stderrStr != "" {
			if allOutput != "" {
				allOutput += " (stderr: " + stderrStr + ")"
			} else {
				allOutput = stderrStr
			}
		}
		if allOutput == "" {
			allOutput = fmt.Sprintf("command failed with exit code: %v", err)
		} else {
			allOutput = fmt.Sprintf("command failed: %s", allOutput)
		}
		return false, allOutput
	}

	// Check if unlock was successful
	outputStr := strings.TrimSpace(string(output))
	stderrStr = strings.TrimSpace(string(stderr))

	// bw unlock may return a session key (a long string)
	// Session keys are typically base64-like strings, 40+ characters
	// Check each line of output for a potential session key
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Session key is typically a long alphanumeric string (base64-like)
		// Check if it looks like a session key: long, no spaces, mostly alphanumeric
		if len(line) >= 40 && !strings.Contains(line, " ") && !strings.Contains(line, "\t") {
			// Additional check: should be mostly alphanumeric with possible base64 chars
			hasValidChars := true
			for _, r := range line {
				if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '+' || r == '/' || r == '=') {
					hasValidChars = false
					break
				}
			}
			if hasValidChars {
				// This looks like a session key
				os.Setenv("BW_SESSION", line)
				return true, ""
			}
		}
	}

	// Check for success messages in both stdout and stderr
	combinedOutput := outputStr + " " + stderrStr
	if strings.Contains(combinedOutput, "You are logged in") ||
		strings.Contains(combinedOutput, "You're logged in") ||
		strings.Contains(combinedOutput, "logged in!") ||
		strings.Contains(combinedOutput, "unlocked") ||
		strings.Contains(combinedOutput, "Unlocked") {
		// If we got a session key in the output, set it
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) >= 40 && !strings.Contains(line, " ") {
				os.Setenv("BW_SESSION", line)
				break
			}
		}
		return true, ""
	}

	// If output is not empty but doesn't match success patterns,
	// it might still be a session key (try setting it anyway if it's a single line)
	if outputStr != "" && !strings.Contains(outputStr, "\n") && len(outputStr) >= 20 {
		// Might be a session key, set it and assume success
		os.Setenv("BW_SESSION", outputStr)
		return true, ""
	}

	// Always check bw status to verify if unlock actually succeeded
	// This is the most reliable way to check if we're unlocked
	// Note: Even if bw unlock returns no output, it might have succeeded
	// So we check status after each unlock attempt
	statusCmd := exec.Command("bw", "status")
	statusOutput, statusErr := statusCmd.CombinedOutput()
	if statusErr == nil {
		statusStr := string(statusOutput)
		// Check if status shows we're unlocked (JSON format: "status":"unlocked")
		if strings.Contains(statusStr, "\"status\":\"unlocked\"") {
			// Unlock succeeded! Even if we didn't get a session key in output,
			// the unlock was successful, so we can proceed
			// Try to extract session key from data.json if available
			// But for now, just return success - subsequent commands should work
			return true, ""
		}
	}

	// If output is not empty, try to extract session key
	if outputStr != "" {
		cleanOutput := strings.TrimSpace(outputStr)
		if len(cleanOutput) >= 20 && len(cleanOutput) < 200 {
			os.Setenv("BW_SESSION", cleanOutput)
			// Verify unlock succeeded with status check
			statusCmd := exec.Command("bw", "status")
			statusOutput, statusErr := statusCmd.CombinedOutput()
			if statusErr == nil {
				statusStr := string(statusOutput)
				if strings.Contains(statusStr, "\"status\":\"unlocked\"") {
					return true, ""
				}
			}
		}
	}

	// If we get here, unlock didn't succeed
	errorMsg := fmt.Sprintf("unlock command completed but status check shows still locked (stdout: %q, stderr: %q)", outputStr, stderrStr)
	return false, errorMsg
}
