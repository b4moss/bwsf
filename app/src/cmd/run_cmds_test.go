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
	origBwFromCfg := newBwClientFromConfig
	origBwForHost := newBwClientForHost
	origFS := newFileSystem
	origLog := newLogger
	origSess := newSessionStore
	origStore := newSecretStore
	origUnlock := newUnlockClient
	origExit := exitFunc
	origConfirm := confirmOverwrite
	origSelect := selectCleanMismatch
	origPass := inputPassword

	rec := &exitRecorder{}
	checkBwInstalled = func() (bool, string) { return true, "/mock/bw" }
	if bw != nil {
		newBwClient = func() core.BwClient { return bw }
		newBwClientFromConfig = func(cfg *config.Config) core.BwClient { return bw }
		newBwClientForHost = func(cfg *config.Config, host *config.Host) core.BwClient { return bw }
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
		newBwClientFromConfig = origBwFromCfg
		newBwClientForHost = origBwForHost
		newFileSystem = origFS
		newLogger = origLog
		newSessionStore = origSess
		newSecretStore = origStore
		newUnlockClient = origUnlock
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
	t.Setenv("USERPROFILE", home)
	return home
}

func writeMinimalConfig(t *testing.T) {
	t.Helper()
	cfg := config.NewEmptyConfig()
	cfg.Settings.Hosts = []config.Host{{
		ID:            config.DefaultHostID,
		Type:          config.HostTypeCloud,
		HostURL:       config.DefaultCloudURL,
		Email:         "test@example.com",
		TargetSection: config.DefaultFolderName,
		IsDefault:     true,
	}}
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

func TestRunList_NoHost(t *testing.T) {
	withTempHome(t)
	require.NoError(t, config.SaveConfig(config.NewEmptyConfig()))
	rec := stubCmdDeps(t, testutil.NewMockBwClient(), nil)
	runList(listCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
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
	rec := stubCmdDeps(t, bw, infra.NewFileSystem())

	setupHostType = "cloud"
	setupEmail = "user@example.com"
	setupFolder = "my-envs"
	migrateYes = true
	t.Cleanup(func() {
		setupHostType, setupEmail, setupURL, setupFolder = "", "", "", ""
		setupSkipHost = false
		setupSaveFiles = nil
		migrateYes = false
	})

	runSetup(setupCmd, nil)
	assert.False(t, rec.called)

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	h := cfg.DefaultHost()
	require.NotNil(t, h)
	assert.Equal(t, "my-envs", h.TargetSection)
}

func TestRunSetup_InvalidFlags(t *testing.T) {
	rec := stubCmdDeps(t, nil, nil)
	setupHostType = "cloud"
	setupEmail = ""
	t.Cleanup(func() {
		setupHostType, setupEmail = "", ""
	})
	runSetup(setupCmd, nil)
	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
}

func TestRunSetup_SkipHost(t *testing.T) {
	withTempHome(t)
	bw := testutil.NewMockBwClient()
	bw.SetupTestData()
	rec := stubCmdDeps(t, bw, infra.NewFileSystem())

	setupSkipHost = true
	setupSaveFiles = []string{".env*", "!.env.local"}
	t.Cleanup(func() {
		setupSkipHost = false
		setupSaveFiles = nil
	})

	runSetup(setupCmd, nil)
	assert.False(t, rec.called)

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	assert.Empty(t, cfg.Settings.Hosts)
	assert.Equal(t, []string{".env*", "!.env.local"}, cfg.Settings.SaveFiles)
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
	rec := stubCmdDeps(t, bw, infra.NewFileSystem())
	setupHostType = "selfhosted"
	setupURL = "https://vault.example"
	setupEmail = "user@example.com"
	migrateYes = true
	t.Cleanup(func() {
		setupHostType, setupEmail, setupURL = "", "", ""
		migrateYes = false
	})

	runSetup(setupCmd, nil)
	assert.False(t, rec.called)

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	h := cfg.DefaultHost()
	require.NotNil(t, h)
	assert.Equal(t, config.HostTypeSelfhost, h.Type)
	assert.Equal(t, "https://vault.example", h.HostURL)
}

func TestRunSetup_NonInteractiveSuccess(t *testing.T) {
	withTempHome(t)

	bw := testutil.NewMockBwClient()
	bw.SetupTestData()
	rec := stubCmdDeps(t, bw, infra.NewFileSystem())

	setupHostType = "cloud"
	setupEmail = "user@example.com"
	setupURL = ""
	setupFolder = ""
	migrateYes = true
	t.Cleanup(func() {
		setupHostType, setupEmail, setupURL, setupFolder = "", "", "", ""
		migrateYes = false
	})

	runSetup(setupCmd, nil)
	assert.False(t, rec.called)

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	h := cfg.DefaultHost()
	require.NotNil(t, h)
	assert.Equal(t, config.HostTypeCloud, h.Type)
	assert.Equal(t, "user@example.com", h.Email)
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

func resetInitFlags(t *testing.T) {
	t.Helper()
	initHost = ""
	initSkipHost = false
	initSaveFiles = nil
	initOverrideProjectName = ""
	initOverrideFlagSet = false
	migrateYes = false
	if f := initCmd.Flags().Lookup("override-project-name"); f != nil {
		f.Changed = false
	}
}

func setInitOverrideFlag(t *testing.T, value string) {
	t.Helper()
	require.NoError(t, initCmd.Flags().Set("override-project-name", value))
	initOverrideProjectName = value
	initOverrideFlagSet = true
}

func readProjectConfigCWD(t *testing.T, dir string) *config.ProjectConfig {
	t.Helper()
	pc, err := config.LoadProjectConfigFile(config.GetProjectConfigWritePath(dir))
	require.NoError(t, err)
	return pc
}

func TestRunInit_NoGlobalConfig(t *testing.T) {
	withTempHome(t)
	dir, _ := chdirTempProject(t)
	rec := stubCmdDeps(t, nil, nil)
	resetInitFlags(t)
	t.Cleanup(func() { resetInitFlags(t) })

	initSkipHost = true
	migrateYes = true
	runInit(initCmd, nil)

	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
	assert.NoFileExists(t, config.GetProjectConfigWritePath(dir))
	_, err := os.Stat(filepath.Join(dir, ".bwsf"))
	assert.True(t, os.IsNotExist(err))
}

func TestRunInit_EmptyHosts(t *testing.T) {
	withTempHome(t)
	require.NoError(t, config.SaveConfig(config.NewEmptyConfig()))
	dir, _ := chdirTempProject(t)
	rec := stubCmdDeps(t, nil, nil)
	resetInitFlags(t)
	t.Cleanup(func() { resetInitFlags(t) })

	initSkipHost = true
	migrateYes = true
	runInit(initCmd, nil)

	assert.False(t, rec.called)
	pc := readProjectConfigCWD(t, dir)
	assert.Empty(t, pc.Host)
	assert.Empty(t, pc.SaveFiles)
	assert.Empty(t, pc.OverrideProjectName)
}

func TestRunInit_HostSelect(t *testing.T) {
	withTempHome(t)
	cfg := config.NewEmptyConfig()
	cfg.Settings.Hosts = []config.Host{
		{ID: "default", Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, Email: "a@b.c", TargetSection: "dotenvs", IsDefault: true},
		{ID: "work", Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, Email: "w@b.c", TargetSection: "dotenvs"},
	}
	require.NoError(t, config.SaveConfig(cfg))
	dir, _ := chdirTempProject(t)
	rec := stubCmdDeps(t, nil, nil)
	resetInitFlags(t)
	t.Cleanup(func() { resetInitFlags(t) })

	initHost = "work"
	migrateYes = true
	runInit(initCmd, nil)

	assert.False(t, rec.called)
	pc := readProjectConfigCWD(t, dir)
	assert.Equal(t, "work", pc.Host)
}

func TestRunInit_SkipHost(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	dir, _ := chdirTempProject(t)
	rec := stubCmdDeps(t, nil, nil)
	resetInitFlags(t)
	t.Cleanup(func() { resetInitFlags(t) })

	initSkipHost = true
	migrateYes = true
	runInit(initCmd, nil)

	assert.False(t, rec.called)
	pc := readProjectConfigCWD(t, dir)
	assert.Empty(t, pc.Host)
}

func TestRunInit_UnknownHost(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	dir, _ := chdirTempProject(t)
	rec := stubCmdDeps(t, nil, nil)
	resetInitFlags(t)
	t.Cleanup(func() { resetInitFlags(t) })

	initHost = "missing"
	migrateYes = true
	runInit(initCmd, nil)

	assert.True(t, rec.called)
	assert.Equal(t, 1, rec.code)
	assert.NoFileExists(t, config.GetProjectConfigWritePath(dir))
}

func TestRunInit_SaveFilesAndOverride(t *testing.T) {
	withTempHome(t)
	require.NoError(t, config.SaveConfig(config.NewEmptyConfig()))
	dir, _ := chdirTempProject(t)
	rec := stubCmdDeps(t, nil, nil)
	resetInitFlags(t)
	t.Cleanup(func() { resetInitFlags(t) })

	initSkipHost = true
	initSaveFiles = []string{".env*", "!.env.local"}
	setInitOverrideFlag(t, "my-api")
	migrateYes = true
	runInit(initCmd, nil)

	assert.False(t, rec.called)
	pc := readProjectConfigCWD(t, dir)
	assert.Equal(t, []string{".env*", "!.env.local"}, pc.SaveFiles)
	assert.Equal(t, "my-api", pc.OverrideProjectName)
}

func TestRunInit_OverwriteYes(t *testing.T) {
	withTempHome(t)
	require.NoError(t, config.SaveConfig(config.NewEmptyConfig()))
	dir, _ := chdirTempProject(t)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".bwsf"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".bwsf", "config.jsonc"), []byte(`{"host":"old"}`), 0o600))

	rec := stubCmdDeps(t, nil, nil)
	resetInitFlags(t)
	t.Cleanup(func() { resetInitFlags(t) })

	initSkipHost = true
	migrateYes = true
	runInit(initCmd, nil)

	assert.False(t, rec.called)
	pc := readProjectConfigCWD(t, dir)
	assert.Empty(t, pc.Host)
}

