package cmd

import (
	"errors"

	"bwsf/src/core"
	"bwsf/src/utils"

	"github.com/spf13/cobra"
)

var cleanHost string

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove local managed files after verifying Bitwarden backup",
	Long:  "Remove bwsf-managed local files (.env*, *.tfvars, *.tfvars.json) after confirming the remote Bitwarden note item has matching (or explicitly handled) contents",
	Run:   runClean,
}

func init() {
	cleanCmd.Flags().String("from", ".", "Directory containing managed files to clean")
	cleanCmd.Flags().StringVar(&cleanHost, "host", "", "Host id from global config hosts[]")
	rootCmd.AddCommand(cleanCmd)
}

func runClean(cmd *cobra.Command, args []string) {
	cfg := loadConfigOrEmpty()

	projectName, fromDir, filter, projectHost, err := resolveProjectAndFileDir(cmd, "from")
	if err != nil {
		utils.Errorln("[ERROR] Failed to resolve project directory:", err)
		exitFunc(1)
		return
	}

	host := resolveHostForCommand(cfg, cleanHost, projectHost)
	bw := newBwClientForHost(cfg, host)
	defer clearAPISession(bw)
	fs := newFileSystem()
	logger := newLogger()

	envFiles, err := core.GetPushedEnvFiles(fromDir, filter, fs)
	if err != nil {
		utils.Errorln("[ERROR] Failed to find managed files:", err)
		exitFunc(1)
		return
	}
	if len(envFiles) == 0 {
		utils.Errorln("[ERROR] No managed files found")
		exitFunc(1)
		return
	}

	utils.Infoln("[INFO] Found", len(envFiles), "managed file(s) to clean:")
	for _, f := range envFiles {
		utils.Infoln("  -", f)
	}

	selectMismatchAction := func(mismatchedFiles []string) (core.CleanMismatchAction, error) {
		choice, selectErr := selectCleanMismatch(mismatchedFiles)
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

	sessions := newSessionStore()
	err = core.CleanEnvCore(
		fromDir,
		projectName,
		filter,
		fs,
		bw,
		cfg,
		inputPassword,
		selectMismatchAction,
		logger,
		sessions,
	)
	if err != nil {
		if errors.Is(err, core.ErrCleanAborted) {
			utils.Infoln("[INFO] Clean aborted. Local files were kept.")
			return
		}
		reportCommandError(err)
		exitFunc(1)
		return
	}

	utils.Successln("[INFO] ✅", len(envFiles), "managed file(s) cleaned successfully!")
}
