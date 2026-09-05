package infra

import (
	"bwsf/src/config"
	"bwsf/src/core"
)

// NewBwClientFromConfig returns the API BwClient for cfg (v0.20+: API only).
// When cfg is nil, an empty config is used; callers should resolve a host first.
func NewBwClientFromConfig(cfg *config.Config) (core.BwClient, error) {
	return NewApiBwClient(cfg), nil
}

// NewBwClientForHost returns an API client bound to a specific host.
func NewBwClientForHost(cfg *config.Config, host *config.Host) (core.BwClient, error) {
	return NewApiBwClientForHost(cfg, host), nil
}
