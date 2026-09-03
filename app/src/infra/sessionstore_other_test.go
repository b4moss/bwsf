//go:build !darwin

package infra

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoopSessionStore(t *testing.T) {
	store := NewSessionStore()
	require.NotNil(t, store)

	got, err := store.Get()
	assert.NoError(t, err)
	assert.Equal(t, "", got)

	assert.NoError(t, store.Set("session-value"))
	got, err = store.Get()
	assert.NoError(t, err)
	assert.Equal(t, "", got, "noop store must not persist")

	assert.NoError(t, store.Delete())
}
