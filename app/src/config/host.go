package config

import (
	"fmt"
	"strings"
)

// ResolveHost picks a host: CLI --host → project host → is_default.
func ResolveHost(cfg *Config, cliHost, projectHost string) (*Host, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil; run `bwsf setup` first")
	}
	cliHost = strings.TrimSpace(cliHost)
	projectHost = strings.TrimSpace(projectHost)

	id := cliHost
	if id == "" {
		id = projectHost
	}
	if id != "" {
		h := cfg.FindHost(id)
		if h == nil {
			return nil, fmt.Errorf("host %q not found in global config hosts", id)
		}
		return h, nil
	}

	if len(cfg.Settings.Hosts) == 0 {
		return nil, fmt.Errorf("no hosts configured; run `bwsf setup` to add a host, or pass --host / project host")
	}
	h := cfg.DefaultHost()
	if h == nil {
		return nil, fmt.Errorf("no default host; set is_default on exactly one host or pass --host")
	}
	return h, nil
}
