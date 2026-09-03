package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectConfig is the schema for `.bwsf/config.(json|jsonc)` (#133 / #177).
type ProjectConfig struct {
	OverrideProjectName string   `json:"override_project_name,omitempty"`
	Host                string   `json:"host,omitempty"`
	SaveFiles           []string `json:"save_files,omitempty"`

	// notSaveFilesPresent is set during parse when the banned key exists.
	notSaveFilesPresent bool `json:"-"`
}

// rawProjectConfig is used to detect banned keys.
type rawProjectConfig struct {
	OverrideProjectName string          `json:"override_project_name"`
	Host                string          `json:"host"`
	SaveFiles           []string        `json:"save_files"`
	NotSaveFiles        json.RawMessage `json:"not_save_files"`
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

// EffectiveHost returns a non-empty project host id, or "".
func (p *ProjectConfig) EffectiveHost() string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(p.Host)
}

// EffectiveSaveFiles returns save_files when non-empty after trim.
func (p *ProjectConfig) EffectiveSaveFiles() []string {
	if p == nil {
		return nil
	}
	return nonEmptyStrings(p.SaveFiles)
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

// Validate rejects not_save_files (removed in v0.20).
func (p *ProjectConfig) Validate() error {
	if p == nil {
		return nil
	}
	if p.notSaveFilesPresent {
		return fmt.Errorf("not_save_files is removed; use save_files with ! prefixes")
	}
	return nil
}

// Normalize clears blank override and empty filter entries.
func (p *ProjectConfig) Normalize() {
	if p == nil {
		return
	}
	p.OverrideProjectName = strings.TrimSpace(p.OverrideProjectName)
	p.Host = strings.TrimSpace(p.Host)
	p.SaveFiles = nonEmptyStrings(p.SaveFiles)
}

// ParseProjectConfigJSONC unmarshals and validates project config bytes.
func ParseProjectConfigJSONC(data []byte) (*ProjectConfig, error) {
	var raw rawProjectConfig
	if err := UnmarshalJSONC(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}
	pc := &ProjectConfig{
		OverrideProjectName: raw.OverrideProjectName,
		Host:                raw.Host,
		SaveFiles:           raw.SaveFiles,
		notSaveFilesPresent: len(raw.NotSaveFiles) > 0 && string(raw.NotSaveFiles) != "null",
	}
	pc.Normalize()
	if err := pc.Validate(); err != nil {
		return nil, err
	}
	return pc, nil
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

const (
	projectConfigDir   = ".bwsf"
	projectConfigJSON  = "config.json"
	projectConfigJSONC = "config.jsonc"
)

// FindProjectConfigPaths walks cwd and ancestors for `.bwsf/config.json` or `.jsonc`.
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
