package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"bwsf/src/config"
	"bwsf/src/core"
	"bwsf/src/infra"
	"bwsf/src/testutil"
	"bwsf/src/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type noopSessionStore struct{}

func (noopSessionStore) Get() (string, error) { return "", nil }
func (noopSessionStore) Set(string) error     { return nil }
func (noopSessionStore) Delete() error        { return nil }

type exitRecorder struct {
	called bool
	code   int
}

func stubCmdDeps(t *testing.T, bw core.BwClient, fs core.FileSystem) *exitRecorder {
	t.Helper()

	origCheck := checkBwInstalled
	origBw := newBwClient
	origFS := newFileSystem
	origLog := newLogger
	origSess := newSessionStore
	origExit := exitFunc
	origConfirm := confirmOverwrite
	origSelect := selectCleanMismatch
	origPass := inputPassword

	rec := &exitRecorder{}
	checkBwInstalled = func() (bool, string) { return true, "/mock/bw" }
	if bw != nil {
		newBwClient = func() core.BwClient { return bw }
	}
	if fs != nil {
		newFileSystem = func() core.FileSystem { return fs }
	} else {
		newFileSystem = func() core.FileSystem { return infra.NewFileSystem() }
	}
	newLogger = func() core.Logger { return infra.NewLogger() }
	newSessionStore = func() core.SessionStore { return noopSessionStore{} }
	exitFunc = func(code int) {
		rec.called = true
		rec.code = code
	}
	inputPassword = func() (string, error) { return "testpassword", nil }
	confirmOverwrite = func(message string) (bool, error) { return true, nil }
	selectCleanMismatch = func(mismatchedFiles []string) (string, error) {
		return utils.CleanMismatchAbort, nil
	}

	t.Cleanup(func() {
		checkBwInstalled = origCheck
		newBwClient = origBw
		newFileSystem = origFS
		newLogger = origLog
		newSessionStore = origSess
		exitFunc = origExit
		confirmOverwrite = origConfirm
		selectCleanMismatch = origSelect
		inputPassword = origPass
	})
	return rec
}

func withTempHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // Windows-safe; harmless on darwin
	return home
}

func writeMinimalConfig(t *testing.T) {
	t.Helper()
	cfg := &config.Config{
		HostType: "cloud",
		Email:    "test@example.com",
	}
	require.NoError(t, config.SaveConfig(cfg))
}

func chdirTempProject(t *testing.T) (dir, projectName string) {
	t.Helper()
	dir = t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))
	orig, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(orig) })
	require.NoError(t, os.Chdir(dir))
	return dir, filepath.Base(dir)
}

func TestRunList_BwNotInstalled(t *testing.T) {
	rec := stubCmdDeps(t, nil, nil)
	checkBwInstalled = func() (bool, string) { return false, "" }

	runList(listCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

func TestRunList_Success(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)

	bw := testutil.NewMockBwClient()
	bw.SetupTestData()
	require.NoError(t, bw.CreateNoteItem("folder-dotenvs-id", "demo", `{"lines":["A=1"]}`))

	rec := stubCmdDeps(t, bw, nil)
	runList(listCmd, nil)
	assert.False(t, rec.called)
}

func TestRunList_Empty(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)

	bw := testutil.NewMockBwClient()
	bw.SetupTestData()
	rec := stubCmdDeps(t, bw, nil)
	runList(listCmd, nil)
	assert.False(t, rec.called)
}

func TestRunPush_Success(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	dir, _ := chdirTempProject(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".env"), []byte("KEY=value\n"), 0o644))

	bw := testutil.NewMockBwClient()
	bw.SetupTestData()
	rec := stubCmdDeps(t, bw, infra.NewFileSystem())

	runPush(pushCmd, nil)
	assert.False(t, rec.called)
	assert.Equal(t, 1, bw.GetItemCount())
}

