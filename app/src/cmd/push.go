package cmd

import (
	"bwsf/src/config"
	"bwsf/src/core"
	"bwsf/src/infra"
	"bwsf/src/utils"
	"os"

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

	utils.Infoln("[INFO] Found", len(envFiles), "managed file(s) to push:")
	for _, f := range envFiles {
		utils.Infoln("  -", f)
	}

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
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}

	utils.Successln("[INFO] ✅", len(envFiles), "managed file(s) pushed successfully!")
}
