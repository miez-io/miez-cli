package lockfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/manuel/miez-cli/internal/lockfile"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	root := t.TempDir()
	entry := lockfile.Entry{
		RepoURL:       "https://github.com/manuel/miez-scrum-team",
		Ref:           "main",
		Commit:        "abc123",
		Version:       "1.0.0",
		DeployedFiles: []string{".github/agents/architect.md"},
		FileHashes:    map[string]string{".github/agents/architect.md": "deadbeef"},
	}
	lock := lockfile.Lockfile{Teams: map[string]lockfile.Entry{"spec-driven-team": entry}}
	if err := lockfile.Save(root, lock); err != nil {
		t.Fatal(err)
	}
	got, err := lockfile.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Teams["spec-driven-team"].Commit != "abc123" {
		t.Fatalf("Commit = %q, want abc123", got.Teams["spec-driven-team"].Commit)
	}
}

func TestBuildEntryHashesEveryDeployedFile(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "workers"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "miez.generated.yaml"), []byte("id: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "workers", "a.md"), []byte("body"), 0o644); err != nil {
		t.Fatal(err)
	}
	entry, err := lockfile.BuildEntry(root, "https://github.com/o/r", "main", "sha1", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(entry.DeployedFiles) != 2 {
		t.Fatalf("DeployedFiles = %#v, want 2 entries", entry.DeployedFiles)
	}
	hash, err := lockfile.HashFile(filepath.Join(root, "workers", "a.md"))
	if err != nil {
		t.Fatal(err)
	}
	if entry.FileHashes["workers/a.md"] != hash {
		t.Fatalf("hash mismatch for workers/a.md")
	}
}
