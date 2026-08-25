package cmd

import (
	"bwsf/src/config"
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
	installed, _ := utils.CheckBwCommand()
	if !installed {
		utils.Errorln("[ERROR] ❌ bw command is not installed...")
		os.Exit(1)
	}

	projectName, outputDir, err := resolveProjectAndFileDir(cmd, "output")
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

	envFiles, err := core.GetPulledEnvFiles(projectName, bw, cfg, utils.InputPassword, logger)
	if err != nil {
		utils.Errorln("[ERROR] Failed to get env files info:", err)
		os.Exit(1)
	}

	if len(envFiles) == 0 {
		utils.Errorln("[ERROR] No env files found in Bitwarden for project:", projectName)
		os.Exit(1)
	}

	utils.Infoln("[INFO] Found", len(envFiles), "env file(s) to pull:")
	for _, f := range envFiles {
		utils.Infoln("  -", f)
	}

	confirmOverwrite := func(path string) (bool, error) {
		return utils.ConfirmOverwrite(fmt.Sprintf("%s already exists. Overwrite? (y/N): ", filepath.Base(path)))
	}

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
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}

	utils.Successln("[INFO] ✅", len(envFiles), "env file(s) pulled successfully!")
}
