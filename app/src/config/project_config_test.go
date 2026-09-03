package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseProjectConfigJSONC_Variants(t *testing.T) {
	t.Run("plain_override_and_not_save", func(t *testing.T) {
		pc, err := ParseProjectConfigJSONC([]byte(`{
			"override_project_name": "my-api",
			"not_save_files": [".env.local", "*.auto.tfvars"]
		}`))
		require.NoError(t, err)
		assert.Equal(t, "my-api", pc.EffectiveOverride())
		assert.Equal(t, []string{".env.local", "*.auto.tfvars"}, pc.EffectiveNotSaveFiles())
		assert.Empty(t, pc.EffectiveSaveFiles())
	})

	t.Run("save_files_only", func(t *testing.T) {
		pc, err := ParseProjectConfigJSONC([]byte(`{"save_files":[".env",".env.production"]}`))
		require.NoError(t, err)
		assert.Equal(t, []string{".env", ".env.production"}, pc.EffectiveSaveFiles())
		assert.Empty(t, pc.EffectiveNotSaveFiles())
	})

	t.Run("jsonc_comments", func(t *testing.T) {
		pc, err := ParseProjectConfigJSONC([]byte(`{
			// comment
			"override_project_name": "x",
		}`))
		require.NoError(t, err)
		assert.Equal(t, "x", pc.EffectiveOverride())
	})

	t.Run("empty_override_unset", func(t *testing.T) {
		pc, err := ParseProjectConfigJSONC([]byte(`{"override_project_name":"  "}`))
		require.NoError(t, err)
		assert.Empty(t, pc.EffectiveOverride())
	})

	t.Run("empty_save_files_unset", func(t *testing.T) {
		pc, err := ParseProjectConfigJSONC([]byte(`{"save_files":[]}`))
		require.NoError(t, err)
		assert.Empty(t, pc.EffectiveSaveFiles())
	})
}

func TestParseProjectConfigJSONC_BothListsError(t *testing.T) {
	_, err := ParseProjectConfigJSONC([]byte(`{
		"save_files": [".env"],
		"not_save_files": [".env.local"]
	}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save_files")
}

func TestParseProjectConfigJSONC_Invalid(t *testing.T) {
	_, err := ParseProjectConfigJSONC([]byte(`{`))
	require.Error(t, err)
}

func TestFindProjectConfigPaths(t *testing.T) {
	repo := t.TempDir()
	sub := filepath.Join(repo, "app")
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".bwsf"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(sub, ".bwsf"), 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))

	rootCfg := filepath.Join(repo, ".bwsf", "config.json")
	subCfg := filepath.Join(sub, ".bwsf", "config.jsonc")
	require.NoError(t, os.WriteFile(rootCfg, []byte(`{}`), 0o600))
	require.NoError(t, os.WriteFile(subCfg, []byte(`{}`), 0o600))

	paths, err := FindProjectConfigPaths(sub)
	require.NoError(t, err)
	require.Len(t, paths, 2)
	assert.Equal(t, subCfg, paths[0])
	assert.Equal(t, rootCfg, paths[1])
}

func TestFindProjectConfigPaths_BothExtensionsError(t *testing.T) {
	dir := t.TempDir()
	bwsf := filepath.Join(dir, ".bwsf")
	require.NoError(t, os.MkdirAll(bwsf, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(bwsf, "config.json"), []byte(`{}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(bwsf, "config.jsonc"), []byte(`{}`), 0o600))

	_, err := FindProjectConfigPaths(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "both")
}

func TestResolveProjectConfig_ZeroOneMany(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		pc, path, err := ResolveProjectConfig(t.TempDir(), nil)
		require.NoError(t, err)
		assert.Nil(t, pc)
		assert.Empty(t, path)
	})

	t.Run("one", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, ".bwsf"), 0o755))
		cfgPath := filepath.Join(dir, ".bwsf", "config.json")
		require.NoError(t, os.WriteFile(cfgPath, []byte(`{"override_project_name":"solo"}`), 0o600))

		pc, path, err := ResolveProjectConfig(dir, nil)
		require.NoError(t, err)
		assert.Equal(t, cfgPath, path)
		assert.Equal(t, "solo", pc.EffectiveOverride())
	})

	t.Run("many_with_selector", func(t *testing.T) {
		repo := t.TempDir()
		sub := filepath.Join(repo, "app")
		require.NoError(t, os.MkdirAll(filepath.Join(repo, ".bwsf"), 0o755))
		require.NoError(t, os.MkdirAll(filepath.Join(sub, ".bwsf"), 0o755))
		rootCfg := filepath.Join(repo, ".bwsf", "config.json")
		subCfg := filepath.Join(sub, ".bwsf", "config.json")
		require.NoError(t, os.WriteFile(rootCfg, []byte(`{"override_project_name":"root"}`), 0o600))
		require.NoError(t, os.WriteFile(subCfg, []byte(`{"override_project_name":"leaf"}`), 0o600))

		pc, path, err := ResolveProjectConfig(sub, func(paths []string) (string, error) {
			require.Len(t, paths, 2)
			return paths[1], nil // choose root
		})
		require.NoError(t, err)
		assert.Equal(t, rootCfg, path)
		assert.Equal(t, "root", pc.EffectiveOverride())
	})

	t.Run("many_without_selector", func(t *testing.T) {
		repo := t.TempDir()
		sub := filepath.Join(repo, "app")
		require.NoError(t, os.MkdirAll(filepath.Join(repo, ".bwsf"), 0o755))
		require.NoError(t, os.MkdirAll(filepath.Join(sub, ".bwsf"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(repo, ".bwsf", "config.json"), []byte(`{}`), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(sub, ".bwsf", "config.json"), []byte(`{}`), 0o600))

		_, _, err := ResolveProjectConfig(sub, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "multiple")
	})
}
