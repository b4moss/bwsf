package cmd

import (
	"bwsf/src/config"
	"bwsf/src/core"
	"bwsf/src/infra"
	"bwsf/src/utils"
	"fmt"
	"os"

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
	// Check if bw command is installed
	installed, _ := utils.CheckBwCommand()
	if !installed {
		utils.Errorln("[ERROR] ❌ bw command is not installed...")
		os.Exit(1)
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		utils.Errorln("[ERROR] Failed to load config:", err)
		os.Exit(1)
	}
	if cfg == nil {
		cfg = &config.Config{}
	}

	// Create dependencies
	bw := infra.NewBwClient()
	logger := infra.NewLogger()

	// Call core logic
	sessions := infra.NewSessionStore()
	items, err := core.ListDotenvsCore(
		bw,
		cfg,
		utils.InputPassword,
		logger,
		sessions,
	)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}

	// Output item names (one per line)
	if len(items) == 0 {
		folderName := config.ResolveFolderName(cfg)
		fmt.Printf("No items found in %s folder\n", folderName)
		return
	}

	for _, item := range items {
		fmt.Println(item.Name)
	}
}
