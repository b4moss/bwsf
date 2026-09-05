package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

// Config is the v2 global configuration (schemaVersion 1).
type Config struct {
	SchemaVersion int            `json:"schemaVersion"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
	AppVersion    string         `json:"app_version"`
	Settings      GlobalSettings `json:"settings"`
}

// GlobalSettings holds save_files and hosts.
type GlobalSettings struct {
	SaveFiles []string `json:"save_files,omitempty"`
	Hosts     []Host   `json:"hosts"`
}

// Host is one Bitwarden host entry.
type Host struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	HostURL          string `json:"host_url"`
	Email            string `json:"email,omitempty"`
	TargetSection    string `json:"target_section"`
	IsDefault        bool   `json:"is_default"`
	DeviceIdentifier string `json:"device_identifier,omitempty"`
}

const (
	configDir       = ".config/bwsf"
	configFileJSON  = "config.json"
	configFileJSONC = "config.jsonc"

	SchemaVersion1 = 1

	HostTypeCloud    = "bitwarden-cloud"
	HostTypeSelfhost = "bitwarden-selfhost"

	DefaultCloudURL   = "https://vault.bitwarden.com"
	DefaultFolderName = "dotenvs"
	DefaultHostID     = "default"
)

// AppVersion is stamped into config on write. cmd sets this at startup.
var AppVersion = "0.0.0"

// nowISO8601 is overridable in tests.
var nowISO8601 = func() string {
	return time.Now().Format(time.RFC3339)
}

// GetConfigDir returns ~/.config/bwsf.
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, configDir), nil
}

// GetConfigPath returns the preferred write path (always config.jsonc).
func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileJSONC), nil
}

// ResolveConfigPaths finds existing config.json / config.jsonc under the config dir.
// Returns ("", "", nil) when neither exists. Both existing is an error.
func ResolveConfigPaths() (jsonPath, jsoncPath string, err error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", "", err
	}
	jp := filepath.Join(dir, configFileJSON)
	jcp := filepath.Join(dir, configFileJSONC)
	hasJSON := fileExists(jp)
	hasJSONC := fileExists(jcp)
	if hasJSON && hasJSONC {
		return "", "", fmt.Errorf("both %s and %s exist; keep only one", jp, jcp)
	}
	if hasJSON {
		return jp, "", nil
	}
	if hasJSONC {
		return "", jcp, nil
	}
	return "", "", nil
}

// MigrateHooks controls flat→v2 migration during LoadConfig.
type MigrateHooks struct {
	// Yes skips the confirm prompt (CLI --yes).
	Yes bool
	// Confirm asks the user; return true to migrate. Nil + !Yes → error on flat.
	Confirm func() (bool, error)
	// Warn emits non-fatal messages (e.g. backend=bw → API warning).
	Warn func(string)
}

// LoadConfig loads and validates global config. Flat files require migration
// (see LoadConfigWithMigrate). Without hooks, flat configs error.
func LoadConfig() (*Config, error) {
	return LoadConfigWithMigrate(MigrateHooks{})
}

// LoadConfigWithMigrate loads config, migrating flat schemas when approved.
func LoadConfigWithMigrate(hooks MigrateHooks) (*Config, error) {
	jsonPath, jsoncPath, err := ResolveConfigPaths()
	if err != nil {
		return nil, err
	}
	if jsonPath == "" && jsoncPath == "" {
		return nil, nil
	}

	path := jsoncPath
	useJSONC := true
	if jsonPath != "" {
		path = jsonPath
		useJSONC = false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if isFlatConfigJSON(data) {
		return migrateFlatConfigFile(path, data, hooks)
	}

	cfg, err := parseGlobalConfigBytes(data, useJSONC)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func parseGlobalConfigBytes(data []byte, jsonc bool) (*Config, error) {
	if err := detectBannedKeys(data); err != nil {
		return nil, err
	}
	var cfg Config
	if jsonc {
		if err := UnmarshalJSONC(data, &cfg); err != nil {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	}
	cfg.Normalize()
	return &cfg, nil
}

// Normalize trims fields and ensures Hosts is non-nil.
func (c *Config) Normalize() {
	if c == nil {
		return
	}
	if c.Settings.Hosts == nil {
		c.Settings.Hosts = []Host{}
	}
	c.Settings.SaveFiles = nonEmptyStrings(c.Settings.SaveFiles)
	for i := range c.Settings.Hosts {
		h := &c.Settings.Hosts[i]
		h.ID = strings.TrimSpace(h.ID)
		h.Type = strings.TrimSpace(h.Type)
		h.HostURL = strings.TrimSpace(h.HostURL)
		h.Email = strings.TrimSpace(h.Email)
		h.TargetSection = strings.TrimSpace(h.TargetSection)
		h.DeviceIdentifier = strings.TrimSpace(h.DeviceIdentifier)
	}
}

// Validate checks schemaVersion and hosts rules.
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	if c.SchemaVersion != SchemaVersion1 {
		return fmt.Errorf("unsupported schemaVersion %d (want %d)", c.SchemaVersion, SchemaVersion1)
	}
	if c.Settings.Hosts == nil {
		c.Settings.Hosts = []Host{}
	}
	seen := make(map[string]struct{})
	defaultCount := 0
	for i, h := range c.Settings.Hosts {
		if err := validateHost(h); err != nil {
			return fmt.Errorf("hosts[%d]: %w", i, err)
		}
		if _, ok := seen[h.ID]; ok {
			return fmt.Errorf("duplicate host id %q", h.ID)
		}
		seen[h.ID] = struct{}{}
		if h.IsDefault {
			defaultCount++
		}
	}
	if len(c.Settings.Hosts) > 0 && defaultCount != 1 {
		return fmt.Errorf("hosts must have exactly one is_default: true (found %d)", defaultCount)
	}
	return nil
}

func detectBannedKeys(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(stripJSONC(data), &raw); err != nil {
		// If we cannot inspect, skip; primary parse will fail.
		return nil
	}
	if _, ok := raw["backend"]; ok {
		return fmt.Errorf("config must not contain backend (API only in v0.20+)")
	}
	if _, ok := raw["not_save_files"]; ok {
		return fmt.Errorf("not_save_files is removed; use save_files with ! prefixes")
	}
	if sraw, ok := raw["settings"]; ok {
		var settings map[string]json.RawMessage
		if err := json.Unmarshal(sraw, &settings); err == nil {
			if _, ok := settings["not_save_files"]; ok {
				return fmt.Errorf("not_save_files is removed; use save_files with ! prefixes")
			}
		}
	}
	return nil
}

func stripJSONC(data []byte) []byte {
	out, err := standardizeJSONC(data)
	if err != nil {
		return data
	}
	return out
}

// validateHost checks one host entry.
func validateHost(h Host) error {
	if h.ID == "" || strings.TrimSpace(h.ID) != h.ID {
		return fmt.Errorf("id must be non-empty without leading/trailing space")
	}
	if strings.Contains(h.ID, "/") {
		return fmt.Errorf("id must not contain '/'")
	}
	for _, r := range h.ID {
		if !unicode.IsPrint(r) || unicode.IsSpace(r) {
			return fmt.Errorf("id must be printable and contain no whitespace")
		}
	}
	switch h.Type {
	case HostTypeCloud, HostTypeSelfhost:
	default:
		return fmt.Errorf("unsupported type %q", h.Type)
	}
	if strings.TrimSpace(h.HostURL) == "" {
		return fmt.Errorf("host_url is required")
	}
	if strings.TrimSpace(h.TargetSection) == "" {
		return fmt.Errorf("target_section is required")
	}
	return nil
}

// EffectiveSaveFiles returns global save_files when non-empty.
func (c *Config) EffectiveSaveFiles() []string {
	if c == nil {
		return nil
	}
	return nonEmptyStrings(c.Settings.SaveFiles)
}

// DefaultHost returns the is_default host, or nil.
func (c *Config) DefaultHost() *Host {
	if c == nil {
		return nil
	}
	for i := range c.Settings.Hosts {
		if c.Settings.Hosts[i].IsDefault {
			return &c.Settings.Hosts[i]
		}
	}
	return nil
}

// FindHost returns a host by id.
func (c *Config) FindHost(id string) *Host {
	if c == nil {
		return nil
	}
	id = strings.TrimSpace(id)
	for i := range c.Settings.Hosts {
		if c.Settings.Hosts[i].ID == id {
			return &c.Settings.Hosts[i]
		}
	}
	return nil
}

// ResolveFolderName returns target_section of the default host, or DefaultFolderName.
// Prefer ResolveHost + host.TargetSection for multi-host callers.
func ResolveFolderName(cfg *Config) string {
	if cfg == nil {
		return DefaultFolderName
	}
	if h := cfg.DefaultHost(); h != nil {
		if name := strings.TrimSpace(h.TargetSection); name != "" {
			return name
		}
	}
	return DefaultFolderName
}

// ValidateFolderName rejects empty or whitespace-only folder names.
func ValidateFolderName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("folder name must not be empty")
	}
	return nil
}

// FormatConfigShow renders human-readable labeled lines for `bwsf config show`.
func FormatConfigShow(cfg *Config) string {
	if cfg == nil {
		cfg = &Config{Settings: GlobalSettings{Hosts: []Host{}}}
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("schemaVersion: %d\n", cfg.SchemaVersion))
	b.WriteString(fmt.Sprintf("app_version: %s\n", cfg.AppVersion))
	if len(cfg.Settings.SaveFiles) == 0 {
		b.WriteString("save_files: (unset)\n")
	} else {
		b.WriteString(fmt.Sprintf("save_files: %s\n", strings.Join(cfg.Settings.SaveFiles, ", ")))
	}
	if len(cfg.Settings.Hosts) == 0 {
		b.WriteString("hosts: (none)\n")
		return b.String()
	}
	b.WriteString(fmt.Sprintf("hosts: (%d)\n", len(cfg.Settings.Hosts)))
	for _, h := range cfg.Settings.Hosts {
		def := ""
		if h.IsDefault {
			def = " [default]"
		}
		b.WriteString(fmt.Sprintf("  - id: %s%s\n", h.ID, def))
		b.WriteString(fmt.Sprintf("    type: %s\n", h.Type))
		b.WriteString(fmt.Sprintf("    host_url: %s\n", h.HostURL))
		b.WriteString(fmt.Sprintf("    email: %s\n", h.Email))
		b.WriteString(fmt.Sprintf("    target_section: %s\n", h.TargetSection))
	}
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

// NewEmptyConfig returns a valid empty v2 config (hosts: []).
func NewEmptyConfig() *Config {
	now := nowISO8601()
	return &Config{
		SchemaVersion: SchemaVersion1,
		CreatedAt:     now,
		UpdatedAt:     now,
		AppVersion:    AppVersion,
		Settings: GlobalSettings{
			Hosts: []Host{},
		},
	}
}

// PrepareForSave sets timestamps and app_version. Preserves created_at when set.
func (c *Config) PrepareForSave() {
	if c == nil {
		return
	}
	now := nowISO8601()
	if strings.TrimSpace(c.CreatedAt) == "" {
		c.CreatedAt = now
	}
	c.UpdatedAt = now
	c.AppVersion = AppVersion
	c.SchemaVersion = SchemaVersion1
	if c.Settings.Hosts == nil {
		c.Settings.Hosts = []Host{}
	}
	c.Normalize()
}

// SaveConfig writes config as config.jsonc and removes config.json if present.
func SaveConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	cfg.PrepareForSave()
	if err := cfg.Validate(); err != nil {
		return err
	}

	dir, err := GetConfigDir()
	if err != nil {
		return err
	}
	if err := mkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	data = append(data, '\n')

	jsoncPath := filepath.Join(dir, configFileJSONC)
	if err := writeFile(jsoncPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	jsonPath := filepath.Join(dir, configFileJSON)
	if fileExists(jsonPath) {
		_ = os.Remove(jsonPath)
	}
	return nil
}

// UpdateHostDeviceIdentifier sets device_identifier on the host with id and saves.
func UpdateHostDeviceIdentifier(cfg *Config, hostID, deviceID string) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	h := cfg.FindHost(hostID)
	if h == nil {
		return fmt.Errorf("host %q not found", hostID)
	}
	h.DeviceIdentifier = deviceID
	return SaveConfig(cfg)
}

// OS helpers used by SaveConfig (overridable in tests; root ignores chmod tricks).
var (
	mkdirAll  = os.MkdirAll
	writeFile = os.WriteFile
)

// OverrideSaveConfigIO swaps SaveConfig filesystem helpers for tests.
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
