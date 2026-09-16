package team

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/manuel/miez-cli/internal/model"
)

// bootstrapTemplates is embedded so every miez binary can create the same
// authoring support bundle without depending on files beside the executable.
//
//go:embed all:bootstrap_templates
var bootstrapTemplates embed.FS

type bootstrapFile struct {
	path     string
	contents []byte
}

// Bootstrap creates a starter team authoring skeleton at Root/teamID. It
// always stays outside .miez/miez_modules/, since a bootstrapped team is not yet
// installed; the user pushes it to its own GitHub repository and installs
// it explicitly with `miez team install <github-url>` when ready.
func (s *Service) Bootstrap(teamID string) error {
	return BootstrapAt(filepath.Join(s.Root, teamID), teamID)
}

// BootstrapAt creates a starter team authoring skeleton for teamID directly
// at destination, independent of any workspace.
func BootstrapAt(destination, teamID string) error {
	if !model.IsValidIdentifier(teamID) {
		return fmt.Errorf("team id %q must be kebab-case", teamID)
	}
	if _, err := os.Stat(filepath.Join(destination, "miez.yaml")); err == nil {
		return fmt.Errorf("team already exists at %s", destination)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	files, err := loadBootstrapFiles(teamID)
	if err != nil {
		return err
	}
	for _, file := range files {
		path := filepath.Join(destination, filepath.FromSlash(file.path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, file.contents, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func loadBootstrapFiles(teamID string) ([]bootstrapFile, error) {
	files := []bootstrapFile{}
	err := fs.WalkDir(bootstrapTemplates, "bootstrap_templates", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		contents, err := bootstrapTemplates.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read bootstrap template %s: %w", path, err)
		}
		relative := strings.TrimPrefix(path, "bootstrap_templates/")
		contents = []byte(strings.ReplaceAll(string(contents), "{{TEAM_ID}}", teamID))
		files = append(files, bootstrapFile{path: relative, contents: contents})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("load bootstrap templates: %w", err)
	}
	sort.Slice(files, func(left, right int) bool { return files[left].path < files[right].path })
	return files, nil
}
