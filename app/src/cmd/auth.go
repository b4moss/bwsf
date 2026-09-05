package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"bwsf/src/config"
	"bwsf/src/core"
	"bwsf/src/infra"
	"bwsf/src/utils"

	"github.com/spf13/cobra"
)

var (
	authClear bool
	authHost  string
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with a Personal API Key",
	Long: `Store a Bitwarden Personal API Key in the OS secret store (macOS Keychain /
Linux secret service) and obtain an Identity access token.

After auth, push/pull/list prompt for your master password to unlock the vault.`,
	Run: runAuth,
}

func init() {
	authCmd.Flags().BoolVar(&authClear, "clear", false, "Remove stored Personal API Key from the OS secret store")
	authCmd.Flags().StringVar(&authHost, "host", "", "Host id from global config hosts[]")
	rootCmd.AddCommand(authCmd)
}

func runAuth(cmd *cobra.Command, args []string) {
	cfg := loadConfigOrEmpty()

	store := infra.NewKeyringStore()

	if authClear {
		host, err := ensureHostConfigForAPI(cfg, authHost)
		if err != nil {
			utils.Errorln("[ERROR]", err)
			os.Exit(1)
		}
		if err := infra.ClearAPICredentials(store, host.ID); err != nil {
			utils.Errorln("[ERROR]", err)
			os.Exit(1)
		}
		utils.Successln("[INFO] ✅ Cleared stored Personal API Key from the OS secret store")
		return
	}

	host, err := ensureHostConfigForAPI(cfg, authHost)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}

	configPath, err := config.GetConfigPath()
	if err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}
	identityBase, err := infra.ResolveIdentityBase(host.Type, host.HostURL)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}
	utils.Infoln("[INFO] Using config: " + configPath)
	utils.Infoln("[INFO] Host: " + host.ID)
	utils.Infoln("[INFO] Identity URL: " + identityBase)

	creds, err := promptAPICredentials(store, host.ID)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}

	client := infra.NewApiBwClientWithDepsForHost(cfg, host, store, infra.NewIdentityClient(), nil)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := client.AuthenticateWithCredentials(ctx, creds, true); err != nil {
		utils.Errorln("[ERROR] Authentication failed: " + err.Error())
		os.Exit(1)
	}

	utils.Successln("[INFO] ✅ Authenticated with Personal API Key (token stored in memory for this process)")
	utils.Infoln("[INFO] Personal API Key is stored in the OS secret store (Keychain / secret service)")
	utils.Infoln("[INFO] Next: run push/pull/list; you will be prompted for your master password to unlock the vault")
}

// ensureHostConfigForAPI resolves or interactively creates a host for API auth.
func ensureHostConfigForAPI(cfg *config.Config, cliHost string) (*config.Host, error) {
	if h, err := config.ResolveHost(cfg, cliHost, ""); err == nil {
		return h, nil
	} else if cliHost != "" {
		return nil, err
	}

	// Missing host settings: prompt and save.
	utils.Infoln("[INFO] Host configuration is required to resolve the Identity URL")
	hostType, err := utils.SelectHostType()
	if err != nil {
		return nil, fmt.Errorf("failed to select host type: %w", err)
	}
	url := ""
	if core.NormalizePromptHostType(hostType) == config.HostTypeSelfhost {
		url, err = utils.InputURL()
		if err != nil {
			return nil, fmt.Errorf("failed to get URL: %w", err)
		}
	}
	email := ""
	if e, eerr := utils.InputEmail(); eerr == nil {
		email = e
	}

	host, err := core.MapPromptHostToV2(hostType, url, email, config.DefaultFolderName)
	if err != nil {
		return nil, err
	}
	core.UpsertDefaultHost(cfg, host)
	if err := config.SaveConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}
	return cfg.FindHost(host.ID), nil
}

func promptAPICredentials(store infra.SecretStore, hostID string) (infra.APICredentials, error) {
	existing, err := infra.LoadAPICredentials(store, hostID)
	if err == nil && existing.ClientID != "" {
		utils.Infoln("[INFO] Found Personal API Key in OS secret store (client_id=" + maskSecret(existing.ClientID) + ")")
		reuse, confirmErr := utils.ConfirmYesNo("Use stored Personal API Key? (y/N): ")
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
	clientID, err := utils.InputText("Enter Personal API Key client_id: ")
	if err != nil {
		return infra.APICredentials{}, err
	}
	clientSecret, err := utils.InputSecret("Enter Personal API Key client_secret: ")
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