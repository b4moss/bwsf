package cmd

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bwsf/src/config"
	"bwsf/src/infra"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthLoginLogoutCmd_Registered(t *testing.T) {
	assert.NotNil(t, authLoginCmd.Flags().Lookup("host"))
	assert.NotNil(t, authLogoutCmd.Flags().Lookup("host"))
	assert.NotNil(t, authLogoutCmd.Flags().Lookup("all"))
	assert.Nil(t, authCmd.Flags().Lookup("clear"))
	assert.Contains(t, authCmd.Long, "API Key")
	assert.Contains(t, authLoginCmd.Long, "unlock")
	assert.Contains(t, authLogoutCmd.Long, "vault_unlock")
}

func TestRunAuth_BareHelpOnly(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	store := infra.NewMemorySecretStore()
	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }

	authCmd.Run(authCmd, nil)
	assert.False(t, rec.called)
	assert.False(t, store.Has("hosts/default/api_client_id"))
}

func TestRunAuthLogout_AllAndHostConflict(t *testing.T) {
	rec := stubCmdDeps(t, nil, nil)
	authLogoutAll = true
	authLogoutHost = "work"
	t.Cleanup(func() { authLogoutAll = false; authLogoutHost = "" })
	runAuthLogout(authLogoutCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

func TestRunAuthLogout_AllEmptyHosts(t *testing.T) {
	withTempHome(t)
	require.NoError(t, config.SaveConfig(config.NewEmptyConfig()))
	rec := stubCmdDeps(t, nil, nil)
	authLogoutAll = true
	t.Cleanup(func() { authLogoutAll = false })
	runAuthLogout(authLogoutCmd, nil)
	assert.False(t, rec.called)
}

func TestRunAuthLogout_UnknownHost(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	rec := stubCmdDeps(t, nil, nil)
	authLogoutHost = "missing"
	t.Cleanup(func() { authLogoutHost = "" })
	runAuthLogout(authLogoutCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

func TestRunAuthLogout_ClearsAPIKeyAndVaultUnlock(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	store := infra.NewMemorySecretStore()
	require.NoError(t, infra.SaveAPICredentials(store, config.DefaultHostID, infra.APICredentials{ClientID: "id", ClientSecret: "sec"}))
	require.NoError(t, infra.SaveVaultUnlock(store, config.DefaultHostID, "blob"))

	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	runAuthLogout(authLogoutCmd, nil)
	assert.False(t, rec.called)

	_, err := infra.LoadAPICredentials(store, config.DefaultHostID)
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
	_, err = infra.LoadVaultUnlock(store, config.DefaultHostID)
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
}

func TestRunAuthLogout_IdempotentMissing(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	store := infra.NewMemorySecretStore()
	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	runAuthLogout(authLogoutCmd, nil)
	assert.False(t, rec.called)
}

func TestRunAuthLogout_HostLeavesOther(t *testing.T) {
	withTempHome(t)
	cfg := config.NewEmptyConfig()
	cfg.Settings.Hosts = []config.Host{
		{ID: config.DefaultHostID, Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, TargetSection: "dotenvs", IsDefault: true, Email: "a@example.com"},
		{ID: "work", Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, TargetSection: "dotenvs", Email: "a@example.com"},
	}
	require.NoError(t, config.SaveConfig(cfg))

	store := infra.NewMemorySecretStore()
	require.NoError(t, infra.SaveAPICredentials(store, config.DefaultHostID, infra.APICredentials{ClientID: "d", ClientSecret: "ds"}))
	require.NoError(t, infra.SaveAPICredentials(store, "work", infra.APICredentials{ClientID: "w", ClientSecret: "ws"}))
	require.NoError(t, infra.SaveVaultUnlock(store, config.DefaultHostID, "d-blob"))
	require.NoError(t, infra.SaveVaultUnlock(store, "work", "w-blob"))

	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	authLogoutHost = "work"
	t.Cleanup(func() { authLogoutHost = "" })
	runAuthLogout(authLogoutCmd, nil)
	assert.False(t, rec.called)

	_, err := infra.LoadAPICredentials(store, "work")
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
	_, err = infra.LoadVaultUnlock(store, "work")
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
	_, err = infra.LoadAPICredentials(store, config.DefaultHostID)
	require.NoError(t, err)
	blob, err := infra.LoadVaultUnlock(store, config.DefaultHostID)
	require.NoError(t, err)
	assert.Equal(t, "d-blob", blob)
}

func TestRunAuthLogout_All(t *testing.T) {
	withTempHome(t)
	cfg := config.NewEmptyConfig()
	cfg.Settings.Hosts = []config.Host{
		{ID: config.DefaultHostID, Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, TargetSection: "dotenvs", IsDefault: true, Email: "a@example.com"},
		{ID: "work", Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, TargetSection: "dotenvs", Email: "a@example.com"},
	}
	require.NoError(t, config.SaveConfig(cfg))

	store := infra.NewMemorySecretStore()
	require.NoError(t, infra.SaveAPICredentials(store, config.DefaultHostID, infra.APICredentials{ClientID: "d", ClientSecret: "ds"}))
	require.NoError(t, infra.SaveAPICredentials(store, "work", infra.APICredentials{ClientID: "w", ClientSecret: "ws"}))
	require.NoError(t, infra.SaveVaultUnlock(store, config.DefaultHostID, "d-blob"))
	require.NoError(t, infra.SaveVaultUnlock(store, "work", "w-blob"))

	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	authLogoutAll = true
	t.Cleanup(func() { authLogoutAll = false })
	runAuthLogout(authLogoutCmd, nil)
	assert.False(t, rec.called)

	_, err := infra.LoadAPICredentials(store, config.DefaultHostID)
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
	_, err = infra.LoadAPICredentials(store, "work")
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
	_, err = infra.LoadVaultUnlock(store, config.DefaultHostID)
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
	_, err = infra.LoadVaultUnlock(store, "work")
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
}

func TestRunAuthLogin_UnknownHost(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	store := infra.NewMemorySecretStore()
	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	authLoginHost = "missing"
	t.Cleanup(func() { authLoginHost = "" })
	runAuthLogin(authLoginCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
	assert.False(t, store.Has("hosts/default/api_client_id"))
}

func TestRunAuthLogin_EmptyHosts(t *testing.T) {
	withTempHome(t)
	require.NoError(t, config.SaveConfig(config.NewEmptyConfig()))
	store := infra.NewMemorySecretStore()
	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	runAuthLogin(authLoginCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

func TestRunAuthLogin_SuccessPersistsKeyAndVaultUnlock(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	store := infra.NewMemorySecretStore()
	crypto := &infra.MockCryptoSession{ExportBlob: "v1:login-blob"}

	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	newAuthClient = func(cfg *config.Config, host *config.Host, s infra.SecretStore) authClient {
		return infra.NewApiBwClientWithDepsForHost(cfg, host, s, &infra.IdentityClient{
			HTTPClient: &http.Client{Transport: roundTripOK()},
		}, crypto)
	}

	runAuthLogin(authLoginCmd, nil)
	assert.False(t, rec.called)

	creds, err := infra.LoadAPICredentials(store, config.DefaultHostID)
	require.NoError(t, err)
	assert.Equal(t, "test.client.id", creds.ClientID)
	blob, err := infra.LoadVaultUnlock(store, config.DefaultHostID)
	require.NoError(t, err)
	assert.Equal(t, "v1:login-blob", blob)
	assert.False(t, crypto.Unlocked())
}

func TestRunAuthLogin_ReuseStoredKey(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	store := infra.NewMemorySecretStore()
	require.NoError(t, infra.SaveAPICredentials(store, config.DefaultHostID, infra.APICredentials{ClientID: "stored.id", ClientSecret: "stored-sec"}))
	crypto := &infra.MockCryptoSession{ExportBlob: "v1:reuse-blob"}

	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	confirmAPIKeyReuse = func(string) (bool, error) { return true, nil }
	inputAPIClientID = func() (string, error) {
		t.Fatal("should not prompt for new client_id when reusing")
		return "", nil
	}
	newAuthClient = func(cfg *config.Config, host *config.Host, s infra.SecretStore) authClient {
		return infra.NewApiBwClientWithDepsForHost(cfg, host, s, &infra.IdentityClient{
			HTTPClient: &http.Client{Transport: roundTripOK()},
		}, crypto)
	}

	runAuthLogin(authLoginCmd, nil)
	assert.False(t, rec.called)
	creds, err := infra.LoadAPICredentials(store, config.DefaultHostID)
	require.NoError(t, err)
	assert.Equal(t, "stored.id", creds.ClientID)
	blob, err := infra.LoadVaultUnlock(store, config.DefaultHostID)
	require.NoError(t, err)
	assert.Equal(t, "v1:reuse-blob", blob)
}

func TestRunAuthLogin_HostFlag(t *testing.T) {
	withTempHome(t)
	cfg := config.NewEmptyConfig()
	cfg.Settings.Hosts = []config.Host{
		{ID: config.DefaultHostID, Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, TargetSection: "dotenvs", IsDefault: true, Email: "a@example.com"},
		{ID: "work", Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, TargetSection: "dotenvs", Email: "a@example.com"},
	}
	require.NoError(t, config.SaveConfig(cfg))

	store := infra.NewMemorySecretStore()
	crypto := &infra.MockCryptoSession{ExportBlob: "v1:work-blob"}
	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	newAuthClient = func(cfg *config.Config, host *config.Host, s infra.SecretStore) authClient {
		return infra.NewApiBwClientWithDepsForHost(cfg, host, s, &infra.IdentityClient{
			HTTPClient: &http.Client{Transport: roundTripOK()},
		}, crypto)
	}
	authLoginHost = "work"
	t.Cleanup(func() { authLoginHost = "" })

	runAuthLogin(authLoginCmd, nil)
	assert.False(t, rec.called)
	_, err := infra.LoadAPICredentials(store, "work")
	require.NoError(t, err)
	blob, err := infra.LoadVaultUnlock(store, "work")
	require.NoError(t, err)
	assert.Equal(t, "v1:work-blob", blob)
	_, err = infra.LoadAPICredentials(store, config.DefaultHostID)
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
}

func TestRunAuthLogin_ProjectHost(t *testing.T) {
	withTempHome(t)
	cfg := config.NewEmptyConfig()
	cfg.Settings.Hosts = []config.Host{
		{ID: config.DefaultHostID, Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, TargetSection: "dotenvs", IsDefault: true, Email: "a@example.com"},
		{ID: "work", Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, TargetSection: "dotenvs", Email: "a@example.com"},
	}
	require.NoError(t, config.SaveConfig(cfg))

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".bwsf"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".bwsf", "config.json"), []byte(`{"host":"work"}`), 0o644))
	orig, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(orig) })

	store := infra.NewMemorySecretStore()
	crypto := &infra.MockCryptoSession{ExportBlob: "v1:proj-blob"}
	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	newAuthClient = func(cfg *config.Config, host *config.Host, s infra.SecretStore) authClient {
		assert.Equal(t, "work", host.ID)
		return infra.NewApiBwClientWithDepsForHost(cfg, host, s, &infra.IdentityClient{
			HTTPClient: &http.Client{Transport: roundTripOK()},
		}, crypto)
	}

	runAuthLogin(authLoginCmd, nil)
	assert.False(t, rec.called)
	blob, err := infra.LoadVaultUnlock(store, "work")
	require.NoError(t, err)
	assert.Equal(t, "v1:proj-blob", blob)
}

func TestRunAuthLogin_IdentityFailNoVaultUnlock(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	store := infra.NewMemorySecretStore()
	crypto := &infra.MockCryptoSession{ExportBlob: "should-not-write"}

	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	newAuthClient = func(cfg *config.Config, host *config.Host, s infra.SecretStore) authClient {
		return infra.NewApiBwClientWithDepsForHost(cfg, host, s, &infra.IdentityClient{
			HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(strings.NewReader(`{"error":"invalid"}`)),
					Header:     make(http.Header),
				}, nil
			})},
		}, crypto)
	}

	runAuthLogin(authLoginCmd, nil)
	assert.True(t, rec.called)
	_, err := infra.LoadVaultUnlock(store, config.DefaultHostID)
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
	// persist=true saves API Key before Identity; fixed behavior for E2.
	_, err = infra.LoadAPICredentials(store, config.DefaultHostID)
	require.NoError(t, err)
}

