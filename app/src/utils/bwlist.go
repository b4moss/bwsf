package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
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

// resolveConfiguredFolderName loads target_section from the default host (default: "dotenvs").
func resolveConfiguredFolderName() string {
	cfg, err := config.LoadConfig()
	if err != nil {
		return config.DefaultFolderName
	}
	return config.ResolveFolderName(cfg)
}

// GetDotenvsFolderID retrieves the folder ID for the configured Bitwarden folder
// (target_section on the default host, default "dotenvs"). Signature kept for BwClient compatibility.
func GetDotenvsFolderID() (string, error) {
	return GetFolderID(resolveConfiguredFolderName())
}

// GetFolderID retrieves the folder ID for the given Bitwarden folder name.
func GetFolderID(folderName string) (string, error) {
	if folderName == "" {
		folderName = config.DefaultFolderName
	}

	if err := requireBwInstalled(); err != nil {
		return "", err
	}

	StartSpinner("Fetching folders...")
	defer StopSpinner()

	res := runBw("bw", []string{"list", "folders"}, bwRunOptions{})
	if res.Err != nil {
		errorMsg := strings.TrimSpace(string(res.Output))
		if errorMsg == "" {
			errorMsg = res.Err.Error()
		}
		if isBwAuthRequiredMessage(errorMsg) {
			return "", ErrBitwardenLocked
		}
		return "", fmt.Errorf("failed to list folders: %s", errorMsg)
	}

	var folders []Folder
	outputStr := strings.TrimSpace(string(res.Output))
	if outputStr == "" {
		return "", fmt.Errorf("no output from bw list folders command")
	}

	if isBwAuthRequiredMessage(outputStr) {
		return "", ErrBitwardenLocked
	}

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

	if err := requireBwInstalled(); err != nil {
		return err
	}

	StartSpinner(fmt.Sprintf("Creating %s folder...", folderName))
	defer StopSpinner()

	folderData := map[string]string{
		"name": folderName,
	}
	jsonData, err := json.Marshal(folderData)
	if err != nil {
		return fmt.Errorf("failed to marshal folder data: %w", err)
	}

	encodedData := base64.StdEncoding.EncodeToString(jsonData)

	res := runBw("bw", []string{"create", "folder", encodedData}, bwRunOptions{})
	if res.Err != nil {
		errorMsg := strings.TrimSpace(string(res.Output))
		if errorMsg == "" {
			errorMsg = res.Err.Error()
		}
		return fmt.Errorf("failed to create folder: %s", errorMsg)
	}

	_ = runBw("bw", []string{"sync"}, bwRunOptions{})

	return nil
}

// ListItemsInFolder retrieves all items in the specified folder
func ListItemsInFolder(folderID string) ([]Item, error) {
	if err := requireBwInstalled(); err != nil {
		return nil, err
	}

	StartSpinner("Listing items...")
	defer StopSpinner()

	res := runBw("bw", []string{"list", "items", "--folderid", folderID}, bwRunOptions{})
	if res.Err != nil {
		errorMsg := strings.TrimSpace(string(res.Output))
		if errorMsg == "" {
			errorMsg = res.Err.Error()
		}
		if isBwAuthRequiredMessage(errorMsg) {
			return nil, ErrBitwardenLocked
		}
		return nil, fmt.Errorf("failed to list items: %s", errorMsg)
	}

	var items []Item
	outputStr := strings.TrimSpace(string(res.Output))
	if outputStr == "" {
		return nil, fmt.Errorf("no output from bw list items command")
	}

	if isBwAuthRequiredMessage(outputStr) {
		return nil, ErrBitwardenLocked
	}

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
	if err := requireBwInstalled(); err != nil {
		return false, err.Error()
	}
	if masterPassword == "" {
		return false, "master password is empty"
	}

	StartSpinner("Unlocking vault...")
	defer StopSpinner()

	session, err := unlockRaw(masterPassword)
	if err != nil {
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

		res := runBw("bw", []string{"unlock", "--raw", "--passwordfile", tmpName}, bwRunOptions{
			Env:             environWithout("BW_SESSION"),
			SeparateStreams: true,
		})
		out := strings.TrimSpace(string(res.Output))
		stderrStr := strings.TrimSpace(string(res.Stderr))
		session = extractSessionKey(out)
		if res.Err != nil || session == "" {
			msg := err.Error()
			if stderrStr != "" {
				msg = stderrStr
			} else if out != "" {
				msg = out
			} else if res.Err != nil {
				msg = res.Err.Error()
			}
			return false, msg
		}
	}

	os.Setenv("BW_SESSION", session)
	return true, ""
}
