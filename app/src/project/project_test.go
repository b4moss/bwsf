package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindGitRoot_Directory(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))
	sub := filepath.Join(repo, "app")
	require.NoError(t, os.Mkdir(sub, 0o755))

	root, ok := FindGitRoot(sub)
	require.True(t, ok)
	assert.Equal(t, repo, root)

	root, ok = FindGitRoot(repo)
	require.True(t, ok)
	assert.Equal(t, repo, root)
}

func TestFindGitRoot_GitFile(t *testing.T) {
	// T-W: .git as a file (worktree-style) behaves like a directory entry
	repo := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".git"), []byte("gitdir: /tmp/somewhere\n"), 0o644))
	sub := filepath.Join(repo, "pkg")
	require.NoError(t, os.Mkdir(sub, 0o755))

	root, ok := FindGitRoot(sub)
	require.True(t, ok)
	assert.Equal(t, repo, root)
}

func TestFindGitRoot_NotFound(t *testing.T) {
	dir := t.TempDir()
	root, ok := FindGitRoot(dir)
	assert.False(t, ok)
	assert.Empty(t, root)
}

func TestResolve_WithGit_RootCwd(t *testing.T) {
	// T-A1
	repo := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))

	ctx, err := Resolve(repo, "")
	require.NoError(t, err)
	assert.Equal(t, repo, ctx.GitRoot)
	assert.Equal(t, filepath.Base(repo), ctx.ProjectName)
	assert.Equal(t, repo, ctx.WorkDir)
	assert.False(t, ctx.UsedCwdFallback)
}

func TestResolve_WithGit_SubdirCwd(t *testing.T) {
	// T-A2
	repo := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))
	sub := filepath.Join(repo, "app")
	require.NoError(t, os.Mkdir(sub, 0o755))

	ctx, err := Resolve(sub, "")
	require.NoError(t, err)
	assert.Equal(t, repo, ctx.GitRoot)
	assert.Equal(t, filepath.Base(repo), ctx.ProjectName)
	assert.Equal(t, repo, ctx.WorkDir)
	assert.False(t, ctx.UsedCwdFallback)
	assert.NotEqual(t, filepath.Base(sub), ctx.ProjectName)
}

func TestResolve_WithGit_OverrideName(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))
	sub := filepath.Join(repo, "app")
	require.NoError(t, os.Mkdir(sub, 0o755))

	ctx, err := Resolve(sub, "my-api")
	require.NoError(t, err)
	assert.Equal(t, "my-api", ctx.ProjectName)
	assert.Equal(t, repo, ctx.WorkDir)
	assert.False(t, ctx.UsedCwdFallback)
}

func TestResolve_WithoutGit_Fallback(t *testing.T) {
	// T-B1 / T-N
	dir := t.TempDir()

	ctx, err := Resolve(dir, "")
	require.NoError(t, err)
	assert.Empty(t, ctx.GitRoot)
	assert.Equal(t, filepath.Base(dir), ctx.ProjectName)
	assert.Equal(t, dir, ctx.WorkDir)
	assert.True(t, ctx.UsedCwdFallback)
}

func TestSelectFileDir(t *testing.T) {
	// T-A3 / T-B2: flag overrides Dir only
	workDir := "/repo"
	assert.Equal(t, workDir, SelectFileDir(false, ".", workDir))
	assert.Equal(t, "/repo/app", SelectFileDir(true, "/repo/app", workDir))
	assert.Equal(t, "/tmp/other", SelectFileDir(true, "/tmp/other", workDir))
}
