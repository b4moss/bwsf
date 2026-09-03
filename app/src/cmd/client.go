package cmd

import (
	"errors"

	"bwsf/src/config"
	"bwsf/src/core"
	"bwsf/src/infra"
	"bwsf/src/utils"
)

// migrateYes skips flat→v2 migration confirm when set (CLI --yes).
var migrateYes bool

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

func migrateHooks() config.MigrateHooks {
	return config.MigrateHooks{
		Yes: migrateYes,
		Confirm: func() (bool, error) {
			return utils.ConfirmYesNo("Legacy config detected. Migrate to v0.20 schema? (y/N): ")
		},
		Warn: func(msg string) {
			utils.Warningln("[WARN]", msg)
		},
	}
}

// loadConfigOrEmpty loads global config (with migrate hooks), or returns an empty v2 config.
func loadConfigOrEmpty() *config.Config {
	cfg, err := config.LoadConfigWithMigrate(migrateHooks())
	if err != nil {
		utils.Errorln("[ERROR] Failed to load config:", err)
		exitFunc(1)
	}
	if cfg == nil {
		return config.NewEmptyConfig()
	}
	return cfg
}

// newBwClientFromConfig builds the BwClient for cfg (default host). Overridable in tests.
var newBwClientFromConfig = func(cfg *config.Config) core.BwClient {
	bw, err := infra.NewBwClientFromConfig(cfg)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
	}
	return bw
}

// newBwClientForHost builds the BwClient for a resolved host. Overridable in tests.
var newBwClientForHost = func(cfg *config.Config, host *config.Host) core.BwClient {
	bw, err := infra.NewBwClientForHost(cfg, host)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
	}
	return bw
}

// requireBwCLIIfNeeded is a no-op in v0.20+ (API only; bw CLI not required).
func requireBwCLIIfNeeded(cfg *config.Config) {
	_ = cfg
}

// EffectiveManagedSaveFiles merges project then global save_files (#177).
func EffectiveManagedSaveFiles(global *config.Config, project *config.ProjectConfig) []string {
	if project != nil {
		if s := project.EffectiveSaveFiles(); len(s) > 0 {
			return s
		}
	}
	if global != nil {
		return global.EffectiveSaveFiles()
	}
	return nil
}

// resolveHostForCommand resolves CLI --host / project host / default.
func resolveHostForCommand(cfg *config.Config, cliHost, projectHost string) *config.Host {
	host, err := config.ResolveHost(cfg, cliHost, projectHost)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
		return nil
	}
	return host
}
