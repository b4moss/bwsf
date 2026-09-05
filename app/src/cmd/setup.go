package cmd

import (
	"fmt"
	"strings"

	"bwsf/src/config"
	"bwsf/src/core"
	"bwsf/src/utils"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	setupFolder    string
	setupHostType  string
	setupURL       string
	setupEmail     string
	setupSkipHost  bool
	setupSaveFiles []string
)

// Interactive stubs (overridable in tests).
var (
	selectSetupHostAction = defaultSelectSetupHostAction
	selectSetupHostID     = defaultSelectSetupHostID
	selectSaveFilesAction = defaultSelectSaveFilesAction
	inputSaveFilesGlobs   = defaultInputSaveFilesGlobs
	confirmMakeDefault    = defaultConfirmMakeDefault
	inputNewHostID        = defaultInputNewHostID
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup Bitwarden host configuration",
	Long:  "Configure Bitwarden hosts and save_files (API only). Authentication is via `bwsf auth login` after setup.",
	Run:   runSetup,
}

func init() {
	setupCmd.Flags().StringVar(&setupFolder, "folder", "", "Target section / Bitwarden folder name (default: dotenvs)")
	setupCmd.Flags().StringVar(&setupHostType, "host-type", "", "Host type: cloud or selfhosted (non-interactive)")
	setupCmd.Flags().StringVar(&setupURL, "url", "", "Self-hosted server URL (required when --host-type=selfhosted)")
	setupCmd.Flags().StringVar(&setupEmail, "email", "", "Account email (non-interactive)")
	setupCmd.Flags().BoolVar(&setupSkipHost, "skip-host", false, "Skip host configuration (hosts: [] allowed)")
	setupCmd.Flags().StringSliceVar(&setupSaveFiles, "save-files", nil, "Global save_files globs (comma or repeat; ! prefix = exclude)")
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) {
	cfg := loadConfigOrEmpty()
	logger := newLogger()

	if nonInteractiveSetup() {
		if err := runSetupNonInteractive(cfg); err != nil {
			utils.Errorln("[ERROR]", err)
			exitFunc(1)
			return
		}
	} else {
		if err := runSetupInteractive(cfg); err != nil {
			utils.Errorln("[ERROR]", err)
			exitFunc(1)
			return
		}
	}

	if err := config.SaveConfig(cfg); err != nil {
		utils.Errorln("[ERROR] failed to save configuration:", err)
		exitFunc(1)
		return
	}

	maybeEnsureFolder(cfg, logger)

	utils.Successln("[INFO] ✅ Configuration saved. Run `bwsf auth login` to authenticate.")
}

func nonInteractiveSetup() bool {
	return setupSkipHost || setupHostType != "" || setupURL != "" || setupEmail != "" ||
		len(setupSaveFiles) > 0 || setupFolder != ""
}

func runSetupNonInteractive(cfg *config.Config) error {
	if setupSkipHost {
		// leave hosts unchanged (may be empty)
	} else if setupHostType != "" || setupEmail != "" || setupURL != "" {
		if err := validateSetupNonInteractiveHostFlags(); err != nil {
			return err
		}
		prior := append([]config.Host(nil), cfg.Settings.Hosts...)
		host, err := core.MapPromptHostToV2(setupHostType, setupURL, setupEmail, setupFolder)
		if err != nil {
			return err
		}
		core.UpsertDefaultHost(cfg, host)
		core.PreserveSetupFields(cfg, &config.Config{Settings: config.GlobalSettings{Hosts: prior}})
	} else if setupFolder != "" && len(cfg.Settings.Hosts) > 0 {
		if err := config.ValidateFolderName(setupFolder); err != nil {
			return err
		}
		if h := cfg.DefaultHost(); h != nil {
			h.TargetSection = strings.TrimSpace(setupFolder)
		}
	}

	if len(setupSaveFiles) > 0 {
		cfg.Settings.SaveFiles = normalizeSaveFiles(setupSaveFiles)
	}
	return cfg.Validate()
}

