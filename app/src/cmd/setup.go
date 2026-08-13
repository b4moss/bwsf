package cmd

import (
	"bwsf/src/config"
	"bwsf/src/core"
	"bwsf/src/infra"
	"bwsf/src/utils"
	"os"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup Bitwarden host configuration",
	Long:  "Configure Bitwarden host (Cloud or Self-hosted). For backend=bw, also login. For backend=api, use `bwsf auth` after setup.",
	Run:   runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) {
	cfg := loadConfigOrEmpty()
	if cfg.GetBackend() == config.BackendAPI {
		runSetupAPI()
		return
	}
	runSetupBW(cfg)
}

// runSetupAPI configures host/email only. Auth is via `bwsf auth`.
func runSetupAPI() {
	utils.Infoln("[INFO] backend=api: setup configures host/email only.")
	utils.Infoln("[INFO] Use `bwsf auth` afterward to store a Personal API Key and obtain an Identity token.")

	logger := infra.NewLogger()
	err := core.SetupAPIConfigCore(
		logger,
		utils.SelectHostType,
		utils.InputURL,
		utils.InputEmail,
	)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}

	utils.Successln("[INFO] ✅ Configuration saved. Run `bwsf auth` to authenticate for API backend.")
}

// runSetupBW keeps the existing CLI Login + folder setup flow.
func runSetupBW(cfg *config.Config) {
	requireBwCLIIfNeeded(cfg)

	bw := newBwClientFromConfig(cfg)
	fs := infra.NewFileSystem()
	logger := infra.NewLogger()

	confirmCreateFolder := func() (bool, error) {
		return utils.ConfirmYesNo("dotenvs folder not found. Create it? (y/N): ")
	}

	err := core.SetupBitwardenCore(
		fs,
		bw,
		logger,
		utils.SelectHostType,
		utils.InputURL,
		utils.InputEmail,
		utils.InputPassword,
		confirmCreateFolder,
	)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}

	utils.Successln("[INFO] ✅ Sign in to Bitwarden was successful!")
}
