package core

import (
	"path/filepath"
)

// ManagedFileFilter applies after isManagedFileName (#133).
// At most one of SaveFiles / NotSaveFiles should be non-empty (validated at config load).
type ManagedFileFilter struct {
	SaveFiles    []string
	NotSaveFiles []string
}

// Allow reports whether basename should be kept after the base managed-file rules.
func (f ManagedFileFilter) Allow(basename string) bool {
	base := filepath.Base(basename)
	if len(f.SaveFiles) > 0 {
		return matchAnyGlob(base, f.SaveFiles)
	}
	if len(f.NotSaveFiles) > 0 {
		return !matchAnyGlob(base, f.NotSaveFiles)
	}
	return true
}

func matchAnyGlob(name string, patterns []string) bool {
	for _, pattern := range patterns {
		ok, err := filepath.Match(pattern, name)
		if err == nil && ok {
			return true
		}
	}
	return false
}
