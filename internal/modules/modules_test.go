package modules_test

import (
	"path/filepath"
	"testing"

	"github.com/manuel/miez-cli/internal/modules"
)

func TestTeamPathIsDistinctFromRenderedAndBootstrapRoots(t *testing.T) {
	root := t.TempDir()
	modulePath, err := modules.TeamPath(root, "spec-driven-team")
	if err != nil {
		t.Fatal(err)
	}
	renderedRoot := filepath.Join(root, ".github")
	bootstrapRoot := filepath.Join(root, "spec-driven-team-authoring")

	if modulePath == renderedRoot || filepath.Dir(modulePath) == renderedRoot {
		t.Fatalf("module path %q overlaps rendered output root %q", modulePath, renderedRoot)
	}
	if modulePath == bootstrapRoot {
		t.Fatalf("module path %q overlaps bootstrap authoring root %q", modulePath, bootstrapRoot)
	}
	if filepath.Base(filepath.Dir(modulePath)) != modules.Dir {
		t.Fatalf("module path %q is not under %q", modulePath, modules.Dir)
	}
}

func TestTeamPathRejectsInvalidID(t *testing.T) {
	if _, err := modules.TeamPath(t.TempDir(), "Not_Valid"); err == nil {
		t.Fatal("TeamPath accepted a non-kebab-case id")
	}
}
