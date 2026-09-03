package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withStdin(t *testing.T, input string) {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)
	_, err = w.WriteString(input)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	orig := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = orig
		_ = r.Close()
	})
}

func TestInputURL_Success(t *testing.T) {
	withStdin(t, "https://vault.example.com\n")
	got, err := InputURL()
	require.NoError(t, err)
	assert.Equal(t, "https://vault.example.com", got)
}

func TestInputURL_Empty(t *testing.T) {
	withStdin(t, "\n")
	_, err := InputURL()
	require.Error(t, err)
}

func TestInputEmail_Success(t *testing.T) {
	withStdin(t, "user@example.com\n")
	got, err := InputEmail()
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", got)
}

func TestInputPassword_FromEnv(t *testing.T) {
	t.Setenv("BWSF_PASSWORD", "secret-from-env")
	got, err := InputPassword()
	require.NoError(t, err)
	assert.Equal(t, "secret-from-env", got)
}

func TestConfirmOverwrite_Yes(t *testing.T) {
	withStdin(t, "y\n")
	ok, err := ConfirmOverwrite("overwrite? ")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestConfirmOverwrite_No(t *testing.T) {
	withStdin(t, "n\n")
	ok, err := ConfirmOverwrite("overwrite? ")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestConfirmYesNo_Yes(t *testing.T) {
	withStdin(t, "yes\n")
	ok, err := ConfirmYesNo("create? ")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestSelectProjectConfigPath_Empty(t *testing.T) {
	_, err := SelectProjectConfigPath(nil)
	require.Error(t, err)
}

func TestSelectProjectConfigPath_Single(t *testing.T) {
	got, err := SelectProjectConfigPath([]string{"/a/.bwsf/config.json"})
	require.NoError(t, err)
	assert.Equal(t, "/a/.bwsf/config.json", got)
}

func TestSelectProjectConfigPath_MultipleNonTTY(t *testing.T) {
	// stdin is typically non-TTY in tests
	_, err := SelectProjectConfigPath([]string{"/a/config.json", "/b/config.json"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-TTY")
}

func TestInputEmail_Empty(t *testing.T) {
	withStdin(t, "   \n")
	_, err := InputEmail()
	require.Error(t, err)
}

func TestCleanMismatchConstants(t *testing.T) {
	assert.Equal(t, "abort", CleanMismatchAbort)
	assert.Equal(t, "overwrite_remote_then_clean", CleanMismatchOverwriteRemoteThenClean)
	assert.Equal(t, "remove_local", CleanMismatchRemoveLocal)
}