func validateSetupNonInteractiveHostFlags() error {
	if setupHostType == "" || setupEmail == "" {
		return fmt.Errorf("non-interactive host setup requires --host-type and --email (or pass --skip-host)")
	}
	if setupHostType != "cloud" && setupHostType != "selfhosted" {
		return fmt.Errorf("--host-type must be cloud or selfhosted")
	}
	if setupHostType == "selfhosted" && strings.TrimSpace(setupURL) == "" {
		return fmt.Errorf("--url is required when --host-type=selfhosted")
	}
	return nil
}

// validateSetupNonInteractiveFlags retained for older tests.
func validateSetupNonInteractiveFlags() error {
	if setupSkipHost {
		return nil
	}
	if !nonInteractiveSetup() {
		return nil
	}
	if setupHostType == "" && setupEmail == "" && setupURL == "" && setupFolder == "" && len(setupSaveFiles) == 0 {
		return nil
	}
	if setupHostType != "" || setupEmail != "" || setupURL != "" {
		return validateSetupNonInteractiveHostFlags()
	}
	return nil
}

func runSetupInteractive(cfg *config.Config) error {
	if len(cfg.Settings.Hosts) == 0 {
		action, err := selectSetupHostAction(false)
		if err != nil {
			return err
		}
		switch action {
		case "add":
			if err := interactiveAddHost(cfg, true); err != nil {
				return err
			}
		case "skip":
			// hosts remain empty
		default:
			return fmt.Errorf("unknown host action %q", action)
		}
	} else {
		action, err := selectSetupHostAction(true)
		if err != nil {
			return err
		}
		switch action {
		case "add":
			if err := interactiveAddHost(cfg, false); err != nil {
				return err
			}
		case "update":
			if err := interactiveUpdateHost(cfg); err != nil {
				return err
			}
		case "default":
			if err := interactiveChangeDefault(cfg); err != nil {
				return err
			}
		case "skip":
			// no host changes
		default:
			return fmt.Errorf("unknown host action %q", action)
		}
	}

	sfAction, err := selectSaveFilesAction()
	if err != nil {
		return err
	}
	switch sfAction {
	case "set":
		globs, err := inputSaveFilesGlobs()
		if err != nil {
			return err
		}
		cfg.Settings.SaveFiles = normalizeSaveFiles(globs)
	case "unset":
		cfg.Settings.SaveFiles = nil
	}
	return nil
}

func interactiveAddHost(cfg *config.Config, first bool) error {
	hostType, err := utils.SelectHostType()
	if err != nil {
		return err
	}
	url := ""
	if core.NormalizePromptHostType(hostType) == config.HostTypeSelfhost {
		url, err = utils.InputURL()
		if err != nil {
			return err
		}
	}
	email, err := utils.InputEmail()
	if err != nil {
		return err
	}
	folder := setupFolder
	if folder == "" {
		folder = config.DefaultFolderName
	}

	host, err := core.MapPromptHostToV2(hostType, url, email, folder)
	if err != nil {
		return err
	}
	if first {
		host.ID = config.DefaultHostID
		host.IsDefault = true
	} else {
		id, idErr := inputNewHostID()
		if idErr != nil {
			return idErr
		}
		host.ID = strings.TrimSpace(id)
		makeDef, cErr := confirmMakeDefault()
		if cErr != nil {
			return cErr
		}
		host.IsDefault = makeDef
		if !makeDef && cfg.DefaultHost() == nil {
			host.IsDefault = true
		}
	}
	prior := append([]config.Host(nil), cfg.Settings.Hosts...)
	if host.IsDefault {
		for i := range cfg.Settings.Hosts {
			cfg.Settings.Hosts[i].IsDefault = false
		}
	}
	cfg.Settings.Hosts = append(cfg.Settings.Hosts, host)
	core.PreserveSetupFields(cfg, &config.Config{Settings: config.GlobalSettings{Hosts: prior}})
	return nil
}

