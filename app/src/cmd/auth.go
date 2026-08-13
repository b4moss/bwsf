package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"bwsf/src/config"
	"bwsf/src/infra"
	"bwsf/src/utils"

	"github.com/spf13/cobra"
)

var (
	authClear bool
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate the API backend with a Personal API Key",
	Long: `Store a Bitwarden Personal API Key in the OS secret store (macOS Keychain /
Linux secret service) and obtain an Identity access token.

Only used when backend is "api" (see bwsf backend --set api).
Vault unlock and CRUD are not handled here (later Issue #53 steps).`,
	Run: runAuth,
}

func init() {
	authCmd.Flags().BoolVar(&authClear, "clear", false, "Remove stored Personal API Key from the OS secret store")
	rootCmd.AddCommand(authCmd)
}

func runAuth(cmd *cobra.Command, args []string) {
	cfg := loadConfigOrEmpty()

	store := infra.NewKeyringStore()

	if authClear {
		if err := infra.ClearAPICredentials(store); err != nil {
			utils.Errorln("[ERROR]", err)
			os.Exit(1)
		}
		utils.Successln("[INFO] ✅ Cleared stored Personal API Key from the OS secret store")
		return
	}

	if cfg.GetBackend() != config.BackendAPI {
		utils.Warningln("[WARN] Current backend is " + cfg.GetBackend() + ". API auth is intended for backend=api.")
		utils.Infoln("[INFO] Switch with: bwsf backend --set api")
	}

	if err := ensureHostConfigForAPI(cfg); err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}

	configPath, err := config.GetConfigPath()
	if err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}
	identityBase, err := infra.ResolveIdentityBase(cfg.HostType, cfg.SelfhostedURL)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}
	utils.Infoln("[INFO] Using config: " + configPath)
	utils.Infoln("[INFO] Identity URL: " + identityBase)

	creds, err := promptAPICredentials(store)
	if err != nil {
		utils.Errorln("[ERROR]", err)
		os.Exit(1)
	}

	client := infra.NewApiBwClientWithDeps(cfg, store, infra.NewIdentityClient(), nil)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := client.AuthenticateWithCredentials(ctx, creds, true); err != nil {
		utils.Errorln("[ERROR] Authentication failed: " + err.Error())
		os.Exit(1)
	}

	utils.Successln("[INFO] ✅ Authenticated with Personal API Key (token stored in memory for this process)")
	utils.Infoln("[INFO] Personal API Key is stored in the OS secret store (Keychain / secret service)")
	utils.Infoln("[INFO] Vault CRUD via API is not implemented yet (Issue #53 Step 4)")
}

func ensureHostConfigForAPI(cfg *config.Config) error {
	if cfg.HostType == "cloud" {
		return nil
	}
	if cfg.HostType == "selfhosted" && strings.TrimSpace(cfg.SelfhostedURL) != "" {
		return nil
	}

	// Missing host settings: prompt and save (does not require bw CLI).
	utils.Infoln("[INFO] Host configuration is required to resolve the Identity URL")
	hostType, err := utils.SelectHostType()
	if err != nil {
		return fmt.Errorf("failed to select host type: %w", err)
	}
	cfg.HostType = hostType
	if hostType == "selfhosted" {
		url, err := utils.InputURL()
		if err != nil {
			return fmt.Errorf("failed to get URL: %w", err)
		}
		cfg.SelfhostedURL = url
	} else {
		cfg.SelfhostedURL = ""
	}
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

func promptAPICredentials(store infra.SecretStore) (infra.APICredentials, error) {
	existing, err := infra.LoadAPICredentials(store)
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
