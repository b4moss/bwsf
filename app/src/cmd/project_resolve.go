package cmd

import (
	"bwsf/src/config"
	"bwsf/src/core"
	"bwsf/src/project"
	"bwsf/src/utils"
	"os"

	"github.com/spf13/cobra"
)

const gitRootFallbackWarn = "[WARN] No .git found in ancestor directories; falling back to current directory"

// resolveProjectAndFileDir resolves Bitwarden Note name, file Dir, managed-file
// filter, and project host id for push/pull/clean.
func resolveProjectAndFileDir(cmd *cobra.Command, flagName string) (projectName, fileDir string, filter core.ManagedFileFilter, projectHost string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", "", filter, "", err
	}

	globalCfg, _ := config.LoadConfigWithMigrate(migrateHooks())

	pc, _, err := config.ResolveProjectConfig(wd, utils.SelectProjectConfigPath)
	if err != nil {
		return "", "", filter, "", err
	}
	override := ""
	if pc != nil {
		override = pc.EffectiveOverride()
		projectHost = pc.EffectiveHost()
	}
	filter = core.ManagedFileFilter{
		SaveFiles: EffectiveManagedSaveFiles(globalCfg, pc),
	}

	ctx, err := project.Resolve(wd, override)
	if err != nil {
		return "", "", filter, "", err
	}
	if ctx.UsedCwdFallback {
		utils.Warningln(gitRootFallbackWarn)
	}

	flagValue, err := cmd.Flags().GetString(flagName)
	if err != nil {
		return "", "", filter, "", err
	}

	fileDir = project.SelectFileDir(cmd.Flags().Changed(flagName), flagValue, ctx.WorkDir)
	return ctx.ProjectName, fileDir, filter, projectHost, nil
}
