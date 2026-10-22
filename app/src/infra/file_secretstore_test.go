package infra

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSecretStore_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewFileSecretStore(dir)

	require.NoError(t, store.Set("hosts/default/api_client_id", "user.abc"))
	got, err := store.Get("hosts/default/api_client_id")
	require.NoError(t, err)
	assert.Equal(t, "user.abc", got)

	// Host-scoped key becomes a single file (no nested dirs).
	entries, err := filepath.Glob(filepath.Join(dir, "*"))
	require.NoError(t, err)
	assert.Len(t, entries, 1)

	require.NoError(t, store.Delete("hosts/default/api_client_id"))
	_, err = store.Get("hosts/default/api_client_id")
	assert.ErrorIs(t, err, ErrSecretNotFound)
}

func TestFileSecretStore_RejectsTraversal(t *testing.T) {
	store := NewFileSecretStore(t.TempDir())
	err := store.Set("../escape", "x")
	require.Error(t, err)
}

func TestDefaultSecretStore_UsesFileWhenEnvSet(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("BWSF_SECRET_DIR", dir)
	store := DefaultSecretStore()
	_, ok := store.(*FileSecretStore)
	assert.True(t, ok)
	require.NoError(t, store.Set("k", "v"))
	got, err := store.Get("k")
	require.NoError(t, err)
	assert.Equal(t, "v", got)
}

func TestDefaultSecretStore_UsesKeyringWhenEnvUnset(t *testing.T) {
	t.Setenv("BWSF_SECRET_DIR", "")
	store := DefaultSecretStore()
	_, ok := store.(*KeyringStore)
	assert.True(t, ok)
}
