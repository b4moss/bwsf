package cmd

import (
	"bwsf/src/core"
	"bwsf/src/utils"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull .env file from Bitwarden",
	Long:  "Pull .env file from Bitwarden and save it to the current directory or specified directory",
	Run:   runPull,
}

func init() {
	pullCmd.Flags().String("output", ".", "Directory to save .env file")
	rootCmd.AddCommand(pullCmd)
}

func runPull(cmd *cobra.Command, args []string) {
	cfg := loadConfigOrEmpty()
	requireBwCLIIfNeeded(cfg)

	projectName, outputDir, _, err := resolveProjectAndFileDir(cmd, "output")
	if err != nil {
		utils.Errorln("[ERROR] Failed to resolve project directory:", err)
		exitFunc(1)
		return
	}

	bw := newBwClientFromConfig(cfg)
	defer clearAPISession(bw)
	fs := newFileSystem()
	logger := newLogger()

	sessions := newSessionStore()
	envFiles, err := core.GetPulledEnvFiles(projectName, bw, cfg, inputPassword, logger, sessions)
	if err != nil {
		reportCommandError(err)
		exitFunc(1)
		return
	}

	if len(envFiles) == 0 {
		utils.Errorln("[ERROR] No env files found in Bitwarden for project:", projectName)
		exitFunc(1)
		return
	}

	utils.Infoln("[INFO] Found", len(envFiles), "env file(s) to pull:")
	for _, f := range envFiles {
		utils.Infoln("  -", f)
	}

	doConfirm := func(path string) (bool, error) {
		return confirmOverwrite(fmt.Sprintf("%s already exists. Overwrite? (y/N): ", filepath.Base(path)))
	}

	err = core.PullEnvCore(
		outputDir,
		projectName,
		fs,
		bw,
		cfg,
		inputPassword,
		doConfirm,
		logger,
		sessions,
	)
	if err != nil {
		reportCommandError(err)
		exitFunc(1)
		return
	}

	utils.Successln("[INFO] ✅", len(envFiles), "env file(s) pulled successfully!")
}
