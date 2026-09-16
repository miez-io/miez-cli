// Package modules locates raw, downloaded team source on disk, kept distinct
// from rendered target output and from team.Bootstrap's not-yet-installed
// authoring output.
package modules

import (
	"fmt"
	"path/filepath"

	"github.com/manuel/miez-cli/internal/model"
)

// Dir is the folder below .miez that holds raw downloaded team source.
const Dir = "miez_modules"

// Root returns the modules directory below the workspace's .miez directory.
func Root(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, ".miez", Dir)
}

// TeamPath returns the raw source directory for teamID.
func TeamPath(workspaceRoot, teamID string) (string, error) {
	if !model.IsValidIdentifier(teamID) {
		return "", fmt.Errorf("team id %q must be kebab-case", teamID)
	}
	return filepath.Join(Root(workspaceRoot), teamID), nil
}
