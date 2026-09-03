package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	projectConfigDir      = ".bwsf"
	projectConfigJSON     = "config.json"
	projectConfigJSONC    = "config.jsonc"
)

// ProjectConfig is the schema for `.bwsf/config.(json|jsonc)` (#133).
type ProjectConfig struct {
	OverrideProjectName string   `json:"override_project_name,omitempty"`
	SaveFiles           []string `json:"save_files,omitempty"`
	NotSaveFiles        []string `json:"not_save_files,omitempty"`
}

// PathSelector chooses one config path when multiple candidates exist.
type PathSelector func(paths []string) (string, error)

// EffectiveOverride returns a non-empty override_project_name, or "".
func (p *ProjectConfig) EffectiveOverride() string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(p.OverrideProjectName)
}

// EffectiveSaveFiles returns save_files when that mode is active (non-empty after trim).
func (p *ProjectConfig) EffectiveSaveFiles() []string {
	if p == nil {
		return nil
	}
	return nonEmptyStrings(p.SaveFiles)
}

// EffectiveNotSaveFiles returns not_save_files when that mode is active.
func (p *ProjectConfig) EffectiveNotSaveFiles() []string {
	if p == nil {
		return nil
	}
	return nonEmptyStrings(p.NotSaveFiles)
}

func nonEmptyStrings(in []string) []string {
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// Validate ensures save_files and not_save_files are not both set.
func (p *ProjectConfig) Validate() error {
	if p == nil {
		return nil
	}
	save := nonEmptyStrings(p.SaveFiles)
	notSave := nonEmptyStrings(p.NotSaveFiles)
	if len(save) > 0 && len(notSave) > 0 {
		return fmt.Errorf("project config: set only one of save_files or not_save_files")
	}
	return nil
}

// Normalize clears blank override and empty filter entries for consistent use.
func (p *ProjectConfig) Normalize() {
	if p == nil {
		return
	}
	p.OverrideProjectName = strings.TrimSpace(p.OverrideProjectName)
	p.SaveFiles = nonEmptyStrings(p.SaveFiles)
	p.NotSaveFiles = nonEmptyStrings(p.NotSaveFiles)
}

// ParseProjectConfigJSONC unmarshals and validates project config bytes.
func ParseProjectConfigJSONC(data []byte) (*ProjectConfig, error) {
	var pc ProjectConfig
	if err := UnmarshalJSONC(data, &pc); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}
	pc.Normalize()
	if err := pc.Validate(); err != nil {
		return nil, err
	}
	return &pc, nil
}

// LoadProjectConfigFile reads and parses a `.bwsf/config.(json|jsonc)` file.
func LoadProjectConfigFile(path string) (*ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config: %w", err)
	}
	return ParseProjectConfigJSONC(data)
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

// FindProjectConfigPaths walks cwd and ancestors for `.bwsf/config.json` or `.jsonc`.
// Candidates are ordered from cwd toward the filesystem root.
// Both config.json and config.jsonc in the same .bwsf directory is an error.
func FindProjectConfigPaths(cwd string) ([]string, error) {
	dir, err := filepath.Abs(cwd)
	if err != nil {
		return nil, err
	}

	var paths []string
	for {
		bwsfDir := filepath.Join(dir, projectConfigDir)
		jsonPath := filepath.Join(bwsfDir, projectConfigJSON)
		jsoncPath := filepath.Join(bwsfDir, projectConfigJSONC)
		hasJSON := fileExists(jsonPath)
		hasJSONC := fileExists(jsoncPath)
		if hasJSON && hasJSONC {
			return nil, fmt.Errorf("both %s and %s exist; keep only one", jsonPath, jsoncPath)
		}
		if hasJSON {
			paths = append(paths, jsonPath)
		} else if hasJSONC {
			paths = append(paths, jsoncPath)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return paths, nil
}

// ResolveProjectConfig finds and loads project config for cwd.
// Returns (nil, "", nil) when no config file exists.
func ResolveProjectConfig(cwd string, selectPath PathSelector) (*ProjectConfig, string, error) {
	paths, err := FindProjectConfigPaths(cwd)
	if err != nil {
		return nil, "", err
	}
	switch len(paths) {
	case 0:
		return nil, "", nil
	case 1:
		pc, err := LoadProjectConfigFile(paths[0])
		if err != nil {
			return nil, "", err
		}
		return pc, paths[0], nil
	default:
		if selectPath == nil {
			return nil, "", fmt.Errorf("multiple project configs found; interactive selection required: %s", strings.Join(paths, ", "))
		}
		chosen, err := selectPath(paths)
		if err != nil {
			return nil, "", err
		}
		chosen = strings.TrimSpace(chosen)
		if chosen == "" {
			return nil, "", fmt.Errorf("no project config path selected")
		}
		pc, err := LoadProjectConfigFile(chosen)
		if err != nil {
			return nil, "", err
		}
		return pc, chosen, nil
	}
}
