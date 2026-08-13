package cmd

import (
	"fmt"
	"os"
	"strings"

	"bwsf/src/config"
	"bwsf/src/core"
	"bwsf/src/infra"
	"bwsf/src/utils"

	"github.com/spf13/cobra"
)

var setupFolder string
var setupHostType string
var setupURL string
var setupEmail string
var setupPassword string
var setupYes bool

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup Bitwarden host configuration",
	Long:  "Configure Bitwarden host (Cloud or Self-hosted). For backend=bw, also login. For backend=api, use `bwsf auth` after setup.",
	Run:   runSetup,
}

func init() {
	setupCmd.Flags().StringVar(&setupFolder, "folder", "", "Bitwarden folder name for .env notes (default: dotenvs)")
	setupCmd.Flags().StringVar(&setupHostType, "host-type", "", "Host type: cloud or selfhosted (non-interactive)")
	setupCmd.Flags().StringVar(&setupURL, "url", "", "Self-hosted server URL (required when --host-type=selfhosted)")
	setupCmd.Flags().StringVar(&setupEmail, "email", "", "Account email (non-interactive)")
	setupCmd.Flags().StringVar(&setupPassword, "password", "", "Master password (non-interactive)")
	setupCmd.Flags().BoolVar(&setupYes, "yes", false, "Assume yes for confirmations (e.g. create folder)")
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

	if _, err := applySetupFolderFlag(); err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}

	selectHostType := utils.SelectHostType
	inputURL := utils.InputURL
	inputEmail := utils.InputEmail
	if nonInteractiveSetup() {
		if setupHostType == "" || setupEmail == "" {
			utils.Errorln("[ERROR] non-interactive API setup requires --host-type and --email")
			os.Exit(1)
		}
		if setupHostType != "cloud" && setupHostType != "selfhosted" {
			utils.Errorln("[ERROR] --host-type must be cloud or selfhosted")
			os.Exit(1)
		}
		if setupHostType == "selfhosted" && strings.TrimSpace(setupURL) == "" {
			utils.Errorln("[ERROR] --url is required when --host-type=selfhosted")
			os.Exit(1)
		}
		selectHostType = func() (string, error) { return setupHostType, nil }
		inputURL = func() (string, error) { return setupURL, nil }
		inputEmail = func() (string, error) { return setupEmail, nil }
	}

	logger := infra.NewLogger()
	err := core.SetupAPIConfigCore(
		logger,
		selectHostType,
		inputURL,
		inputEmail,
	)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}

	cfg := loadConfigOrEmpty()
	folderName := config.ResolveFolderName(cfg)
	bw := newBwClientFromConfig(cfg)
	defer clearAPISession(bw)

	inputPassword := utils.InputPassword
	if setupPassword != "" {
		inputPassword = func() (string, error) { return setupPassword, nil }
	}
	confirmCreateFolder := func() (bool, error) {
		return utils.ConfirmYesNo(fmt.Sprintf("%s folder not found. Create it? (y/N): ", folderName))
	}
	if setupYes {
		confirmCreateFolder = func() (bool, error) { return true, nil }
	}

	if err := core.EnsureConfiguredFolderCore(bw, cfg, logger, inputPassword, confirmCreateFolder); err != nil {
		if core.IsNotAuthenticatedError(err) {
			utils.Infoln("[INFO] Folder check skipped until `bwsf auth` (and unlock) succeeds.")
		} else {
			utils.Errorln("[ERROR]", err)
			os.Exit(1)
		}
	}

	utils.Successln("[INFO] ✅ Configuration saved. Run `bwsf auth` to authenticate for API backend.")
}

// runSetupBW keeps the existing CLI Login + folder setup flow.
func runSetupBW(cfg *config.Config) {
	requireBwCLIIfNeeded(cfg)

	if err := validateSetupNonInteractiveFlags(); err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}

	folderName, err := applySetupFolderFlag()
	if err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}

	bw := newBwClientFromConfig(cfg)
	fs := infra.NewFileSystem()
	logger := infra.NewLogger()

	selectHostType := utils.SelectHostType
	inputURL := utils.InputURL
	inputEmail := utils.InputEmail
	inputPassword := utils.InputPassword
	confirmCreateFolder := func() (bool, error) {
		return utils.ConfirmYesNo(fmt.Sprintf("%s folder not found. Create it? (y/N): ", folderName))
	}

	if nonInteractiveSetup() {
		selectHostType = func() (string, error) { return setupHostType, nil }
		inputURL = func() (string, error) { return setupURL, nil }
		inputEmail = func() (string, error) { return setupEmail, nil }
		inputPassword = func() (string, error) { return setupPassword, nil }
		confirmCreateFolder = func() (bool, error) { return setupYes, nil }
	}

	err = core.SetupBitwardenCore(
		fs,
		bw,
		logger,
		selectHostType,
		inputURL,
		inputEmail,
		inputPassword,
		confirmCreateFolder,
	)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}

	utils.Successln("[INFO] ✅ Sign in to Bitwarden was successful!")
}

func applySetupFolderFlag() (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	if cfg == nil {
		cfg = &config.Config{}
	}

	if setupFolder != "" {
		if err := config.ValidateFolderName(setupFolder); err != nil {
			return "", err
		}
		cfg.FolderName = strings.TrimSpace(setupFolder)
		if err := config.SaveConfig(cfg); err != nil {
			return "", fmt.Errorf("failed to save folder name: %w", err)
		}
	}

	return config.ResolveFolderName(cfg), nil
}

func nonInteractiveSetup() bool {
	return setupHostType != "" || setupURL != "" || setupEmail != "" || setupPassword != ""
}

func validateSetupNonInteractiveFlags() error {
	if !nonInteractiveSetup() {
		return nil
	}
	if setupHostType == "" || setupEmail == "" || setupPassword == "" {
		return fmt.Errorf("non-interactive setup requires --host-type, --email, and --password")
	}
	if setupHostType != "cloud" && setupHostType != "selfhosted" {
		return fmt.Errorf("--host-type must be cloud or selfhosted")
	}
	if setupHostType == "selfhosted" && strings.TrimSpace(setupURL) == "" {
		return fmt.Errorf("--url is required when --host-type=selfhosted")
	}
	return nil
}
