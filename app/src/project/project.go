package project

import (
	"os"
	"path/filepath"
)

// Context holds resolved project name and default file directory.
type Context struct {
	GitRoot         string
	ProjectName     string
	WorkDir         string
	UsedCwdFallback bool
}

// FindGitRoot walks ancestors of start looking for a .git entry (directory or file).
// It does not invoke the git CLI. ok is false when no .git is found.
func FindGitRoot(start string) (root string, ok bool) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}

	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return dir, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// Resolve derives ProjectName and WorkDir from cwd.
// overrideProjectName, when non-empty, wins for ProjectName (#133 reserved slot).
func Resolve(cwd, overrideProjectName string) (Context, error) {
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return Context{}, err
	}

	ctx := Context{}
	if root, found := FindGitRoot(abs); found {
		ctx.GitRoot = root
		ctx.WorkDir = root
		if overrideProjectName != "" {
			ctx.ProjectName = overrideProjectName
		} else {
			ctx.ProjectName = filepath.Base(root)
		}
		return ctx, nil
	}

	ctx.WorkDir = abs
	ctx.UsedCwdFallback = true
	if overrideProjectName != "" {
		ctx.ProjectName = overrideProjectName
	} else {
		ctx.ProjectName = filepath.Base(abs)
	}
	return ctx, nil
}

// SelectFileDir returns the explicit flag value when the flag was changed;
// otherwise it returns the default workDir from Resolve.
func SelectFileDir(flagChanged bool, flagValue, workDir string) string {
	if flagChanged {
		return flagValue
	}
	return workDir
}
