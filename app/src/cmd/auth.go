package cmd

import (
	"context"
	"errors"
	"time"

	"bwsf/src/config"
	"bwsf/src/infra"
	"bwsf/src/utils"

	"github.com/spf13/cobra"
)

var (
	authLoginHost  string
	authLogoutHost string
	authLogoutAll  bool
)

// Interactive stubs for auth login (overridable in tests).
var (
	confirmAPIKeyReuse   = utils.ConfirmYesNo
	inputAPIClientID     = func() (string, error) { return utils.InputText("Enter Personal API Key client_id: ") }
	inputAPIClientSecret = func() (string, error) {
		return utils.InputSecret("Enter Personal API Key client_secret: ")
	}
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Personal API Key authentication",
	Long: `Authenticate with a Bitwarden Personal API Key (auth login / auth logout).

auth login / logout manage API Keys in the OS secret store.
logout also clears the vault_unlock session blob for the host.
unlock / lock manage vault sessions only (they do not change API Keys).

Running auth with no subcommand prints this help.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store API Key, verify Identity, and unlock the vault",
	Long: `Store a Personal API Key for the resolved host, verify it with Identity,
then unlock the vault and persist vault_unlock (same as bwsf unlock).

Host resolution: --host → project .bwsf host → global is_default.
Does not create hosts — run bwsf setup first.`,
	Run: runAuthLogin,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove API Key and vault session for a host",
	Long: `Remove Personal API Key and vault_unlock for the resolved host.

Unlike bwsf lock (vault session only), logout clears API credentials too.
Host resolution: --host → project .bwsf host → global is_default.
Use --all to clear every registered host (no-op if hosts is empty).`,
	Run: runAuthLogout,
}

func init() {
	authLoginCmd.Flags().StringVar(&authLoginHost, "host", "", "Host id from global config hosts[]")
	authLogoutCmd.Flags().StringVar(&authLogoutHost, "host", "", "Host id from global config hosts[]")
	authLogoutCmd.Flags().BoolVar(&authLogoutAll, "all", false, "Logout all registered hosts")
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}

func runAuthLogin(cmd *cobra.Command, args []string) {
	cfg := loadConfigOrEmpty()
	projectHost := loadProjectHostID()
	host := resolveHostForCommand(cfg, authLoginHost, projectHost)
	if host == nil {
		return
	}

	store := newSecretStore()

	configPath, err := config.GetConfigPath()
	if err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
		return
	}
	identityBase, err := infra.ResolveIdentityBase(host.Type, host.HostURL)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
		return
	}
	utils.Infoln("[INFO] Using config: " + configPath)
	utils.Infoln("[INFO] Host: " + host.ID)
	utils.Infoln("[INFO] Identity URL: " + identityBase)

	creds, err := promptAPICredentials(store, host.ID)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
		return
	}

	client := newAuthClient(cfg, host, store)
	defer client.ClearSession()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := client.AuthenticateWithCredentials(ctx, creds, true); err != nil {
		utils.Errorln("[ERROR] Authentication failed: " + err.Error())
		exitFunc(1)
		return
	}

	utils.Successln("[INFO] ✅ Authenticated with Personal API Key")
	utils.Infoln("[INFO] Personal API Key is stored in the OS secret store (Keychain / secret service)")

	if err := unlockVaultForHost(client); err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
		return
	}

	utils.Successln("[INFO] ✅ Vault unlocked for host " + host.ID)
	utils.Infoln("[INFO] Session persisted in OS secret store until bwsf lock or auth logout")
}

func runAuthLogout(cmd *cobra.Command, args []string) {
	if authLogoutAll && authLogoutHost != "" {
		utils.Errorln("[ERROR] --all and --host cannot be combined")
		exitFunc(1)
		return
	}

	cfg := loadConfigOrEmpty()
	store := newSecretStore()

	if authLogoutAll {
		if len(cfg.Settings.Hosts) == 0 {
			utils.Successln("[INFO] ✅ No hosts configured; nothing to logout")
			return
		}
		for i := range cfg.Settings.Hosts {
			id := cfg.Settings.Hosts[i].ID
			if err := logoutHost(store, id); err != nil {
				utils.Errorln("[ERROR]", err)
				exitFunc(1)
				return
			}
		}
		utils.Successln("[INFO] ✅ Logged out all hosts")
		return
	}

	projectHost := loadProjectHostID()
	host := resolveHostForCommand(cfg, authLogoutHost, projectHost)
	if host == nil {
		return
	}
	if err := logoutHost(store, host.ID); err != nil {
		utils.Errorln("[ERROR]", err)
		exitFunc(1)
		return
	}
	utils.Successln("[INFO] ✅ Logged out host " + host.ID)
}

func logoutHost(store infra.SecretStore, hostID string) error {
	if err := infra.ClearAPICredentials(store, hostID); err != nil {
		return err
	}
	return infra.LockVaultSessionForHost(store, hostID)
}

func promptAPICredentials(store infra.SecretStore, hostID string) (infra.APICredentials, error) {
	existing, err := infra.LoadAPICredentials(store, hostID)
	if err == nil && existing.ClientID != "" {
		utils.Infoln("[INFO] Found Personal API Key in OS secret store (client_id=" + maskSecret(existing.ClientID) + ")")
		reuse, confirmErr := confirmAPIKeyReuse("Use stored Personal API Key? (y/N): ")
		if confirmErr != nil {
			return infra.APICredentials{}, confirmErr
		}
		if reuse {
			return existing, nil
		}
	} else if err != nil && !errors.Is(err, infra.ErrSecretNotFound) {
		return infra.APICredentials{}, err
	}

	utils.Infoln("[INFO] Create a Personal API Key in the Bitwarden web vault: Account Settings → Security → Keys")
	clientID, err := inputAPIClientID()
	if err != nil {
		return infra.APICredentials{}, err
	}
	clientSecret, err := inputAPIClientSecret()
	if err != nil {
		return infra.APICredentials{}, err
	}
	return infra.APICredentials{ClientID: clientID, ClientSecret: clientSecret}, nil
}

func maskSecret(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "…" + s[len(s)-2:]
}