func TestRunAuthLogin_BadPasswordKeepsExistingVaultUnlock(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	store := infra.NewMemorySecretStore()
	require.NoError(t, infra.SaveVaultUnlock(store, config.DefaultHostID, "keep-me"))
	crypto := &infra.MockCryptoSession{UnlockErr: assertErr("bad password")}

	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	newAuthClient = func(cfg *config.Config, host *config.Host, s infra.SecretStore) authClient {
		return infra.NewApiBwClientWithDepsForHost(cfg, host, s, &infra.IdentityClient{
			HTTPClient: &http.Client{Transport: roundTripOK()},
		}, crypto)
	}

	runAuthLogin(authLoginCmd, nil)
	assert.True(t, rec.called)
	blob, err := infra.LoadVaultUnlock(store, config.DefaultHostID)
	require.NoError(t, err)
	assert.Equal(t, "keep-me", blob)
	_, err = infra.LoadAPICredentials(store, config.DefaultHostID)
	require.NoError(t, err)
}

func TestRunAuthLogin_EmptyPassword(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	store := infra.NewMemorySecretStore()
	crypto := &infra.MockCryptoSession{ExportBlob: "nope"}

	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	inputPassword = func() (string, error) { return "", nil }
	newAuthClient = func(cfg *config.Config, host *config.Host, s infra.SecretStore) authClient {
		return infra.NewApiBwClientWithDepsForHost(cfg, host, s, &infra.IdentityClient{
			HTTPClient: &http.Client{Transport: roundTripOK()},
		}, crypto)
	}

	runAuthLogin(authLoginCmd, nil)
	assert.True(t, rec.called)
	_, err := infra.LoadVaultUnlock(store, config.DefaultHostID)
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
}

func TestBoundary_LockKeepsAPIKey_LogoutClears(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	store := infra.NewMemorySecretStore()
	require.NoError(t, infra.SaveAPICredentials(store, config.DefaultHostID, infra.APICredentials{ClientID: "id", ClientSecret: "sec"}))
	require.NoError(t, infra.SaveVaultUnlock(store, config.DefaultHostID, "blob"))

	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	runLock(lockCmd, nil)
	assert.False(t, rec.called)
	_, err := infra.LoadAPICredentials(store, config.DefaultHostID)
	require.NoError(t, err)
	_, err = infra.LoadVaultUnlock(store, config.DefaultHostID)
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)

	require.NoError(t, infra.SaveVaultUnlock(store, config.DefaultHostID, "blob2"))
	runAuthLogout(authLogoutCmd, nil)
	assert.False(t, rec.called)
	_, err = infra.LoadAPICredentials(store, config.DefaultHostID)
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)

	runUnlock(unlockCmd, nil)
	assert.True(t, rec.called)
}
