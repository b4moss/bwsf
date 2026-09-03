package cmd

import (
	"bwsf/src/config"
	"bwsf/src/core"
	"bwsf/src/utils"

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
	installed, _ := checkBwInstalled()
	if !installed {
		utils.Errorln("[ERROR] ❌ bw command is not installed...")
		exitFunc(1)
		return
	}

	projectName, fromDir, filter, err := resolveProjectAndFileDir(cmd, "from")
	if err != nil {
		utils.Errorln("[ERROR] Failed to resolve project directory:", err)
		exitFunc(1)
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		utils.Errorln("[ERROR] Failed to load config:", err)
		exitFunc(1)
		return
	}
	if cfg == nil {
		cfg = &config.Config{}
	}

	bw := newBwClient()
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

	utils.Infoln("[INFO] Found", len(envFiles), "managed file(s) to push:")
	for _, f := range envFiles {
		utils.Infoln("  -", f)
	}

	sessions := newSessionStore()
	err = core.PushEnvCore(
		fromDir,
		projectName,
		filter,
		fs,
		bw,
		cfg,
		inputPassword,
		logger,
		sessions,
	)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
		return
	}

	utils.Successln("[INFO] ✅", len(envFiles), "managed file(s) pushed successfully!")
}
