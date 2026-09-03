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

// UnmarshalConfigJSONC parses JSON or JSONC into Config.
func UnmarshalConfigJSONC(data []byte, dst *Config) error {
	if dst == nil {
		return fmt.Errorf("config destination is nil")
	}
	return UnmarshalJSONC(data, dst)
}
