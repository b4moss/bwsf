package cmd

import (
	"bwsf/src/utils"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "bwsf",
	Short:   "CLI tool to manage .env files using Bitwarden",
	Long:    "bwsf is a CLI tool that uses Bitwarden to manage .env files",
	Version: Version,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&migrateYes, "yes", false, "Assume yes for confirmations (e.g. config migration)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		utils.Error("Error: %v\n", err)
		os.Exit(1)
	}
}
