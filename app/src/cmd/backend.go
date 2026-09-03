package cmd

import (
	"fmt"
	"os"

	"bwsf/src/config"
	"bwsf/src/utils"

	"github.com/spf13/cobra"
)

var backendSet string

var backendCmd = &cobra.Command{
	Use:   "backend",
	Short: "Show or set the Bitwarden backend",
	Long:  "Show the current Bitwarden backend (bw or api), or update it with --set",
	Run:   runBackend,
}

func init() {
	backendCmd.Flags().StringVar(&backendSet, "set", "", "Set backend to \"bw\" or \"api\"")
	rootCmd.AddCommand(backendCmd)
}

func runBackend(cmd *cobra.Command, args []string) {
	if backendSet != "" {
		if err := setBackend(backendSet); err != nil {
			utils.Errorln("[ERROR]", err)
			os.Exit(1)
		}
		utils.Successln("[INFO] ✅ Backend set to " + backendSet)
		return
	}

	backend, err := currentBackend()
	if err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}
	fmt.Println(backend)
}

func currentBackend() (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	return cfg.GetBackend(), nil
}

func setBackend(value string) error {
	if !config.IsValidBackend(value) {
		return fmt.Errorf("invalid backend %q: use %q or %q", value, config.BackendBW, config.BackendAPI)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg == nil {
		cfg = &config.Config{}
	}

	cfg.Backend = value
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}
