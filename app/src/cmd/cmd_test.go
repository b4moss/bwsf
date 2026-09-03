package cmd

import (
	"os"
	"testing"

	"bwsf/src/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd_Initialized(t *testing.T) {
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "bwsf", rootCmd.Use)
}

func TestRootCmd_Description(t *testing.T) {
	assert.NotEmpty(t, rootCmd.Short)
	assert.NotEmpty(t, rootCmd.Long)
}

func TestPushCmd_Registered(t *testing.T) {
	assert.NotNil(t, pushCmd)
	assert.Equal(t, "push", pushCmd.Use)
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "push" {
			found = true
			break
		}
	}
	assert.True(t, found, "push command should be registered")
}

func TestPullCmd_Registered(t *testing.T) {
	assert.NotNil(t, pullCmd)
	assert.Equal(t, "pull", pullCmd.Use)
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "pull" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestListCmd_Registered(t *testing.T) {
	assert.NotNil(t, listCmd)
	assert.Equal(t, "list", listCmd.Use)
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "list" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestSetupCmd_Registered(t *testing.T) {
	assert.NotNil(t, setupCmd)
	assert.Equal(t, "setup", setupCmd.Use)
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "setup" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestUnlockLockCmd_Registered(t *testing.T) {
	assert.NotNil(t, unlockCmd)
	assert.Equal(t, "unlock", unlockCmd.Use)
	assert.NotNil(t, unlockCmd.Flags().Lookup("host"))

	assert.NotNil(t, lockCmd)
	assert.Equal(t, "lock", lockCmd.Use)
	assert.NotNil(t, lockCmd.Flags().Lookup("host"))
	assert.NotNil(t, lockCmd.Flags().Lookup("all"))

	foundUnlock, foundLock := false, false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "unlock" {
			foundUnlock = true
		}
		if cmd.Use == "lock" {
			foundLock = true
		}
	}
	assert.True(t, foundUnlock)
	assert.True(t, foundLock)
}

func TestBackendCmd_Registered(t *testing.T) {
	assert.NotNil(t, backendCmd)
	assert.Equal(t, "backend", backendCmd.Use)
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "backend" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestCleanCmd_Registered(t *testing.T) {
	assert.NotNil(t, cleanCmd)
	assert.Equal(t, "clean", cleanCmd.Use)
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "clean" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestPushCmd_FromFlag(t *testing.T) {
	flag := pushCmd.Flags().Lookup("from")
	assert.NotNil(t, flag)
	assert.Equal(t, ".", flag.DefValue)
}

func TestPushCmd_HostFlag(t *testing.T) {
	assert.NotNil(t, pushCmd.Flags().Lookup("host"))
}

func TestSetupCmd_NonInteractiveFlags(t *testing.T) {
	for _, name := range []string{"host-type", "url", "email", "skip-host", "save-files", "folder"} {
		assert.NotNil(t, setupCmd.Flags().Lookup(name), "missing flag --%s", name)
	}
	assert.NotNil(t, rootCmd.PersistentFlags().Lookup("yes"), "missing persistent --yes")
}

func TestValidateSetupNonInteractiveFlags_Partial(t *testing.T) {
	setupHostType = "selfhosted"
	setupURL = ""
	setupEmail = ""
	t.Cleanup(func() {
		setupHostType = ""
		setupURL = ""
		setupEmail = ""
	})

	err := validateSetupNonInteractiveFlags()
	assert.Error(t, err)
}

func TestValidateSetupNonInteractiveFlags_SelfhostedOK(t *testing.T) {
	setupHostType = "selfhosted"
	setupURL = "https://vaultwarden:80"
	setupEmail = "smoke@bwsf.local"
	t.Cleanup(func() {
		setupHostType = ""
		setupURL = ""
		setupEmail = ""
	})

	err := validateSetupNonInteractiveFlags()
	assert.NoError(t, err)
}

func TestValidateSetupNonInteractiveFlags_SkipHost(t *testing.T) {
	setupSkipHost = true
	t.Cleanup(func() { setupSkipHost = false })
	assert.NoError(t, validateSetupNonInteractiveFlags())
}

func TestPullCmd_OutputFlag(t *testing.T) {
	flag := pullCmd.Flags().Lookup("output")
	assert.NotNil(t, flag)
	assert.Equal(t, ".", flag.DefValue)
}

func TestBackendCmd_SetFlagRemoved(t *testing.T) {
	assert.Nil(t, backendCmd.Flags().Lookup("set"))
}

func TestCleanCmd_FromFlag(t *testing.T) {
	flag := cleanCmd.Flags().Lookup("from")
	assert.NotNil(t, flag)
	assert.Equal(t, ".", flag.DefValue)
}

func TestPushCmd_Description(t *testing.T) {
	assert.NotEmpty(t, pushCmd.Short)
	assert.NotEmpty(t, pushCmd.Long)
}

func TestPullCmd_Description(t *testing.T) {
	assert.NotEmpty(t, pullCmd.Short)
	assert.NotEmpty(t, pullCmd.Long)
}

func TestListCmd_Description(t *testing.T) {
	assert.NotEmpty(t, listCmd.Short)
	assert.NotEmpty(t, listCmd.Long)
}

func TestSetupCmd_Description(t *testing.T) {
	assert.NotEmpty(t, setupCmd.Short)
	assert.NotEmpty(t, setupCmd.Long)
	assert.Contains(t, setupCmd.Long, "bwsf auth")
}

func TestBackendCmd_Description(t *testing.T) {
	assert.NotEmpty(t, backendCmd.Short)
	assert.NotEmpty(t, backendCmd.Long)
	assert.Contains(t, backendCmd.Long, "removed")
}

func TestAuthCmd_Registered(t *testing.T) {
	assert.NotNil(t, authCmd)
	assert.Equal(t, "auth", authCmd.Use)
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "auth" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestAuthCmd_ClearFlag(t *testing.T) {
	flag := authCmd.Flags().Lookup("clear")
	assert.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
}

func TestAuthCmd_Description(t *testing.T) {
	assert.NotEmpty(t, authCmd.Short)
	assert.NotEmpty(t, authCmd.Long)
}

func TestBackendCmd_Removed(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	rec := &exitRecorder{}
	origExit := exitFunc
	exitFunc = func(code int) { rec.called = true; rec.code = code }
	t.Cleanup(func() { exitFunc = origExit })

	runBackend(backendCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

func TestCleanCmd_Description(t *testing.T) {
	assert.NotEmpty(t, cleanCmd.Short)
	assert.NotEmpty(t, cleanCmd.Long)
}

func TestConfigCmd_Registered(t *testing.T) {
	assert.NotNil(t, configCmd)
	assert.Equal(t, "config", configCmd.Use)

	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "config" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestAppVersionStamped(t *testing.T) {
	assert.Equal(t, Version, config.AppVersion)
	require.NotEmpty(t, Version)
}
