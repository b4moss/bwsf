package config

import (
	"encoding/json"
	"fmt"

	"github.com/tailscale/hujson"
)

// UnmarshalJSONC parses JSON or JSONC into dst using hujson
// (comments and trailing commas allowed), then encoding/json.
func UnmarshalJSONC(data []byte, dst any) error {
	if dst == nil {
		return fmt.Errorf("destination is nil")
	}
	standardized, err := hujson.Standardize(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(standardized, dst)
}

// standardizeJSONC returns strict JSON bytes from JSONC input.
func standardizeJSONC(data []byte) ([]byte, error) {
	return hujson.Standardize(data)
}

// UnmarshalConfigJSONC parses JSONC into Config and rejects banned keys.
func UnmarshalConfigJSONC(data []byte, dst *Config) error {
	if dst == nil {
		return fmt.Errorf("config destination is nil")
	}
	if err := detectBannedKeys(data); err != nil {
		return err
	}
	if err := UnmarshalJSONC(data, dst); err != nil {
		return err
	}
	dst.Normalize()
	return nil
}