func TestRunInit_OverwriteConfirmNo(t *testing.T) {
	withTempHome(t)
	require.NoError(t, config.SaveConfig(config.NewEmptyConfig()))
	dir, _ := chdirTempProject(t)
	existing := filepath.Join(dir, ".bwsf", "config.jsonc")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".bwsf"), 0o755))
	require.NoError(t, os.WriteFile(existing, []byte(`{"host":"keep"}`), 0o600))

	origConfirm := confirmInitOverwrite
	confirmInitOverwrite = func(string) (bool, error) { return false, nil }
	t.Cleanup(func() { confirmInitOverwrite = origConfirm })

	rec := stubCmdDeps(t, nil, nil)
	resetInitFlags(t)
	t.Cleanup(func() { resetInitFlags(t) })

	initSkipHost = true
	runInit(initCmd, nil)

	assert.True(t, rec.called)
	data, err := os.ReadFile(existing)
	require.NoError(t, err)
	assert.Contains(t, string(data), "keep")
}

func TestRunInit_OverwriteConfirmYesConvertsJSON(t *testing.T) {
	withTempHome(t)
	require.NoError(t, config.SaveConfig(config.NewEmptyConfig()))
	dir, _ := chdirTempProject(t)
	bwsf := filepath.Join(dir, ".bwsf")
	require.NoError(t, os.MkdirAll(bwsf, 0o755))
	jsonPath := filepath.Join(bwsf, "config.json")
	require.NoError(t, os.WriteFile(jsonPath, []byte(`{"host":"old"}`), 0o600))

	origConfirm := confirmInitOverwrite
	confirmInitOverwrite = func(string) (bool, error) { return true, nil }
	t.Cleanup(func() { confirmInitOverwrite = origConfirm })

	rec := stubCmdDeps(t, nil, nil)
	resetInitFlags(t)
	t.Cleanup(func() { resetInitFlags(t) })

	initSkipHost = true
	runInit(initCmd, nil)

	assert.False(t, rec.called)
	require.FileExists(t, config.GetProjectConfigWritePath(dir))
	assert.NoFileExists(t, jsonPath)
}

