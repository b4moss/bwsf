package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	HostType         string `json:"host_type"`                   // "cloud" or "selfhosted"
	SelfhostedURL    string `json:"selfhosted_url"`              // URL for self-hosted instance
	Email            string `json:"email"`                       // Email address
	FolderName       string `json:"folder_name,omitempty"`       // Bitwarden folder for .env notes
	Backend          string `json:"backend,omitempty"`           // "bw" (CLI) or "api"; default "bw" when unset
	DeviceIdentifier string `json:"device_identifier,omitempty"` // Stable device id for Identity (api backend)
}

const (
	configDir  = ".config/bwsf"
	configFile = "config.json"

	// BackendBW uses the Bitwarden CLI (`bw`) adapter.
	BackendBW = "bw"
	// BackendAPI uses the Bitwarden API adapter (not fully implemented yet).
	BackendAPI = "api"

	// DefaultFolderName is the Bitwarden folder used when folder_name is unset.
	DefaultFolderName = "dotenvs"
)

// GetBackend returns the configured backend, defaulting to "bw" when unset (backward compatible).
func (c *Config) GetBackend() string {
	if c == nil || c.Backend == "" {
		return BackendBW
	}
	return c.Backend
}

// IsValidBackend reports whether backend is a supported value ("bw" or "api").
func IsValidBackend(backend string) bool {
	return backend == BackendBW || backend == BackendAPI
}

// ResolveFolderName returns the configured folder name, or DefaultFolderName when empty.
func ResolveFolderName(cfg *Config) string {
	if cfg == nil {
		return DefaultFolderName
	}
	name := strings.TrimSpace(cfg.FolderName)
	if name == "" {
		return DefaultFolderName
	}
	return name
}

// ValidateFolderName rejects empty or whitespace-only folder names.
func ValidateFolderName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("folder name must not be empty")
	}
	return nil
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, configDir, configFile), nil
}

// LoadConfig loads the configuration from file
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil // Config doesn't exist yet, return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDirPath := filepath.Dir(configPath)
	if err := mkdirAll(configDirPath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := writeFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// OS helpers used by SaveConfig (overridable in tests; root ignores chmod tricks).
var (
	mkdirAll  = os.MkdirAll
	writeFile = os.WriteFile
)

// OverrideSaveConfigIO swaps SaveConfig filesystem helpers for tests. Call the
// returned function to restore the previous implementations.
func OverrideSaveConfigIO(
	mkdirAllFn func(string, os.FileMode) error,
	writeFileFn func(string, []byte, os.FileMode) error,
) func() {
	prevMkdir, prevWrite := mkdirAll, writeFile
	if mkdirAllFn != nil {
		mkdirAll = mkdirAllFn
	}
	if writeFileFn != nil {
		writeFile = writeFileFn
	}
	return func() {
		mkdirAll, writeFile = prevMkdir, prevWrite
	}
}
