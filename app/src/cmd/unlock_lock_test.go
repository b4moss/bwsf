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

func TestUnlockLockCmd_Registered(t *testing.T) {
	foundUnlock, foundLock := false, false
	for _, c := range rootCmd.Commands() {
		switch c.Use {
		case "unlock":
			foundUnlock = true
		case "lock":
			foundLock = true
		}
	}
	assert.True(t, foundUnlock)
	assert.True(t, foundLock)
	assert.NotNil(t, unlockCmd.Flags().Lookup("host"))
	assert.NotNil(t, lockCmd.Flags().Lookup("host"))
	assert.NotNil(t, lockCmd.Flags().Lookup("all"))
	assert.Contains(t, unlockCmd.Long, "auth")
	assert.Contains(t, lockCmd.Long, "API")
}

func TestRunLock_AllAndHostConflict(t *testing.T) {
	rec := stubCmdDeps(t, nil, nil)
	lockAll = true
	lockHost = "work"
	t.Cleanup(func() { lockAll = false; lockHost = "" })
	runLock(lockCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

func TestRunLock_AllEmptyHosts(t *testing.T) {
	withTempHome(t)
	require.NoError(t, config.SaveConfig(config.NewEmptyConfig()))
	rec := stubCmdDeps(t, nil, nil)
	lockAll = true
	t.Cleanup(func() { lockAll = false })
	runLock(lockCmd, nil)
	assert.False(t, rec.called)
}

func TestRunLock_UnknownHost(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	rec := stubCmdDeps(t, nil, nil)
	lockHost = "missing"
	t.Cleanup(func() { lockHost = "" })
	runLock(lockCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

func TestRunLock_AllDeletesVaultUnlockKeepsAPIKeys(t *testing.T) {
	withTempHome(t)
	cfg := config.NewEmptyConfig()
	cfg.Settings.Hosts = []config.Host{
		{ID: config.DefaultHostID, Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, TargetSection: "dotenvs", IsDefault: true, Email: "a@example.com"},
		{ID: "work", Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, TargetSection: "dotenvs", Email: "a@example.com"},
	}
	require.NoError(t, config.SaveConfig(cfg))

	store := infra.NewMemorySecretStore()
	require.NoError(t, infra.SaveVaultUnlock(store, config.DefaultHostID, "d"))
	require.NoError(t, infra.SaveVaultUnlock(store, "work", "w"))
	require.NoError(t, infra.SaveAPICredentials(store, config.DefaultHostID, infra.APICredentials{ClientID: "id", ClientSecret: "sec"}))
	require.NoError(t, infra.SaveAPICredentials(store, "work", infra.APICredentials{ClientID: "id2", ClientSecret: "sec2"}))

	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	lockAll = true
	t.Cleanup(func() { lockAll = false })
	runLock(lockCmd, nil)
	assert.False(t, rec.called)

	_, err := infra.LoadVaultUnlock(store, config.DefaultHostID)
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
	_, err = infra.LoadVaultUnlock(store, "work")
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
	_, err = infra.LoadAPICredentials(store, config.DefaultHostID)
	require.NoError(t, err)
	_, err = infra.LoadAPICredentials(store, "work")
	require.NoError(t, err)
}

func TestRunLock_SingleHostLeavesOther(t *testing.T) {
	withTempHome(t)
	cfg := config.NewEmptyConfig()
	cfg.Settings.Hosts = []config.Host{
		{ID: config.DefaultHostID, Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, TargetSection: "dotenvs", IsDefault: true, Email: "a@example.com"},
		{ID: "work", Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, TargetSection: "dotenvs", Email: "a@example.com"},
	}
	require.NoError(t, config.SaveConfig(cfg))

	store := infra.NewMemorySecretStore()
	require.NoError(t, infra.SaveVaultUnlock(store, config.DefaultHostID, "default-blob"))
	require.NoError(t, infra.SaveVaultUnlock(store, "work", "work-blob"))

	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	lockHost = "work"
	t.Cleanup(func() { lockHost = "" })
	runLock(lockCmd, nil)
	assert.False(t, rec.called)

	_, err := infra.LoadVaultUnlock(store, "work")
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
	blob, err := infra.LoadVaultUnlock(store, config.DefaultHostID)
	require.NoError(t, err)
	assert.Equal(t, "default-blob", blob)
}

func TestRunUnlock_NoAPIKey(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	store := infra.NewMemorySecretStore()
	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	runUnlock(unlockCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

func TestRunUnlock_EmptyPassword(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	store := infra.NewMemorySecretStore()
	require.NoError(t, infra.SaveAPICredentials(store, config.DefaultHostID, infra.APICredentials{ClientID: "id", ClientSecret: "sec"}))
	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	inputPassword = func() (string, error) { return "", nil }
	runUnlock(unlockCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
	_, err := infra.LoadVaultUnlock(store, config.DefaultHostID)
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
}

func TestRunUnlock_SuccessPersistsVaultUnlock(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	store := infra.NewMemorySecretStore()
	require.NoError(t, infra.SaveAPICredentials(store, config.DefaultHostID, infra.APICredentials{ClientID: "id", ClientSecret: "sec"}))

	crypto := &infra.MockCryptoSession{ExportBlob: "v1:unlock-blob"}
	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	newUnlockClient = func(cfg *config.Config, host *config.Host, s infra.SecretStore) unlockClient {
		return infra.NewApiBwClientWithDepsForHost(cfg, host, s, &infra.IdentityClient{
			HTTPClient: &http.Client{Transport: roundTripOK()},
		}, crypto)
	}

	runUnlock(unlockCmd, nil)
	assert.False(t, rec.called)
	blob, err := infra.LoadVaultUnlock(store, config.DefaultHostID)
	require.NoError(t, err)
	assert.Equal(t, "v1:unlock-blob", blob)
	assert.False(t, crypto.Unlocked(), "ClearSession should lock memory crypto")
}

func TestRunUnlock_FailedUnlockKeepsExistingBlob(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	store := infra.NewMemorySecretStore()
	require.NoError(t, infra.SaveAPICredentials(store, config.DefaultHostID, infra.APICredentials{ClientID: "id", ClientSecret: "sec"}))
	require.NoError(t, infra.SaveVaultUnlock(store, config.DefaultHostID, "keep-me"))

	crypto := &infra.MockCryptoSession{UnlockErr: assertErr("bad password")}
	rec := stubCmdDeps(t, nil, nil)
	newSecretStore = func() infra.SecretStore { return store }
	newUnlockClient = func(cfg *config.Config, host *config.Host, s infra.SecretStore) unlockClient {
		return infra.NewApiBwClientWithDepsForHost(cfg, host, s, &infra.IdentityClient{
			HTTPClient: &http.Client{Transport: roundTripOK()},
		}, crypto)
	}

	runUnlock(unlockCmd, nil)
	assert.True(t, rec.called)
	blob, err := infra.LoadVaultUnlock(store, config.DefaultHostID)
	require.NoError(t, err)
	assert.Equal(t, "keep-me", blob)
}

type assertErr string

func (e assertErr) Error() string { return string(e) }

func roundTripOK() http.RoundTripper {
	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"access_token":"tok","expires_in":3600}`)),
			Header:     make(http.Header),
		}, nil
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestLoadProjectHostID(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".bwsf"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".bwsf", "config.json"), []byte(`{"host":"work"}`), 0o644))
	orig, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(orig) })
	assert.Equal(t, "work", loadProjectHostID())
}

func TestSessionLifecycle_ClearSessionKeepsKeychain(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	store := infra.NewMemorySecretStore()
	require.NoError(t, infra.SaveAPICredentials(store, config.DefaultHostID, infra.APICredentials{ClientID: "id", ClientSecret: "sec"}))
	require.NoError(t, infra.SaveVaultUnlock(store, config.DefaultHostID, "v1:session"))

	crypto := &infra.MockCryptoSession{}
	client := infra.NewApiBwClientWithDepsForHost(
		loadConfigOrEmpty(),
		loadConfigOrEmpty().DefaultHost(),
		store,
		&infra.IdentityClient{HTTPClient: &http.Client{Transport: roundTripOK()}},
		crypto,
	)
	ok, err := client.TryRestoreVaultUnlock()
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, client.IsUnlocked())

	client.ClearSession()
	assert.False(t, client.IsUnlocked())
	blob, err := infra.LoadVaultUnlock(store, config.DefaultHostID)
	require.NoError(t, err)
	assert.Equal(t, "v1:session", blob)
	_, err = infra.LoadAPICredentials(store, config.DefaultHostID)
	require.NoError(t, err)

	ok, err = client.TryRestoreVaultUnlock()
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, client.IsUnlocked())
}
