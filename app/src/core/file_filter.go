package core

import (
	"path/filepath"
	"strings"
)

// ManagedFileFilter applies after isManagedFileName (#133 / #177).
// SaveFiles entries are positive globs, or negative when prefixed with "!".
type ManagedFileFilter struct {
	SaveFiles []string
}

// Allow reports whether basename should be kept after the base managed-file rules.
func (f ManagedFileFilter) Allow(basename string) bool {
	base := filepath.Base(basename)
	positives, negatives := splitSaveFilePatterns(f.SaveFiles)

	candidates := true
	if len(positives) > 0 {
		candidates = matchAnyGlob(base, positives)
	}
	if !candidates {
		return false
	}
	if len(negatives) > 0 && matchAnyGlob(base, negatives) {
		return false
	}
	return true
}

func splitSaveFilePatterns(patterns []string) (positives, negatives []string) {
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "!") {
			neg := strings.TrimSpace(strings.TrimPrefix(p, "!"))
			if neg != "" {
				negatives = append(negatives, neg)
			}
			continue
		}
		positives = append(positives, p)
	}
	return positives, negatives
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
