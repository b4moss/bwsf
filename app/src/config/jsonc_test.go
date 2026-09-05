package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func expectedFixtureConfig() *Config {
	cfg := NewEmptyConfig()
	cfg.Settings.Hosts = []Host{{
		ID:            DefaultHostID,
		Type:          HostTypeCloud,
		HostURL:       DefaultCloudURL,
		Email:         "user@example.com",
		TargetSection: DefaultFolderName,
		IsDefault:     true,
	}}
	return cfg
}

func assertConfigEqual(t *testing.T, got *Config) {
	t.Helper()
	want := expectedFixtureConfig()
	require.NotNil(t, got)
	h := got.DefaultHost()
	require.NotNil(t, h)
	wh := want.DefaultHost()
	assert.Equal(t, wh.Type, h.Type)
	assert.Equal(t, wh.HostURL, h.HostURL)
	assert.Equal(t, wh.Email, h.Email)
	assert.Equal(t, wh.TargetSection, h.TargetSection)
}

func assertNoJSONCComments(t *testing.T, text string) {
	t.Helper()
	assert.NotContains(t, text, "\n  //")
	assert.NotContains(t, text, "/*")
	assert.False(t, strings.Contains(text, ",\n}"), "trailing comma before closing brace")
}

func TestUnmarshalConfigJSONC_Fixtures(t *testing.T) {
	files := []string{
		"plain.json",
		"comments.jsonc",
		"trailing_comma.jsonc",
		"comments_and_trailing_comma.jsonc",
	}
	for _, name := range files {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", name))
			require.NoError(t, err)

			var cfg Config
			err = UnmarshalConfigJSONC(data, &cfg)
			require.NoError(t, err)
			assertConfigEqual(t, &cfg)
		})
	}
}

func TestUnmarshalConfigJSONC_Invalid(t *testing.T) {
	t.Run("broken", func(t *testing.T) {
		var cfg Config
		err := UnmarshalConfigJSONC([]byte(`{"schemaVersion":`), &cfg)
		assert.Error(t, err)
	})
	t.Run("empty", func(t *testing.T) {
		var cfg Config
		err := UnmarshalConfigJSONC([]byte(""), &cfg)
		assert.Error(t, err)
	})
	t.Run("whitespace_only", func(t *testing.T) {
		var cfg Config
		err := UnmarshalConfigJSONC([]byte("   \n\t"), &cfg)
		assert.Error(t, err)
	})
}

func TestLoadConfig_JSONCFixture(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	require.NoError(t, os.Setenv("HOME", tmpDir))
	defer os.Setenv("HOME", origHome)

	fixture, err := os.ReadFile(filepath.Join("testdata", "comments_and_trailing_comma.jsonc"))
	require.NoError(t, err)

	configDir := filepath.Join(tmpDir, ".config", "bwsf")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.jsonc"), fixture, 0600))

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assertConfigEqual(t, cfg)
}

func TestSaveConfig_WritesJSONC(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	require.NoError(t, os.Setenv("HOME", tmpDir))
	defer os.Setenv("HOME", origHome)

	require.NoError(t, SaveConfig(expectedFixtureConfig()))

	content, err := os.ReadFile(filepath.Join(tmpDir, ".config", "bwsf", "config.jsonc"))
	require.NoError(t, err)
	text := string(content)
	assertNoJSONCComments(t, text)
	assert.Contains(t, text, `"type": "bitwarden-cloud"`)
	assert.Contains(t, text, `"email": "user@example.com"`)
}

func TestLoadConfig_JSONC_SaveConfig_RoundTrip(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	require.NoError(t, os.Setenv("HOME", tmpDir))
	defer os.Setenv("HOME", origHome)

	fixture, err := os.ReadFile(filepath.Join("testdata", "comments.jsonc"))
	require.NoError(t, err)

	configDir := filepath.Join(tmpDir, ".config", "bwsf")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	configPath := filepath.Join(configDir, "config.jsonc")
	require.NoError(t, os.WriteFile(configPath, fixture, 0600))

	loaded, err := LoadConfig()
	require.NoError(t, err)
	assertConfigEqual(t, loaded)

	require.NoError(t, SaveConfig(loaded))

	saved, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assertNoJSONCComments(t, string(saved))

	reloaded, err := LoadConfig()
	require.NoError(t, err)
	assertConfigEqual(t, reloaded)
}
