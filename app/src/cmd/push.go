package cmd

import (
	"bwsf/src/core"
	"bwsf/src/infra"
	"bwsf/src/utils"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push .env file to Bitwarden",
	Long:  "Push .env file from specified directory to Bitwarden as a note item",
	Run:   runPush,
}

func init() {
	pushCmd.Flags().String("from", ".", "Directory containing .env file")
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) {
	cfg := loadConfigOrEmpty()
	requireBwCLIIfNeeded(cfg)

	// Get --from flag value
	fromDir, err := cmd.Flags().GetString("from")
	if err != nil {
		utils.Errorln("[ERROR] Failed to get --from flag:", err)
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

	// Get list of env files to be pushed
	envFiles, err := core.GetPushedEnvFiles(fromDir, fs)
	if err != nil {
		utils.Errorln("[ERROR] Failed to find .env files:", err)
		os.Exit(1)
	}

	if len(envFiles) == 0 {
		utils.Errorln("[ERROR] No .env files found")
		os.Exit(1)
	}

	// Display files to be pushed
	utils.Infoln("[INFO] Found", len(envFiles), "env file(s) to push:")
	for _, f := range envFiles {
		utils.Infoln("  -", f)
	}

	// Call core logic
	err = core.PushEnvCore(
		fromDir,
		projectName,
		fs,
		bw,
		cfg,
		utils.InputPassword,
		logger,
	)
	if err != nil {
		reportCommandError(err)
		os.Exit(1)
	}

	utils.Successln("[INFO] ✅", len(envFiles), "env file(s) pushed successfully!")
}
