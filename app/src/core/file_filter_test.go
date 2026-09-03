package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagedFileFilter_Allow(t *testing.T) {
	assert.True(t, ManagedFileFilter{}.Allow(".env"))

	notSave := ManagedFileFilter{SaveFiles: []string{"!.env.local", "!*.auto.tfvars"}}
	assert.True(t, notSave.Allow(".env"))
	assert.False(t, notSave.Allow(".env.local"))
	assert.False(t, notSave.Allow("x.auto.tfvars"))
	assert.True(t, notSave.Allow("foo.tfvars"))

	save := ManagedFileFilter{SaveFiles: []string{".env", ".env.production"}}
	assert.True(t, save.Allow(".env"))
	assert.False(t, save.Allow(".env.local"))
}

func TestGetPushedEnvFiles_WithFilters(t *testing.T) {
	fs := &mockFileSystem{
		dirEntries: []DirEntry{
			&mockDirEntry{name: ".env"},
			&mockDirEntry{name: ".env.local"},
			&mockDirEntry{name: ".env.example"},
			&mockDirEntry{name: "foo.tfvars"},
			&mockDirEntry{name: "x.auto.tfvars"},
			&mockDirEntry{name: "secrets.yaml"},
		},
	}

	t.Run("none", func(t *testing.T) {
		names, err := GetPushedEnvFiles(".", ManagedFileFilter{}, fs)
		require.NoError(t, err)
		assert.Equal(t, []string{".env", ".env.local", "foo.tfvars", "x.auto.tfvars"}, names)
	})

	t.Run("negation_via_bang", func(t *testing.T) {
		names, err := GetPushedEnvFiles(".", ManagedFileFilter{
			SaveFiles: []string{"!.env.local", "!*.auto.tfvars"},
		}, fs)
		require.NoError(t, err)
		assert.Equal(t, []string{".env", "foo.tfvars"}, names)
	})

	t.Run("save_files", func(t *testing.T) {
		names, err := GetPushedEnvFiles(".", ManagedFileFilter{
			SaveFiles: []string{".env", ".env.production"},
		}, fs)
		require.NoError(t, err)
		assert.Equal(t, []string{".env"}, names)
	})

	t.Run("save_files_outside_base", func(t *testing.T) {
		names, err := GetPushedEnvFiles(".", ManagedFileFilter{
			SaveFiles: []string{"secrets.yaml"},
		}, fs)
		require.NoError(t, err)
		assert.Empty(t, names)
	})
}
