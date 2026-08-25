package cmd

import (
	"bwsf/src/project"
	"bwsf/src/utils"
	"os"

	"github.com/spf13/cobra"
)

const gitRootFallbackWarn = "[WARN] No .git found in ancestor directories; falling back to current directory"

// resolveProjectAndFileDir resolves Bitwarden Note name and the file Dir for
// push/pull/clean according to Issue #134 (git root vs cwd fallback + flag).
func resolveProjectAndFileDir(cmd *cobra.Command, flagName string) (projectName, fileDir string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	ctx, err := project.Resolve(wd, "")
	if err != nil {
		return "", "", err
	}
	if ctx.UsedCwdFallback {
		utils.Warningln(gitRootFallbackWarn)
	}

	flagValue, err := cmd.Flags().GetString(flagName)
	if err != nil {
		return "", "", err
	}

	fileDir = project.SelectFileDir(cmd.Flags().Changed(flagName), flagValue, ctx.WorkDir)
	return ctx.ProjectName, fileDir, nil
}
