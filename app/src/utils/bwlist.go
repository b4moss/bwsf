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

// ErrBitwardenLocked is returned when Bitwarden CLI is locked or unauthenticated
var ErrBitwardenLocked = errors.New("Bitwarden CLI is locked")

// isBwAuthRequiredMessage reports whether bw CLI output indicates auth is required (#161).
func isBwAuthRequiredMessage(s string) bool {
	return strings.Contains(s, "Master password") ||
		strings.Contains(s, "master password") ||
		strings.Contains(s, "You are not logged in") ||
		strings.Contains(s, "Vault is locked")
}

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
		if isBwAuthRequiredMessage(errorMsg) {
			return "", ErrBitwardenLocked
		}
		return "", fmt.Errorf("failed to list folders: %s", errorMsg)
	}

	// Parse JSON output
	var folders []Folder
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return "", fmt.Errorf("no output from bw list folders command")
	}

	// Check if Bitwarden CLI requires auth (locked / unauthenticated)
	if isBwAuthRequiredMessage(outputStr) {
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
		if isBwAuthRequiredMessage(errorMsg) {
			return nil, ErrBitwardenLocked
		}
		return nil, fmt.Errorf("failed to list items: %s", errorMsg)
	}

	// Parse JSON output
	var items []Item
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return nil, fmt.Errorf("no output from bw list items command")
	}

	// Check if Bitwarden CLI requires auth (locked / unauthenticated)
	if isBwAuthRequiredMessage(outputStr) {
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

// BwUnlock executes bw unlock with the master password and sets BW_SESSION.
// Uses the same --passwordenv --raw path as login's unlockRaw (the reliable non-interactive API).
func BwUnlock(masterPassword string) (bool, string) {
	if _, err := exec.LookPath("bw"); err != nil {
		return false, "bw command is not installed"
	}
	if masterPassword == "" {
		return false, "master password is empty"
	}

	StartSpinner("Unlocking vault...")
	defer StopSpinner()

	session, err := unlockRaw(masterPassword)
	if err != nil {
		// Fallback: passwordfile with trailing newline ("first line" per bw docs).
		// Empty password files make bw exit 0 with no output — treat that as failure.
		tmpFile, tmpErr := os.CreateTemp("", "bwsf-password-*")
		if tmpErr != nil {
			return false, err.Error()
		}
		tmpName := tmpFile.Name()
		defer os.Remove(tmpName)

		if _, wErr := tmpFile.WriteString(masterPassword + "\n"); wErr != nil {
			tmpFile.Close()
			return false, err.Error()
		}
		if cErr := tmpFile.Close(); cErr != nil {
			return false, err.Error()
		}

		cmd := exec.Command("bw", "unlock", "--raw", "--passwordfile", tmpName)
		cmd.Env = environWithout("BW_SESSION") // avoid stale session confusing unlock
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		runErr := cmd.Run()
		out := strings.TrimSpace(stdoutBuf.String())
		stderrStr := strings.TrimSpace(stderrBuf.String())
		session = extractSessionKey(out)
		if runErr != nil || session == "" {
			msg := err.Error()
			if stderrStr != "" {
				msg = stderrStr
			} else if out != "" {
				msg = out
			} else if runErr != nil {
				msg = runErr.Error()
			}
			return false, msg
		}
	}

	os.Setenv("BW_SESSION", session)
	return true, ""
}
