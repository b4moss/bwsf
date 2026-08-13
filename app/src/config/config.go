package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	HostType      string `json:"host_type"`             // "cloud" or "selfhosted"
	SelfhostedURL string `json:"selfhosted_url"`        // URL for self-hosted instance
	Email         string `json:"email"`                 // Email address
	FolderName    string `json:"folder_name,omitempty"` // Bitwarden folder for .env notes
}

const (
	configDir  = ".config/bwsf"
	configFile = "config.json"

	// DefaultFolderName is the Bitwarden folder used when folder_name is unset.
	DefaultFolderName = "dotenvs"
)

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

// FormatConfigShow renders human-readable labeled lines for `bwsf config show`.
func FormatConfigShow(cfg *Config) string {
	if cfg == nil {
		cfg = &Config{}
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Host type: %s\n", cfg.HostType))
	b.WriteString(fmt.Sprintf("Self-hosted URL: %s\n", cfg.SelfhostedURL))
	b.WriteString(fmt.Sprintf("Email: %s\n", cfg.Email))
	b.WriteString(fmt.Sprintf("Folder name: %s\n", ResolveFolderName(cfg)))
	return b.String()
}

// LoadConfigShowText loads config and returns show text, or an error when missing/invalid.
func LoadConfigShowText() (string, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return "", err
	}
	if cfg == nil {
		return "", fmt.Errorf("no config found; run `bwsf setup` first")
	}
	return FormatConfigShow(cfg), nil
}

// SaveConfig saves the configuration to file
func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDirPath := filepath.Dir(configPath)
	if err := os.MkdirAll(configDirPath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
