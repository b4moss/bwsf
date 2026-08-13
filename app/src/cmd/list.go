package cmd

import (
	"bwsf/src/core"
	"bwsf/src/infra"
	"bwsf/src/utils"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List items in the dotenvs folder",
	Long:  "List all items in the dotenvs folder from Bitwarden",
	Run:   runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	cfg := loadConfigOrEmpty()
	requireBwCLIIfNeeded(cfg)

	// Create dependencies
	bw := newBwClientFromConfig(cfg)
	defer clearAPISession(bw)
	logger := infra.NewLogger()

	// Call core logic
	items, err := core.ListDotenvsCore(
		bw,
		cfg,
		utils.InputPassword,
		logger,
	)
	if err != nil {
		reportCommandError(err)
		os.Exit(1)
	}

	// Output item names (one per line)
	if len(items) == 0 {
		fmt.Println("No items found in dotenvs folder")
		return
	}

	for _, item := range items {
		fmt.Println(item.Name)
	}
}
