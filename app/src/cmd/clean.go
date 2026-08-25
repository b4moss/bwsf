package cmd

import (
	"bwsf/src/config"
	"bwsf/src/core"
	"bwsf/src/infra"
	"bwsf/src/utils"
	"errors"
	"os"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove local managed files after verifying Bitwarden backup",
	Long:  "Remove bwsf-managed local files (.env*, *.tfvars, *.tfvars.json) after confirming the remote Bitwarden note item has matching (or explicitly handled) contents",
	Run:   runClean,
}

func init() {
	cleanCmd.Flags().String("from", ".", "Directory containing managed files to clean")
	rootCmd.AddCommand(cleanCmd)
}

func runClean(cmd *cobra.Command, args []string) {
	installed, _ := utils.CheckBwCommand()
	if !installed {
		utils.Errorln("[ERROR] ❌ bw command is not installed...")
		os.Exit(1)
	}

	projectName, fromDir, err := resolveProjectAndFileDir(cmd, "from")
	if err != nil {
		utils.Errorln("[ERROR] Failed to resolve project directory:", err)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		utils.Errorln("[ERROR] Failed to load config:", err)
		os.Exit(1)
	}
	if cfg == nil {
		cfg = &config.Config{}
	}

	bw := infra.NewBwClient()
	fs := infra.NewFileSystem()
	logger := infra.NewLogger()

	envFiles, err := core.GetPushedEnvFiles(fromDir, fs)
	if err != nil {
		utils.Errorln("[ERROR] Failed to find managed files:", err)
		os.Exit(1)
	}
	if len(envFiles) == 0 {
		utils.Errorln("[ERROR] No managed files found")
		os.Exit(1)
	}

	utils.Infoln("[INFO] Found", len(envFiles), "managed file(s) to clean:")
	for _, f := range envFiles {
		utils.Infoln("  -", f)
	}

	selectMismatchAction := func(mismatchedFiles []string) (core.CleanMismatchAction, error) {
		choice, selectErr := utils.SelectCleanMismatchAction(mismatchedFiles)
		if selectErr != nil {
			return core.CleanActionAbort, selectErr
		}
		switch choice {
		case utils.CleanMismatchOverwriteRemoteThenClean:
			return core.CleanActionOverwriteRemoteThenClean, nil
		case utils.CleanMismatchRemoveLocal:
			return core.CleanActionRemoveLocal, nil
		default:
			return core.CleanActionAbort, nil
		}
	}

	err = core.CleanEnvCore(
		fromDir,
		projectName,
		fs,
		bw,
		cfg,
		utils.InputPassword,
		selectMismatchAction,
		logger,
	)
	if err != nil {
		if errors.Is(err, core.ErrCleanAborted) {
			utils.Infoln("[INFO] Clean aborted. Local files were kept.")
			return
		}
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}

	utils.Successln("[INFO] ✅", len(envFiles), "managed file(s) cleaned successfully!")
}