func interactiveUpdateHost(cfg *config.Config) error {
	id, err := selectSetupHostID(cfg)
	if err != nil {
		return err
	}
	h := cfg.FindHost(id)
	if h == nil {
		return fmt.Errorf("host %q not found", id)
	}
	hostType, err := utils.SelectHostType()
	if err != nil {
		return err
	}
	url := ""
	if core.NormalizePromptHostType(hostType) == config.HostTypeSelfhost {
		url, err = utils.InputURL()
		if err != nil {
			return err
		}
	}
	email, err := utils.InputEmail()
	if err != nil {
		return err
	}
	folder := h.TargetSection
	if setupFolder != "" {
		folder = setupFolder
	} else if text, terr := utils.InputText(fmt.Sprintf("Target section [%s]: ", h.TargetSection)); terr == nil {
		folder = strings.TrimSpace(text)
	}

	updated, err := core.MapPromptHostToV2(hostType, url, email, folder)
	if err != nil {
		return err
	}
	updated.ID = h.ID
	updated.IsDefault = h.IsDefault
	updated.DeviceIdentifier = h.DeviceIdentifier
	*h = updated
	return nil
}

func interactiveChangeDefault(cfg *config.Config) error {
	id, err := selectSetupHostID(cfg)
	if err != nil {
		return err
	}
	found := false
	for i := range cfg.Settings.Hosts {
		cfg.Settings.Hosts[i].IsDefault = cfg.Settings.Hosts[i].ID == id
		if cfg.Settings.Hosts[i].IsDefault {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("host %q not found", id)
	}
	return nil
}

func maybeEnsureFolder(cfg *config.Config, logger core.Logger) {
	if cfg.DefaultHost() == nil {
		utils.Infoln("[INFO] No host configured; skip folder ensure. Run `bwsf auth login` after adding a host.")
		return
	}
	folderName := config.ResolveFolderName(cfg)
	bw := newBwClientFromConfig(cfg)
	defer clearAPISession(bw)

	confirmCreateFolder := func() (bool, error) {
		return utils.ConfirmYesNo(fmt.Sprintf("%s folder not found. Create it? (y/N): ", folderName))
	}
	if migrateYes {
		confirmCreateFolder = func() (bool, error) { return true, nil }
	}

	if err := core.EnsureConfiguredFolderCore(bw, cfg, logger, utils.InputPassword, confirmCreateFolder); err != nil {
		if core.IsNotAuthenticatedError(err) {
			utils.Infoln("[INFO] Folder check skipped until `bwsf auth login` succeeds.")
		} else {
			utils.Warningln("[WARN] Folder ensure:", err)
		}
	}
}

func normalizeSaveFiles(in []string) []string {
	var out []string
	for _, s := range in {
		for _, part := range strings.Split(s, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

func defaultSelectSetupHostAction(hasHosts bool) (string, error) {
	items := []string{"Add host", "Skip"}
	if hasHosts {
		items = []string{"Add host", "Update existing host", "Change default host", "Skip"}
	}
	prompt := promptui.Select{Label: "Host configuration", Items: items}
	idx, _, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to select host action: %w", err)
	}
	if !hasHosts {
		if idx == 0 {
			return "add", nil
		}
		return "skip", nil
	}
	switch idx {
	case 0:
		return "add", nil
	case 1:
		return "update", nil
	case 2:
		return "default", nil
	default:
		return "skip", nil
	}
}

func defaultSelectSetupHostID(cfg *config.Config) (string, error) {
	ids := make([]string, 0, len(cfg.Settings.Hosts))
	for _, h := range cfg.Settings.Hosts {
		label := h.ID
		if h.IsDefault {
			label += " [default]"
		}
		ids = append(ids, label)
	}
	prompt := promptui.Select{Label: "Select host", Items: ids}
	idx, _, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to select host: %w", err)
	}
	return cfg.Settings.Hosts[idx].ID, nil
}

func defaultSelectSaveFilesAction() (string, error) {
	prompt := promptui.Select{
		Label: "save_files",
		Items: []string{"Set save_files", "Leave unset / clear"},
	}
	idx, _, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to select save_files action: %w", err)
	}
	if idx == 0 {
		return "set", nil
	}
	return "unset", nil
}

func defaultInputSaveFilesGlobs() ([]string, error) {
	text, err := utils.InputText("Enter save_files globs (comma-separated, !prefix to exclude): ")
	if err != nil {
		return nil, err
	}
	return normalizeSaveFiles([]string{text}), nil
}

func defaultConfirmMakeDefault() (bool, error) {
	return utils.ConfirmYesNo("Make this host the default? (y/N): ")
}

func defaultInputNewHostID() (string, error) {
	return utils.InputText("Enter host id: ")
}
