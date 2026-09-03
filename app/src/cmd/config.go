package cmd

import (
	"fmt"

	"bwsf/src/config"
	"bwsf/src/utils"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Inspect bwsf configuration",
	Long:  "Inspect local bwsf configuration stored under ~/.config/bwsf/",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  "Show the current bwsf configuration values (host type, URL, email, folder name)",
	Run:   runConfigShow,
}

func init() {
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) {
	text, err := config.LoadConfigShowText()
	if err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
		return
	}
	fmt.Print(text)
}
