package cmd

import (
	"errors"

	"bwsf/src/config"
	"bwsf/src/core"
	"bwsf/src/infra"
	"bwsf/src/utils"
)

type sessionClearer interface {
	ClearSession()
}

func clearAPISession(bw core.BwClient) {
	if c, ok := bw.(sessionClearer); ok {
		c.ClearSession()
	}
}

func reportCommandError(err error) {
	if err == nil {
		return
	}
	if core.IsNotAuthenticatedError(err) || errors.Is(err, infra.ErrAPINotAuthenticated) {
		utils.Errorln("[ERROR]", err)
		utils.Infoln("[INFO] Run: bwsf auth")
		return
	}
	utils.Errorln("[ERROR]", err)
}

// loadConfigOrEmpty loads ~/.config/bwsf/config.json, or returns an empty config when missing.
func loadConfigOrEmpty() *config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		utils.Errorln("[ERROR] Failed to load config:", err)
		exitFunc(1)
	}
	if cfg == nil {
		return &config.Config{}
	}
	return cfg
}

// newBwClientFromConfig builds the BwClient for cfg. Overridable in tests.
var newBwClientFromConfig = func(cfg *config.Config) core.BwClient {
	bw, err := infra.NewBwClientFromConfig(cfg)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
	}
	return bw
}

// requireBwCLIIfNeeded exits when the bw CLI is required but not installed.
func requireBwCLIIfNeeded(cfg *config.Config) {
	if cfg.GetBackend() != config.BackendBW {
		return
	}
	installed, _ := checkBwInstalled()
	if !installed {
		utils.Errorln("[ERROR] ❌ bw command is not installed...")
		exitFunc(1)
	}
}
