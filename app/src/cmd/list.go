package cmd

import (
	"fmt"

	"bwsf/src/config"
	"bwsf/src/core"

	"github.com/spf13/cobra"
)

var listHost string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List items in the configured Bitwarden folder",
	Long:  "List all items in the configured Bitwarden folder (default: dotenvs)",
	Run:   runList,
}

func init() {
	listCmd.Flags().StringVar(&listHost, "host", "", "Host id from global config hosts[]")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	cfg := loadConfigOrEmpty()

	host := resolveHostForCommand(cfg, listHost, "")
	bw := newBwClientForHost(cfg, host)
	defer clearAPISession(bw)
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
		reportCommandError(err)
		exitFunc(1)
		return
	}

	if len(items) == 0 {
		folderName := host.TargetSection
		if folderName == "" {
			folderName = config.ResolveFolderName(cfg)
		}
		fmt.Printf("No items found in %s folder\n", folderName)
		return
	}

	for _, item := range items {
		fmt.Println(item.Name)
	}
}
