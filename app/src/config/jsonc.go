package config

import (
	"encoding/json"
	"fmt"

	"github.com/tailscale/hujson"
)

// UnmarshalConfigJSONC parses JSON or JSONC into Config using hujson
// (comments and trailing commas allowed), then encoding/json.
func UnmarshalConfigJSONC(data []byte, dst *Config) error {
	if dst == nil {
		return fmt.Errorf("config destination is nil")
	}
	standardized, err := hujson.Standardize(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(standardized, dst)
}
