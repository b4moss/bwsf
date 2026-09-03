package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigPath_Success(t *testing.T) {
	path, err := GetConfigPath()
	assert.NoError(t, err)
	assert.Contains(t, path, ".config/bwsf/config.jsonc")
	homeDir, _ := os.UserHomeDir()
	assert.True(t, filepath.HasPrefix(path, homeDir))
}

func TestLoadConfig_FileNotExist(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfig_V2Success(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "bwsf")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	content := `{
  "schemaVersion": 1,
  "settings": {
    "hosts": [{
      "id": "default",
      "type": "bitwarden-selfhost",
      "host_url": "https://bw.example.com",
      "email": "test@example.com",
      "target_section": "dotenvs",
      "is_default": true
    }]
  }
}`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.json"), []byte(content), 0600))

	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	h := cfg.DefaultHost()
	require.NotNil(t, h)
	assert.Equal(t, HostTypeSelfhost, h.Type)
	assert.Equal(t, "https://bw.example.com", h.HostURL)
	assert.Equal(t, "test@example.com", h.Email)
}

func TestLoadConfig_FlatRequiresMigrate(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "bwsf")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	content := `{"host_type":"selfhosted","selfhosted_url":"https://bw.example.com","email":"test@example.com"}`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.json"), []byte(content), 0600))

	_, err := LoadConfig()
	assert.Error(t, err)

	cfg, err := LoadConfigWithMigrate(MigrateHooks{Yes: true})
	require.NoError(t, err)
	require.NotNil(t, cfg)
	h := cfg.DefaultHost()
	require.NotNil(t, h)
	assert.Equal(t, HostTypeSelfhost, h.Type)
	assert.Equal(t, "test@example.com", h.Email)
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "bwsf")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.json"), []byte("not valid json"), 0600))

	cfg, err := LoadConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfig_BothFilesError(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "bwsf")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{}`), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.jsonc"), []byte(`{}`), 0600))

	_, err := LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "both")
}

func TestSaveConfig_Success(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := NewEmptyConfig()
	cfg.Settings.Hosts = []Host{{
		ID: DefaultHostID, Type: HostTypeCloud, HostURL: DefaultCloudURL,
		Email: "user@example.com", TargetSection: DefaultFolderName, IsDefault: true,
	}}
	require.NoError(t, SaveConfig(cfg))

	jsoncPath := filepath.Join(tmpDir, ".config", "bwsf", "config.jsonc")
	assert.FileExists(t, jsoncPath)
	assert.NoFileExists(t, filepath.Join(tmpDir, ".config", "bwsf", "config.json"))

	loaded, err := LoadConfig()
	require.NoError(t, err)
	h := loaded.DefaultHost()
	require.NotNil(t, h)
	assert.Equal(t, "user@example.com", h.Email)
}

func TestResolveHost(t *testing.T) {
	cfg := NewEmptyConfig()
	cfg.Settings.Hosts = []Host{
		{ID: "work", Type: HostTypeCloud, HostURL: DefaultCloudURL, TargetSection: "dotenvs", IsDefault: false},
		{ID: "default", Type: HostTypeCloud, HostURL: DefaultCloudURL, TargetSection: "dotenvs", IsDefault: true},
	}

	h, err := ResolveHost(cfg, "work", "")
	require.NoError(t, err)
	assert.Equal(t, "work", h.ID)

	h, err = ResolveHost(cfg, "", "work")
	require.NoError(t, err)
	assert.Equal(t, "work", h.ID)

	h, err = ResolveHost(cfg, "", "")
	require.NoError(t, err)
	assert.Equal(t, "default", h.ID)

	_, err = ResolveHost(cfg, "missing", "")
	assert.Error(t, err)
}

func TestFormatConfigShow(t *testing.T) {
	cfg := NewEmptyConfig()
	cfg.Settings.Hosts = []Host{{
		ID: DefaultHostID, Type: HostTypeCloud, HostURL: DefaultCloudURL,
		Email: "a@b.c", TargetSection: "team", IsDefault: true,
	}}
	out := FormatConfigShow(cfg)
	assert.Contains(t, out, "schemaVersion:")
	assert.Contains(t, out, "id: default [default]")
	assert.Contains(t, out, "target_section: team")
}

func TestNewEmptyConfig_Valid(t *testing.T) {
	cfg := NewEmptyConfig()
	assert.NoError(t, cfg.Validate())
	assert.Empty(t, cfg.Settings.Hosts)
}

func TestValidate_RequiresOneDefault(t *testing.T) {
	cfg := NewEmptyConfig()
	cfg.Settings.Hosts = []Host{
		{ID: "a", Type: HostTypeCloud, HostURL: DefaultCloudURL, TargetSection: "dotenvs", IsDefault: false},
	}
	assert.Error(t, cfg.Validate())
}

func TestBannedBackendKey(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "bwsf")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	content := `{"schemaVersion":1,"backend":"api","settings":{"hosts":[]}}`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.json"), []byte(content), 0600))

	_, err := LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backend")
}
