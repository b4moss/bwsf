package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"bwsf/src/config"
	"bwsf/src/infra"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnlockLockCmd_HelpMentionsAuthBoundary(t *testing.T) {
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

func TestRunLock_SingleHost(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)

	// Use memory store via LockVaultSessionForHost path: we can't inject keyring easily,
	// so exercise helper + resolve path with unknown host for error case.
	rec := stubCmdDeps(t, nil, nil)
	lockHost = "missing"
	t.Cleanup(func() { lockHost = "" })
	runLock(lockCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

func TestRunLock_AllDeletesVaultUnlock(t *testing.T) {
	withTempHome(t)
	cfg := config.NewEmptyConfig()
	cfg.Settings.Hosts = []config.Host{
		{ID: config.DefaultHostID, Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, TargetSection: "dotenvs", IsDefault: true, Email: "a@example.com"},
		{ID: "work", Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, TargetSection: "dotenvs", Email: "a@example.com"},
	}
	require.NoError(t, config.SaveConfig(cfg))

	store := infra.NewMemorySecretStore()
	require.NoError(t, infra.SetVaultUnlock(store, config.DefaultHostID, "d"))
	require.NoError(t, infra.SetVaultUnlock(store, "work", "w"))
	require.NoError(t, infra.SaveAPICredentials(store, config.DefaultHostID, infra.APICredentials{ClientID: "id", ClientSecret: "sec"}))
	require.NoError(t, infra.SaveAPICredentials(store, "work", infra.APICredentials{ClientID: "id2", ClientSecret: "sec2"}))

	for _, id := range []string{config.DefaultHostID, "work"} {
		require.NoError(t, infra.LockVaultSessionForHost(store, id))
	}
	_, err := infra.GetVaultUnlock(store, config.DefaultHostID)
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
	_, err = infra.GetVaultUnlock(store, "work")
	assert.ErrorIs(t, err, infra.ErrSecretNotFound)
	_, err = infra.LoadAPICredentials(store, config.DefaultHostID)
	require.NoError(t, err)
}

func TestRunUnlock_NoAPIKey(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	rec := stubCmdDeps(t, nil, nil)
	// KeyringStore will not find credentials in this HOME
	runUnlock(unlockCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

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
