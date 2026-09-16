package artifacts

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/manuel/miez-cli/internal/model"
)

func TestReadWorkerRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("---\n---\n# Outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "worker.md")); err != nil {
		t.Fatal(err)
	}

	team := TeamDir{Root: root}
	if _, err := team.ReadWorker(model.Worker{ID: "worker", Path: "worker.md"}); err == nil {
		t.Fatal("ReadWorker accepted a symlink")
	}
}

func TestReadWorkerRejectsParentPath(t *testing.T) {
	team := TeamDir{Root: t.TempDir()}
	if _, err := team.ReadWorker(model.Worker{ID: "worker", Path: "../outside.md"}); err == nil {
		t.Fatal("ReadWorker accepted a parent path")
	}
}

func TestLoadTeamRejectsUnknownManifestField(t *testing.T) {
	root := t.TempDir()
	manifest := []byte("id: example-team\nversion: 1.0.0\nname: Example\nunexpected: true\n")
	if err := os.WriteFile(filepath.Join(root, TeamIndexFileName), manifest, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadTeam(root); err == nil {
		t.Fatal("LoadTeam accepted an unknown manifest field")
	}
}

func TestLoadTeamRejectsModelDisplayName(t *testing.T) {
	root := t.TempDir()
	manifest := []byte("id: example-team\nversion: 1.0.0\nname: Example\nmodels:\n  - id: gpt-6\n    name: GPT-6\n    copilot: GPT-6 (copilot)\n")
	if err := os.WriteFile(filepath.Join(root, TeamIndexFileName), manifest, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadTeam(root); err == nil {
		t.Fatal("LoadTeam accepted a model display name")
	}
}

func TestCopyDirectoryRejectsSymlink(t *testing.T) {
	source := t.TempDir()
	if err := os.Symlink(source, filepath.Join(source, "self")); err != nil {
		t.Fatal(err)
	}
	if err := CopyDirectory(t.Context(), source, t.TempDir()); err == nil {
		t.Fatal("CopyDirectory accepted a symlink")
	}
}
