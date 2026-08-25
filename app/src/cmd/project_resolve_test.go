package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveProjectAndFileDir_WithGitSubdir(t *testing.T) {
	// T-A2 via cmd helper: Name/Dir from git root when flag unset
	repo := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))
	sub := filepath.Join(repo, "app")
	require.NoError(t, os.Mkdir(sub, 0o755))

	orig, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(orig) })
	require.NoError(t, os.Chdir(sub))

	c := &cobra.Command{}
	c.Flags().String("from", ".", "from")

	name, dir, err := resolveProjectAndFileDir(c, "from")
	require.NoError(t, err)
	assert.Equal(t, filepath.Base(repo), name)
	assert.Equal(t, repo, dir)
}

func TestResolveProjectAndFileDir_WithGitFlagOverridesDir(t *testing.T) {
	// T-A3: explicit --from keeps Name at git root basename
	repo := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))
	sub := filepath.Join(repo, "app")
	require.NoError(t, os.Mkdir(sub, 0o755))

	orig, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(orig) })
	require.NoError(t, os.Chdir(sub))

	c := &cobra.Command{}
	c.Flags().String("from", ".", "from")
	require.NoError(t, c.Flags().Set("from", sub))

	name, dir, err := resolveProjectAndFileDir(c, "from")
	require.NoError(t, err)
	assert.Equal(t, filepath.Base(repo), name)
	assert.Equal(t, sub, dir)
}

func TestResolveProjectAndFileDir_WithoutGitFallback(t *testing.T) {
	// T-B1: cwd Name/Dir + fallback (warning is side effect; not asserted here)
	dir := t.TempDir()

	orig, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(orig) })
	require.NoError(t, os.Chdir(dir))

	c := &cobra.Command{}
	c.Flags().String("from", ".", "from")

	name, fileDir, err := resolveProjectAndFileDir(c, "from")
	require.NoError(t, err)
	assert.Equal(t, filepath.Base(dir), name)
	assert.Equal(t, dir, fileDir)
}