func TestRunPush_BwNotInstalled(t *testing.T) {
	rec := stubCmdDeps(t, nil, nil)
	checkBwInstalled = func() (bool, string) { return false, "" }
	runPush(pushCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

func TestRunPush_NoManagedFiles(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	_, _ = chdirTempProject(t)

	bw := testutil.NewMockBwClient()
	bw.SetupTestData()
	rec := stubCmdDeps(t, bw, infra.NewFileSystem())
	runPush(pushCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

func TestRunClean_NoManagedFiles(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	_, _ = chdirTempProject(t)

	bw := testutil.NewMockBwClient()
	bw.SetupTestData()
	rec := stubCmdDeps(t, bw, infra.NewFileSystem())
	runClean(cleanCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

func TestRunSetup_WithFolderFlag(t *testing.T) {
	withTempHome(t)

	bw := testutil.NewMockBwClient()
	bw.SetupTestData()
	bw.LoginFunc = func(email, password, serverURL string) error { return nil }

	rec := stubCmdDeps(t, bw, infra.NewFileSystem())

	setupHostType = "cloud"
	setupEmail = "user@example.com"
	setupPassword = "secret"
	setupFolder = "my-envs"
	setupYes = true
	t.Cleanup(func() {
		setupHostType, setupEmail, setupPassword, setupURL, setupFolder = "", "", "", "", ""
		setupYes = false
	})

	runSetup(setupCmd, nil)
	assert.False(t, rec.called)

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "my-envs", cfg.FolderName)
}

func TestRunSetup_InvalidFlags(t *testing.T) {
	rec := stubCmdDeps(t, nil, nil)
	setupHostType = "cloud"
	setupEmail = ""
	setupPassword = ""
	t.Cleanup(func() {
		setupHostType, setupEmail, setupPassword = "", "", ""
	})
	runSetup(setupCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

func TestRunPull_Success(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	dir, projectName := chdirTempProject(t)

	bw := testutil.NewMockBwClient()
	bw.SetupTestData()
	notes, err := json.Marshal(map[string]core.EnvData{
		".env": {Lines: []string{"KEY=from-remote"}},
	})
	require.NoError(t, err)
	require.NoError(t, bw.CreateNoteItem("folder-dotenvs-id", projectName, string(notes)))

	rec := stubCmdDeps(t, bw, infra.NewFileSystem())
	runPull(pullCmd, nil)
	assert.False(t, rec.called)

	data, err := os.ReadFile(filepath.Join(dir, ".env"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "KEY=from-remote")
}

func TestRunPull_NoRemoteItems(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	_, _ = chdirTempProject(t)

	bw := testutil.NewMockBwClient()
	bw.SetupTestData()
	rec := stubCmdDeps(t, bw, infra.NewFileSystem())
	runPull(pullCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

func TestRunClean_Success(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	dir, projectName := chdirTempProject(t)
	envPath := filepath.Join(dir, ".env")
	require.NoError(t, os.WriteFile(envPath, []byte("KEY=value\n"), 0o644))

	bw := testutil.NewMockBwClient()
	bw.SetupTestData()
	fs := infra.NewFileSystem()
	rec := stubCmdDeps(t, bw, fs)

	// Seed remote to match local via push path first.
	runPush(pushCmd, nil)
	require.False(t, rec.called)
	require.Equal(t, 1, bw.GetItemCount())
	_ = projectName

	rec.called = false
	runClean(cleanCmd, nil)
	assert.False(t, rec.called)
	_, err := os.Stat(envPath)
	assert.True(t, os.IsNotExist(err))
}

func TestRunClean_AbortedKeepsFiles(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	dir, projectName := chdirTempProject(t)
	envPath := filepath.Join(dir, ".env")
	require.NoError(t, os.WriteFile(envPath, []byte("KEY=local\n"), 0o644))

	bw := testutil.NewMockBwClient()
	bw.SetupTestData()
	notes, err := json.Marshal(map[string]core.EnvData{
		".env": {Lines: []string{"KEY=remote"}},
	})
	require.NoError(t, err)
	require.NoError(t, bw.CreateNoteItem("folder-dotenvs-id", projectName, string(notes)))

	rec := stubCmdDeps(t, bw, infra.NewFileSystem())
	selectCleanMismatch = func(mismatchedFiles []string) (string, error) {
		return utils.CleanMismatchAbort, nil
	}
	runClean(cleanCmd, nil)
	assert.False(t, rec.called)
	_, err = os.Stat(envPath)
	assert.NoError(t, err)
}

func TestRunClean_RemoveLocalOnMismatch(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	dir, projectName := chdirTempProject(t)
	envPath := filepath.Join(dir, ".env")
	require.NoError(t, os.WriteFile(envPath, []byte("KEY=local\n"), 0o644))

	bw := testutil.NewMockBwClient()
	bw.SetupTestData()
	notes, err := json.Marshal(map[string]core.EnvData{
		".env": {Lines: []string{"KEY=remote"}},
	})
	require.NoError(t, err)
	require.NoError(t, bw.CreateNoteItem("folder-dotenvs-id", projectName, string(notes)))

	rec := stubCmdDeps(t, bw, infra.NewFileSystem())
	selectCleanMismatch = func(mismatchedFiles []string) (string, error) {
		return utils.CleanMismatchRemoveLocal, nil
	}
	runClean(cleanCmd, nil)
	assert.False(t, rec.called)
	_, err = os.Stat(envPath)
	assert.True(t, os.IsNotExist(err))
}

func TestRunPull_OverwriteExisting(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	dir, projectName := chdirTempProject(t)
	envPath := filepath.Join(dir, ".env")
	require.NoError(t, os.WriteFile(envPath, []byte("KEY=old\n"), 0o644))

	bw := testutil.NewMockBwClient()
	bw.SetupTestData()
	notes, err := json.Marshal(map[string]core.EnvData{
		".env": {Lines: []string{"KEY=from-remote"}},
	})
	require.NoError(t, err)
	require.NoError(t, bw.CreateNoteItem("folder-dotenvs-id", projectName, string(notes)))

	rec := stubCmdDeps(t, bw, infra.NewFileSystem())
	confirmOverwrite = func(message string) (bool, error) { return true, nil }
	runPull(pullCmd, nil)
	assert.False(t, rec.called)
	data, err := os.ReadFile(envPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "KEY=from-remote")
}

func TestRunSetup_Selfhosted(t *testing.T) {
	withTempHome(t)

	bw := testutil.NewMockBwClient()
	bw.SetupTestData()
	bw.LoginFunc = func(email, password, serverURL string) error {
		assert.Equal(t, "https://vault.example", serverURL)
		return nil
	}

	rec := stubCmdDeps(t, bw, infra.NewFileSystem())
	setupHostType = "selfhosted"
	setupURL = "https://vault.example"
	setupEmail = "user@example.com"
	setupPassword = "secret"
	t.Cleanup(func() {
		setupHostType, setupEmail, setupPassword, setupURL = "", "", "", ""
	})

	runSetup(setupCmd, nil)
	assert.False(t, rec.called)
}

func TestRunSetup_NonInteractiveSuccess(t *testing.T) {
	withTempHome(t)

	bw := testutil.NewMockBwClient()
	bw.SetupTestData()
	// Folder already exists from SetupTestData; Login still required.
	bw.LoginFunc = func(email, password, serverURL string) error { return nil }

	rec := stubCmdDeps(t, bw, infra.NewFileSystem())

	setupHostType = "cloud"
	setupEmail = "user@example.com"
	setupPassword = "secret"
	setupURL = ""
	setupFolder = ""
	setupYes = false
	t.Cleanup(func() {
		setupHostType, setupEmail, setupPassword, setupURL, setupFolder = "", "", "", "", ""
		setupYes = false
	})

	runSetup(setupCmd, nil)
	assert.False(t, rec.called)

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "cloud", cfg.HostType)
	assert.Equal(t, "user@example.com", cfg.Email)
}

func TestRunSetup_BwNotInstalled(t *testing.T) {
	rec := stubCmdDeps(t, nil, nil)
	checkBwInstalled = func() (bool, string) { return false, "" }
	runSetup(setupCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

func TestRunConfigShow_Success(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	rec := stubCmdDeps(t, nil, nil)
	runConfigShow(configShowCmd, nil)
	assert.False(t, rec.called)
}

func TestRunConfigShow_NoConfig(t *testing.T) {
	withTempHome(t)
	rec := stubCmdDeps(t, nil, nil)
	runConfigShow(configShowCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}
