package core

import (
	"errors"
	"testing"

	"bwsf/src/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockVaultRestorer struct {
	mockBwClient
	restoreOK   bool
	restoreErr  error
	discardErr  error
	restoreCalls int
	discardCalls int
}

func (m *mockVaultRestorer) TryRestoreVaultUnlock() (bool, error) {
	m.restoreCalls++
	m.calls = append(m.calls, "TryRestoreVaultUnlock")
	if m.restoreErr != nil {
		return false, m.restoreErr
	}
	return m.restoreOK, nil
}

func (m *mockVaultRestorer) DiscardVaultUnlock() error {
	m.discardCalls++
	m.calls = append(m.calls, "DiscardVaultUnlock")
	return m.discardErr
}

func TestWithUnlockRetry_VaultUnlockRestoreSkipsPrompt(t *testing.T) {
	bw := &mockVaultRestorer{restoreOK: true}
	logger := &mockLogger{}
	prompted := 0
	err := WithUnlockRetry(bw, &config.Config{}, func() (string, error) {
		prompted++
		return "pw", nil
	}, logger, nil, func() error { return nil })
	require.NoError(t, err)
	assert.Equal(t, 0, prompted)
	assert.Equal(t, 1, bw.restoreCalls)
	assert.Equal(t, 0, bw.discardCalls)
}

func TestWithUnlockRetry_InvalidVaultUnlockDiscardsAndPrompts(t *testing.T) {
	bw := &mockVaultRestorer{restoreOK: true}
	logger := &mockLogger{}
	prompted := 0
	callCount := 0
	err := WithUnlockRetry(bw, &config.Config{}, func() (string, error) {
		prompted++
		return "pw", nil
	}, logger, nil, func() error {
		callCount++
		if callCount == 1 {
			return errors.New("API vault is locked. Enter your master password to unlock decryption keys")
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, prompted)
	assert.Equal(t, 1, bw.discardCalls)
	assert.Contains(t, bw.calls, "Unlock")
}

func TestWithUnlockRetry_NoVaultUnlockPrompts(t *testing.T) {
	bw := &mockVaultRestorer{restoreOK: false}
	logger := &mockLogger{}
	prompted := 0
	callCount := 0
	err := WithUnlockRetry(bw, &config.Config{}, func() (string, error) {
		prompted++
		return "pw", nil
	}, logger, nil, func() error {
		callCount++
		if callCount == 1 {
			return errors.New("API vault is locked")
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, prompted)
	assert.Equal(t, 0, bw.discardCalls)
}

func TestWithUnlockRetry_RestoreErrorDiscards(t *testing.T) {
	bw := &mockVaultRestorer{restoreErr: errors.New("keychain read failed")}
	logger := &mockLogger{}
	callCount := 0
	err := WithUnlockRetry(bw, &config.Config{}, func() (string, error) {
		return "pw", nil
	}, logger, nil, func() error {
		callCount++
		if callCount == 1 {
			return errors.New("API vault is locked")
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, bw.discardCalls)
}
