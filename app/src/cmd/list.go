package cmd

import (
	"bwsf/src/config"
	"bwsf/src/core"
	"bwsf/src/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List items in the configured Bitwarden folder",
	Long:  "List all items in the configured Bitwarden folder (default: dotenvs)",
	Run:   runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	installed, _ := checkBwInstalled()
	if !installed {
		utils.Errorln("[ERROR] ❌ bw command is not installed...")
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
	logger := newLogger()

	sessions := newSessionStore()
	items, err := core.ListDotenvsCore(
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

	if len(items) == 0 {
		folderName := config.ResolveFolderName(cfg)
		fmt.Printf("No items found in %s folder\n", folderName)
		return
	}

	for _, item := range items {
		fmt.Println(item.Name)
	}
}
