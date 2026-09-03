package cmd

import (
	"fmt"
	"strings"

	"bwsf/src/config"
	"bwsf/src/core"
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
	Long:  "Configure Bitwarden host (Cloud or Self-hosted) and login credentials",
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
	installed, _ := checkBwInstalled()
	if !installed {
		utils.Errorln("[ERROR] ❌ bw command is not installed...")
		exitFunc(1)
		return
	}

	if err := validateSetupNonInteractiveFlags(); err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
		return
	}

	folderName := config.DefaultFolderName

	// Persist --folder before core setup so RealBwClient reads it for folder ops.
	if setupFolder != "" {
		if err := config.ValidateFolderName(setupFolder); err != nil {
			utils.Errorln("[ERROR]", err)
			exitFunc(1)
			return
		}
		folderName = strings.TrimSpace(setupFolder)

		cfg, err := config.LoadConfig()
		if err != nil {
			utils.Errorln("[ERROR] Failed to load config:", err)
			exitFunc(1)
			return
		}
		if cfg == nil {
			cfg = &config.Config{}
		}
		cfg.FolderName = folderName
		if err := config.SaveConfig(cfg); err != nil {
			utils.Errorln("[ERROR] Failed to save folder name:", err)
			exitFunc(1)
			return
		}
	} else {
		cfg, err := config.LoadConfig()
		if err != nil {
			utils.Errorln("[ERROR] Failed to load config:", err)
			exitFunc(1)
			return
		}
		folderName = config.ResolveFolderName(cfg)
	}

	bw := newBwClient()
	fs := newFileSystem()
	logger := newLogger()

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

	err := core.SetupBitwardenCore(
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
		exitFunc(1)
		return
	}

	utils.Successln("[INFO] ✅ Sign in to Bitwarden was successful!")
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
