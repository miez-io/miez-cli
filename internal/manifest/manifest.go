// Package manifest owns miez.yaml package and workspace metadata.
package manifest

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/manuel/miez-cli/internal/model"
	"gopkg.in/yaml.v3"
)

// FileName is the manifest's fixed path, at the repository root.
const FileName = "miez.yaml"

// Manifest is either a team package manifest or a workspace manifest. A team
// package uses its metadata fields; a workspace uses Teams to name sources.
type Manifest struct {
	ID           string              `yaml:"id,omitempty"`
	Version      string              `yaml:"version,omitempty"`
	Name         string              `yaml:"name,omitempty"`
	Description  string              `yaml:"description,omitempty"`
	Author       string              `yaml:"author,omitempty"`
	DefaultModel string              `yaml:"default_model,omitempty"`
	Models       []model.ModelOption `yaml:"models,omitempty"`
	MCP          []model.MCPServer   `yaml:"mcp,omitempty"`
	Teams        map[string]string   `yaml:"teams,omitempty"`
}

// Path returns the manifest path below workspaceRoot.
func Path(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, FileName)
}

// Load reads the manifest, returning an empty Manifest when absent.
func Load(workspaceRoot string) (Manifest, error) {
	data, err := os.ReadFile(Path(workspaceRoot))
	if errors.Is(err, os.ErrNotExist) {
		return Manifest{Teams: map[string]string{}}, nil
	}
	if err != nil {
		return Manifest{}, fmt.Errorf("read %s: %w", FileName, err)
	}
	var manifest Manifest
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse %s: %w", FileName, err)
	}
	if manifest.Teams == nil {
		manifest.Teams = map[string]string{}
	}
	return manifest, nil
}

// LoadTeamPackage loads and validates the authored package manifest for a
// team source. It is intentionally separate from Load because a workspace
// manifest may be valid without containing team package metadata.
func LoadTeamPackage(root string) (Manifest, error) {
	manifest, err := Load(root)
	if err != nil {
		return Manifest{}, err
	}
	if !model.IsValidIdentifier(manifest.ID) {
		return Manifest{}, fmt.Errorf("team package id %q is not kebab-case", manifest.ID)
	}
	if manifest.Version == "" {
		return Manifest{}, fmt.Errorf("team package version is required")
	}
	if manifest.Name == "" {
		return Manifest{}, fmt.Errorf("team package name is required")
	}
	return manifest, nil
}

// Save writes the manifest atomically.
func Save(workspaceRoot string, manifest Manifest) error {
	data, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("encode %s: %w", FileName, err)
	}
	return os.WriteFile(Path(workspaceRoot), data, 0o644)
}

// Resolve looks up name in the manifest's teams map. ok is false when name is
// not registered, so callers can distinguish "unknown name" from "is a URL".
func (manifest Manifest) Resolve(name string) (url string, ok bool) {
	url, ok = manifest.Teams[name]
	return url, ok
}

// Register maps name to url, adding or overwriting the entry.
func (manifest *Manifest) Register(name, url string) {
	if manifest.Teams == nil {
		manifest.Teams = map[string]string{}
	}
	manifest.Teams[name] = url
}
