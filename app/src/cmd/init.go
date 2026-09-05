package cmd

import (
	"fmt"
	"os"
	"strings"

	"bwsf/src/config"
	"bwsf/src/utils"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	initHost                string
	initSkipHost            bool
	initSaveFiles           []string
	initOverrideProjectName string
	initOverrideFlagSet     bool
)

// Interactive stubs for init (overridable in tests).
var (
	selectInitHostID          = defaultSelectInitHostID
	selectOverrideNameAction  = defaultSelectOverrideNameAction
	inputOverrideProjectName  = defaultInputOverrideProjectName
	confirmInitOverwrite      = defaultConfirmInitOverwrite
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create project .bwsf/config.jsonc",
	Long: `Create a project configuration file at ./.bwsf/config.jsonc.

Requires an existing global config (run bwsf setup first; hosts: [] is OK).
Optionally selects a host id from global hosts[], save_files, and override_project_name.
Does not touch Keychain, unlock/lock, or auth.`,
	Run: runInit,
}

func init() {
	initCmd.Flags().StringVar(&initHost, "host", "", "Host id to write into project config (non-interactive)")
	initCmd.Flags().BoolVar(&initSkipHost, "skip-host", false, "Omit project host key (non-interactive)")
	initCmd.Flags().StringSliceVar(&initSaveFiles, "save-files", nil, "Project save_files globs (comma or repeat; ! prefix = exclude)")
	initCmd.Flags().StringVar(&initOverrideProjectName, "override-project-name", "", "override_project_name value (empty omits the key)")
	initCmd.PreRun = func(cmd *cobra.Command, args []string) {
		initOverrideFlagSet = cmd.Flags().Changed("override-project-name")
	}
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfigWithMigrate(migrateHooks())
	if err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
		return
	}
	if cfg == nil {
		utils.Errorln("[ERROR] no global config found; run `bwsf setup` first")
		exitFunc(1)
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
		return
	}

	if err := ensureInitOverwriteOK(cwd); err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
		return
	}

	pc := &config.ProjectConfig{}
	if nonInteractiveInit() {
		if err := applyInitNonInteractive(cfg, pc); err != nil {
			utils.Errorln("[ERROR]", err)
			exitFunc(1)
			return
		}
	} else {
		if err := applyInitInteractive(cfg, pc); err != nil {
			utils.Errorln("[ERROR]", err)
			exitFunc(1)
			return
		}
	}

	if err := config.SaveProjectConfig(cwd, pc); err != nil {
		utils.Errorln("[ERROR] failed to write project config:", err)
		exitFunc(1)
		return
	}

	utils.Successln("[INFO] ✅ Wrote project config:", config.GetProjectConfigWritePath(cwd))
}

func nonInteractiveInit() bool {
	return initSkipHost || initHost != "" || len(initSaveFiles) > 0 || initOverrideFlagSet
}

func ensureInitOverwriteOK(cwd string) error {
	if !config.LocalProjectConfigExists(cwd) {
		return nil
	}
	if migrateYes {
		return nil
	}
	ok, err := confirmInitOverwrite(config.GetProjectConfigWritePath(cwd))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("project config already exists; overwrite declined")
	}
	return nil
}

func applyInitNonInteractive(cfg *config.Config, pc *config.ProjectConfig) error {
	if initSkipHost && initHost != "" {
		return fmt.Errorf("cannot combine --skip-host and --host")
	}
	if initHost != "" {
		if cfg.FindHost(initHost) == nil {
			return fmt.Errorf("host %q not found in global config", initHost)
		}
		pc.Host = initHost
	}

	if len(initSaveFiles) > 0 {
		pc.SaveFiles = normalizeSaveFiles(initSaveFiles)
	}
	if initOverrideFlagSet {
		name := strings.TrimSpace(initOverrideProjectName)
		if name != "" {
			pc.OverrideProjectName = name
		}
	}
	return nil
}

func applyInitInteractive(cfg *config.Config, pc *config.ProjectConfig) error {
	if len(cfg.Settings.Hosts) > 0 {
		id, err := selectInitHostID(cfg)
		if err != nil {
			return err
		}
		id = strings.TrimSpace(id)
		if id != "" {
			pc.Host = id
		}
	}

	sfAction, err := selectSaveFilesAction()
	if err != nil {
		return err
	}
	if sfAction == "set" {
		globs, err := inputSaveFilesGlobs()
		if err != nil {
			return err
		}
		if sf := normalizeSaveFiles(globs); len(sf) > 0 {
			pc.SaveFiles = sf
		}
		// empty globs → leave unset (F3)
	}

	ovAction, err := selectOverrideNameAction()
	if err != nil {
		return err
	}
	if ovAction == "set" {
		name, err := inputOverrideProjectName()
		if err != nil {
			return err
		}
		name = strings.TrimSpace(name)
		if name != "" {
			pc.OverrideProjectName = name
		}
	}
	return nil
}

func defaultSelectInitHostID(cfg *config.Config) (string, error) {
	items := make([]string, 0, len(cfg.Settings.Hosts)+1)
	items = append(items, "Skip (no host key)")
	for _, h := range cfg.Settings.Hosts {
		label := h.ID
		if h.IsDefault {
			label += " [default]"
		}
		items = append(items, label)
	}
	prompt := promptui.Select{Label: "Project host", Items: items}
	idx, _, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to select project host: %w", err)
	}
	if idx == 0 {
		return "", nil
	}
	return cfg.Settings.Hosts[idx-1].ID, nil
}

func defaultSelectOverrideNameAction() (string, error) {
	prompt := promptui.Select{
		Label: "override_project_name",
		Items: []string{"Set override_project_name", "Leave unset"},
	}
	idx, _, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to select override action: %w", err)
	}
	if idx == 0 {
		return "set", nil
	}
	return "unset", nil
}

func defaultInputOverrideProjectName() (string, error) {
	return utils.InputText("Enter override_project_name: ")
}

func defaultConfirmInitOverwrite(path string) (bool, error) {
	return utils.ConfirmOverwrite(fmt.Sprintf("Project config %s already exists. Overwrite? (y/N): ", path))
}
