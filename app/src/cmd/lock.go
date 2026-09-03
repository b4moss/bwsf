package cmd

import (
	"bwsf/src/infra"
	"bwsf/src/utils"

	"github.com/spf13/cobra"
)

var (
	lockHost string
	lockAll  bool
)

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Lock the vault session for a host",
	Long: `Remove the opaque vault_unlock session blob from the OS secret store.

Personal API Keys are kept — use bwsf auth --clear (or future auth logout) to remove them.
Host resolution: --host → project .bwsf host → global is_default.
Use --all to clear vault_unlock for every registered host (no-op if hosts is empty).`,
	Run: runLock,
}

func init() {
	lockCmd.Flags().StringVar(&lockHost, "host", "", "Host id from global config hosts[]")
	lockCmd.Flags().BoolVar(&lockAll, "all", false, "Lock vault sessions for all registered hosts")
	rootCmd.AddCommand(lockCmd)
}

func runLock(cmd *cobra.Command, args []string) {
	if lockAll && lockHost != "" {
		utils.Errorln("[ERROR] --all and --host cannot be combined")
		exitFunc(1)
		return
	}

	cfg := loadConfigOrEmpty()
	store := infra.NewKeyringStore()

	if lockAll {
		if len(cfg.Settings.Hosts) == 0 {
			utils.Successln("[INFO] ✅ No hosts configured; nothing to lock")
			return
		}
		for i := range cfg.Settings.Hosts {
			id := cfg.Settings.Hosts[i].ID
			if err := infra.LockVaultSessionForHost(store, id); err != nil {
				utils.Errorln("[ERROR]", err)
				exitFunc(1)
				return
			}
		}
		utils.Successln("[INFO] ✅ Locked vault sessions for all hosts")
		return
	}

	projectHost := loadProjectHostID()
	host := resolveHostForCommand(cfg, lockHost, projectHost)
	if host == nil {
		return
	}
	if err := infra.LockVaultSessionForHost(store, host.ID); err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
		return
	}
	utils.Successln("[INFO] ✅ Locked vault session for host " + host.ID)
}
