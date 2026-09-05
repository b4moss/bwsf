package cmd

import (
	"errors"
	"os"

	"bwsf/src/config"
	"bwsf/src/infra"
	"bwsf/src/utils"

	"github.com/spf13/cobra"
)

var unlockHost string

var unlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Unlock the vault session for a host",
	Long: `Unlock the Bitwarden vault for the resolved host and persist an opaque
session blob in the OS secret store (vault_unlock).

This does not store or change Personal API Keys — use bwsf auth login / logout for that.
Host resolution: --host → project .bwsf host → global is_default.`,
	Run: runUnlock,
}

func init() {
	unlockCmd.Flags().StringVar(&unlockHost, "host", "", "Host id from global config hosts[]")
	rootCmd.AddCommand(unlockCmd)
}

func runUnlock(cmd *cobra.Command, args []string) {
	cfg := loadConfigOrEmpty()
	projectHost := loadProjectHostID()
	host := resolveHostForCommand(cfg, unlockHost, projectHost)
	if host == nil {
		return
	}

	store := newSecretStore()
	_, err := infra.LoadAPICredentials(store, host.ID)
	if err != nil {
		if errors.Is(err, infra.ErrSecretNotFound) {
			utils.Errorln("[ERROR] No Personal API Key stored for host", host.ID)
			utils.Infoln("[INFO] Run: bwsf auth login")
			exitFunc(1)
			return
		}
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
		return
	}

	client := newUnlockClient(cfg, host, store)
	defer client.ClearSession()

	if err := unlockVaultForHost(client); err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
		return
	}

	utils.Successln("[INFO] ✅ Vault unlocked for host " + host.ID)
	utils.Infoln("[INFO] Session persisted in OS secret store until bwsf lock (or auth logout)")
}

// unlockVaultForHost prompts for the master password and unlocks via client.
// Caller owns ClearSession (defer). Does not print success messages.
func unlockVaultForHost(client unlockClient) error {
	password, err := inputPassword()
	if err != nil {
		return err
	}
	if password == "" {
		return errors.New("master password cannot be empty")
	}
	return client.Unlock(password)
}

// loadProjectHostID returns project config host id when present (non-interactive).
func loadProjectHostID() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	pc, _, err := config.ResolveProjectConfig(wd, func(paths []string) (string, error) {
		if len(paths) == 0 {
			return "", nil
		}
		return paths[0], nil
	})
	if err != nil || pc == nil {
		return ""
	}
	return pc.EffectiveHost()
}
