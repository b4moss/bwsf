package cmd

import (
	"bwsf/src/utils"

	"github.com/spf13/cobra"
)

var backendCmd = &cobra.Command{
	Use:   "backend",
	Short: "Backend command (removed)",
	Long:  "The backend command was removed in v0.20. bwsf uses the API backend only.",
	Run:   runBackend,
}

func init() {
	rootCmd.AddCommand(backendCmd)
}

func runBackend(cmd *cobra.Command, args []string) {
	utils.Errorln("[ERROR] backend command removed; API only")
	exitFunc(1)
}