func TestRunInit_WritesCWDNotGitRoot(t *testing.T) {
	withTempHome(t)
	require.NoError(t, config.SaveConfig(config.NewEmptyConfig()))

	repo := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))
	sub := filepath.Join(repo, "app")
	require.NoError(t, os.MkdirAll(sub, 0o755))

	orig, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(orig) })
	require.NoError(t, os.Chdir(sub))

	rec := stubCmdDeps(t, nil, nil)
	resetInitFlags(t)
	t.Cleanup(func() { resetInitFlags(t) })

	initSkipHost = true
	migrateYes = true
	runInit(initCmd, nil)

	assert.False(t, rec.called)
	require.FileExists(t, config.GetProjectConfigWritePath(sub))
	assert.NoFileExists(t, config.GetProjectConfigWritePath(repo))
}

func TestRunInit_InteractiveHostAndSaveFiles(t *testing.T) {
	withTempHome(t)
	cfg := config.NewEmptyConfig()
	cfg.Settings.Hosts = []config.Host{
		{ID: "default", Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, Email: "a@b.c", TargetSection: "dotenvs", IsDefault: true},
		{ID: "work", Type: config.HostTypeCloud, HostURL: config.DefaultCloudURL, Email: "w@b.c", TargetSection: "dotenvs"},
	}
	require.NoError(t, config.SaveConfig(cfg))
	dir, _ := chdirTempProject(t)

	origSelectHost := selectInitHostID
	origSF := selectSaveFilesAction
	origGlobs := inputSaveFilesGlobs
	origOV := selectOverrideNameAction
	selectInitHostID = func(*config.Config) (string, error) { return "work", nil }
	selectSaveFilesAction = func() (string, error) { return "set", nil }
	inputSaveFilesGlobs = func() ([]string, error) { return []string{".env*", "!.env.local"}, nil }
	selectOverrideNameAction = func() (string, error) { return "unset", nil }
	t.Cleanup(func() {
		selectInitHostID = origSelectHost
		selectSaveFilesAction = origSF
		inputSaveFilesGlobs = origGlobs
		selectOverrideNameAction = origOV
	})

	rec := stubCmdDeps(t, nil, nil)
	resetInitFlags(t)
	t.Cleanup(func() { resetInitFlags(t) })

	runInit(initCmd, nil)

	assert.False(t, rec.called)
	pc := readProjectConfigCWD(t, dir)
	assert.Equal(t, "work", pc.Host)
	assert.Equal(t, []string{".env*", "!.env.local"}, pc.SaveFiles)
	assert.Empty(t, pc.OverrideProjectName)
}

