package infra

import (
	"fmt"

	"bwsf/src/config"
	"bwsf/src/core"
)

// NewBwClientFromConfig selects a BwClient implementation based on cfg.Backend.
// When cfg is nil or Backend is unset, the API adapter is used (default backend).
func NewBwClientFromConfig(cfg *config.Config) (core.BwClient, error) {
	backend := cfg.GetBackend()

	switch backend {
	case config.BackendBW:
		return NewBwClient(), nil
	case config.BackendAPI:
		return NewApiBwClient(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported backend %q: use %q or %q", backend, config.BackendBW, config.BackendAPI)
	}
}
