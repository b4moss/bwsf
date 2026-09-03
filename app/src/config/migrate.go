package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// legacyFlatConfig is the pre-v0.20 flat schema (migration only).
type legacyFlatConfig struct {
	HostType         string `json:"host_type"`
	SelfhostedURL    string `json:"selfhosted_url"`
	Email            string `json:"email"`
	FolderName       string `json:"folder_name"`
	Backend          string `json:"backend"`
	DeviceIdentifier string `json:"device_identifier"`
}

// isFlatConfigJSON reports whether data looks like the old flat schema.
func isFlatConfigJSON(data []byte) bool {
	standardized := stripJSONC(data)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(standardized, &raw); err != nil {
		return false
	}
	if _, ok := raw["schemaVersion"]; ok {
		return false
	}
	_, hasHostType := raw["host_type"]
	_, hasSettings := raw["settings"]
	return hasHostType || (!hasSettings && (raw["email"] != nil || raw["backend"] != nil || raw["folder_name"] != nil))
}

func migrateFlatConfigFile(path string, data []byte, hooks MigrateHooks) (*Config, error) {
	ok, err := approveMigration(hooks)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("config migration declined; refuse to run with legacy config (re-run with confirmation or --yes)")
	}

	flat, err := parseLegacyFlat(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse legacy config: %w", err)
	}

	cfg, warnBW := mapFlatToV2(flat)
	if warnBW && hooks.Warn != nil {
		hooks.Warn("legacy backend=bw was dropped; v0.20+ uses the API backend only")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("migrated config is invalid: %w", err)
	}

	if err := backupConfigFile(path, data); err != nil {
		return nil, err
	}
	if err := SaveConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to save migrated config: %w", err)
	}
	return cfg, nil
}

func approveMigration(hooks MigrateHooks) (bool, error) {
	if hooks.Yes {
		return true, nil
	}
	if hooks.Confirm == nil {
		return false, fmt.Errorf("legacy config detected; confirm migration interactively or pass --yes")
	}
	return hooks.Confirm()
}

func parseLegacyFlat(data []byte) (*legacyFlatConfig, error) {
	var flat legacyFlatConfig
	if err := UnmarshalJSONC(data, &flat); err != nil {
		return nil, err
	}
	return &flat, nil
}

func mapFlatToV2(flat *legacyFlatConfig) (*Config, bool) {
	now := nowISO8601()
	host := Host{
		ID:            DefaultHostID,
		IsDefault:     true,
		TargetSection: DefaultFolderName,
	}
	ht := strings.TrimSpace(flat.HostType)
	switch ht {
	case "selfhosted":
		host.Type = HostTypeSelfhost
		host.HostURL = strings.TrimSpace(flat.SelfhostedURL)
	default:
		host.Type = HostTypeCloud
		host.HostURL = DefaultCloudURL
	}
	if email := strings.TrimSpace(flat.Email); email != "" {
		host.Email = email
	}
	if folder := strings.TrimSpace(flat.FolderName); folder != "" {
		host.TargetSection = folder
	}
	if dev := strings.TrimSpace(flat.DeviceIdentifier); dev != "" {
		host.DeviceIdentifier = dev
	}

	warnBW := strings.TrimSpace(flat.Backend) == "bw"
	cfg := &Config{
		SchemaVersion: SchemaVersion1,
		CreatedAt:     now,
		UpdatedAt:     now,
		AppVersion:    AppVersion,
		Settings: GlobalSettings{
			Hosts: []Host{host},
		},
	}
	return cfg, warnBW
}

func backupConfigFile(path string, data []byte) error {
	ts := time.Now().UTC().Format("20060102T150405Z")
	bak := path + ".bak-" + ts
	if err := writeFile(bak, data, 0600); err != nil {
		return fmt.Errorf("failed to write config backup: %w", err)
	}
	return nil
}