func TestRunInit_InteractiveSkipHost(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	dir, _ := chdirTempProject(t)

	origSelectHost := selectInitHostID
	origSF := selectSaveFilesAction
	origOV := selectOverrideNameAction
	origInput := inputOverrideProjectName
	selectInitHostID = func(*config.Config) (string, error) { return "", nil }
	selectSaveFilesAction = func() (string, error) { return "unset", nil }
	selectOverrideNameAction = func() (string, error) { return "set", nil }
	inputOverrideProjectName = func() (string, error) { return "my-api", nil }
	t.Cleanup(func() {
		selectInitHostID = origSelectHost
		selectSaveFilesAction = origSF
		selectOverrideNameAction = origOV
		inputOverrideProjectName = origInput
	})

	rec := stubCmdDeps(t, nil, nil)
	resetInitFlags(t)
	t.Cleanup(func() { resetInitFlags(t) })

	runInit(initCmd, nil)

	assert.False(t, rec.called)
	pc := readProjectConfigCWD(t, dir)
	assert.Empty(t, pc.Host)
	assert.Equal(t, "my-api", pc.OverrideProjectName)
}

func TestRunInit_EmptyHosts_NoHostPrompt(t *testing.T) {
	withTempHome(t)
	require.NoError(t, config.SaveConfig(config.NewEmptyConfig()))
	dir, _ := chdirTempProject(t)

	hostCalled := false
	origSelectHost := selectInitHostID
	origSF := selectSaveFilesAction
	origOV := selectOverrideNameAction
	selectInitHostID = func(*config.Config) (string, error) {
		hostCalled = true
		return "", nil
	}
	selectSaveFilesAction = func() (string, error) { return "unset", nil }
	selectOverrideNameAction = func() (string, error) { return "unset", nil }
	t.Cleanup(func() {
		selectInitHostID = origSelectHost
		selectSaveFilesAction = origSF
		selectOverrideNameAction = origOV
	})

	rec := stubCmdDeps(t, nil, nil)
	resetInitFlags(t)
	t.Cleanup(func() { resetInitFlags(t) })

	runInit(initCmd, nil)

	assert.False(t, rec.called)
	assert.False(t, hostCalled)
	assert.Empty(t, readProjectConfigCWD(t, dir).Host)
}

func TestRunInit_EmptySaveFilesGlobsUnset(t *testing.T) {
	withTempHome(t)
	writeMinimalConfig(t)
	dir, _ := chdirTempProject(t)

	origSelectHost := selectInitHostID
	origSF := selectSaveFilesAction
	origGlobs := inputSaveFilesGlobs
	origOV := selectOverrideNameAction
	selectInitHostID = func(*config.Config) (string, error) { return "", nil }
	selectSaveFilesAction = func() (string, error) { return "set", nil }
	inputSaveFilesGlobs = func() ([]string, error) { return []string{"  ", ","}, nil }
	selectOverrideNameAction = func() (string, error) { return "unset", nil }
	t.Cleanup(func() {
		selectInitHostID = origSelectHost
		selectSaveFilesAction = origSF
		inputSaveFilesGlobs = origGlobs
		selectOverrideNameAction = origOV
	})

	rec := stubCmdDeps(t, nil, nil)
	resetInitFlags(t)
	t.Cleanup(func() { resetInitFlags(t) })

	runInit(initCmd, nil)

	assert.False(t, rec.called)
	data, err := os.ReadFile(config.GetProjectConfigWritePath(dir))
	require.NoError(t, err)
	assert.NotContains(t, string(data), "save_files")
}
