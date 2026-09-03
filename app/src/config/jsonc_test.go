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
	return &Config{
		HostType:      "cloud",
		SelfhostedURL: "",
		Email:         "user@example.com",
	}
}

func assertConfigEqual(t *testing.T, got *Config) {
	t.Helper()
	want := expectedFixtureConfig()
	require.NotNil(t, got)
	assert.Equal(t, want.HostType, got.HostType)
	assert.Equal(t, want.SelfhostedURL, got.SelfhostedURL)
	assert.Equal(t, want.Email, got.Email)
	assert.Equal(t, want.FolderName, got.FolderName)
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
		err := UnmarshalConfigJSONC([]byte(`{"host_type":`), &cfg)
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
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.json"), fixture, 0600))

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assertConfigEqual(t, cfg)
}

func TestSaveConfig_StrictJSON(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	require.NoError(t, os.Setenv("HOME", tmpDir))
	defer os.Setenv("HOME", origHome)

	require.NoError(t, SaveConfig(expectedFixtureConfig()))

	content, err := os.ReadFile(filepath.Join(tmpDir, ".config", "bwsf", "config.json"))
	require.NoError(t, err)
	text := string(content)
	assert.NotContains(t, text, "//")
	assert.NotContains(t, text, "/*")
	assert.False(t, strings.Contains(text, ",\n}"), "trailing comma before closing brace")
	assert.Contains(t, text, `"host_type": "cloud"`)
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
	configPath := filepath.Join(configDir, "config.json")
	require.NoError(t, os.WriteFile(configPath, fixture, 0600))

	loaded, err := LoadConfig()
	require.NoError(t, err)
	assertConfigEqual(t, loaded)

	require.NoError(t, SaveConfig(loaded))

	saved, err := os.ReadFile(configPath)
	require.NoError(t, err)
	text := string(saved)
	assert.NotContains(t, text, "//")
	assert.NotContains(t, text, "/*")
	assert.False(t, strings.Contains(text, ",\n}"))

	reloaded, err := LoadConfig()
	require.NoError(t, err)
	assertConfigEqual(t, reloaded)
}
