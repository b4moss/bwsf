package cmd

import (
	"bwsf/src/core"
	"bwsf/src/infra"
	"bwsf/src/utils"
	"fmt"
	"os"
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

	// Get --output flag value
	outputDir, err := cmd.Flags().GetString("output")
	if err != nil {
		utils.Errorln("[ERROR] Failed to get --output flag:", err)
		os.Exit(1)
	}

	// Get current working directory name as project name
	wd, err := os.Getwd()
	if err != nil {
		utils.Errorln("[ERROR] Failed to get current working directory:", err)
		os.Exit(1)
	}
	projectName := filepath.Base(wd)

	// Create dependencies
	bw := newBwClientFromConfig(cfg)
	defer clearAPISession(bw)
	fs := infra.NewFileSystem()
	logger := infra.NewLogger()

	// Get list of env files to be pulled
	envFiles, err := core.GetPulledEnvFiles(projectName, bw, cfg, utils.InputPassword, logger)
	if err != nil {
		reportCommandError(err)
		os.Exit(1)
	}

	if len(envFiles) == 0 {
		utils.Errorln("[ERROR] No env files found in Bitwarden for project:", projectName)
		os.Exit(1)
	}

	// Display files to be pulled
	utils.Infoln("[INFO] Found", len(envFiles), "env file(s) to pull:")
	for _, f := range envFiles {
		utils.Infoln("  -", f)
	}

	// confirmOverwrite wrapper
	confirmOverwrite := func(path string) (bool, error) {
		return utils.ConfirmOverwrite(fmt.Sprintf("%s already exists. Overwrite? (y/N): ", filepath.Base(path)))
	}

	// Call core logic
	err = core.PullEnvCore(
		outputDir,
		projectName,
		fs,
		bw,
		cfg,
		utils.InputPassword,
		confirmOverwrite,
		logger,
	)
	if err != nil {
		reportCommandError(err)
		os.Exit(1)
	}

	utils.Successln("[INFO] ✅", len(envFiles), "env file(s) pulled successfully!")
}
