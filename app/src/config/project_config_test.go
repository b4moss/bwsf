package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseProjectConfigJSONC_Variants(t *testing.T) {
	t.Run("not_save_files_rejected", func(t *testing.T) {
		_, err := ParseProjectConfigJSONC([]byte(`{
			"override_project_name": "my-api",
			"not_save_files": [".env.local", "*.auto.tfvars"]
		}`))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not_save_files")
	})

	t.Run("save_files_only", func(t *testing.T) {
		pc, err := ParseProjectConfigJSONC([]byte(`{"save_files":[".env",".env.production"]}`))
		require.NoError(t, err)
		assert.Equal(t, []string{".env", ".env.production"}, pc.EffectiveSaveFiles())
	})

	t.Run("save_files_with_negation", func(t *testing.T) {
		pc, err := ParseProjectConfigJSONC([]byte(`{"save_files":[".env*", "!.env.local"]}`))
		require.NoError(t, err)
		assert.Equal(t, []string{".env*", "!.env.local"}, pc.EffectiveSaveFiles())
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

	t.Run("host_field", func(t *testing.T) {
		pc, err := ParseProjectConfigJSONC([]byte(`{"host":"work"}`))
		require.NoError(t, err)
		assert.Equal(t, "work", pc.EffectiveHost())
	})
}

func TestParseProjectConfigJSONC_BothListsError(t *testing.T) {
	_, err := ParseProjectConfigJSONC([]byte(`{
		"save_files": [".env"],
		"not_save_files": [".env.local"]
	}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not_save_files")
}

func TestParseProjectConfigJSONC_Invalid(t *testing.T) {
	_, err := ParseProjectConfigJSONC([]byte(`{`))
	require.Error(t, err)
}

func TestGetProjectConfigWritePath(t *testing.T) {
	dir := "/tmp/proj"
	assert.Equal(t, filepath.Join(dir, ".bwsf", "config.jsonc"), GetProjectConfigWritePath(dir))
}

func TestSaveProjectConfig_CreatesJSONC(t *testing.T) {
	dir := t.TempDir()
	pc := &ProjectConfig{
		Host:                "work",
		OverrideProjectName: "my-api",
		SaveFiles:           []string{".env*", "!.env.local"},
	}
	require.NoError(t, SaveProjectConfig(dir, pc))

	path := GetProjectConfigWritePath(dir)
	require.FileExists(t, path)
	assert.NoFileExists(t, filepath.Join(dir, ".bwsf", "config.json"))

	loaded, err := LoadProjectConfigFile(path)
	require.NoError(t, err)
	assert.Equal(t, "work", loaded.EffectiveHost())
	assert.Equal(t, "my-api", loaded.EffectiveOverride())
	assert.Equal(t, []string{".env*", "!.env.local"}, loaded.EffectiveSaveFiles())
}

func TestSaveProjectConfig_OmitsEmptyOptionalKeys(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, SaveProjectConfig(dir, &ProjectConfig{}))

	data, err := os.ReadFile(GetProjectConfigWritePath(dir))
	require.NoError(t, err)
	assert.NotContains(t, string(data), "host")
	assert.NotContains(t, string(data), "override_project_name")
	assert.NotContains(t, string(data), "save_files")
}

func TestSaveProjectConfig_RemovesSiblingJSON(t *testing.T) {
	dir := t.TempDir()
	bwsf := filepath.Join(dir, ".bwsf")
	require.NoError(t, os.MkdirAll(bwsf, 0o755))
	jsonPath := filepath.Join(bwsf, "config.json")
	require.NoError(t, os.WriteFile(jsonPath, []byte(`{"host":"old"}`), 0o600))

	require.NoError(t, SaveProjectConfig(dir, &ProjectConfig{Host: "new"}))
	require.FileExists(t, GetProjectConfigWritePath(dir))
	assert.NoFileExists(t, jsonPath)

	loaded, err := LoadProjectConfigFile(GetProjectConfigWritePath(dir))
	require.NoError(t, err)
	assert.Equal(t, "new", loaded.EffectiveHost())
}

func TestSaveProjectConfig_Nil(t *testing.T) {
	err := SaveProjectConfig(t.TempDir(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestFindLocalProjectConfigFiles(t *testing.T) {
	dir := t.TempDir()
	jp, jcp, err := FindLocalProjectConfigFiles(dir)
	require.NoError(t, err)
	assert.Empty(t, jp)
	assert.Empty(t, jcp)

	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".bwsf"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".bwsf", "config.jsonc"), []byte(`{}`), 0o600))
	jp, jcp, err = FindLocalProjectConfigFiles(dir)
	require.NoError(t, err)
	assert.Empty(t, jp)
	assert.NotEmpty(t, jcp)
	assert.True(t, LocalProjectConfigExists(dir))
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
