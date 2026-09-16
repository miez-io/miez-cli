package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/manuel/miez-cli/internal/model"
)

func TestInitializeCreatesV3StateDirectories(t *testing.T) {
	root := t.TempDir()
	workspace, err := Initialize(root, []string{"copilot"}, "")
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		workspace.ConfigPath(),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("stat %s: %v", path, err)
		}
	}
	for _, path := range []string{
		filepath.Join(root, ".miez", "changes"),
		filepath.Join(root, ".miez", "runs"),
		filepath.Join(root, ".miez", "manifest.json"),
		filepath.Join(root, ".miez", "compiled.json"),
		filepath.Join(root, "miez_modules"),
		filepath.Join(root, "miez.lock.yaml"),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("unexpected initialized path %s: %v", path, err)
		}
	}
	ignoreData, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range []string{".miez/mcp.local.yaml", ".miez/miez_modules/"} {
		if !strings.Contains(string(ignoreData), entry) {
			t.Errorf(".gitignore = %q, want it to contain %q", ignoreData, entry)
		}
	}
	if got := workspace.Config.Targets; len(got) != 1 || got[0] != "copilot" {
		t.Fatalf("targets = %#v, want [copilot]", got)
	}
}

func TestLegacyManifestRejectsAbsoluteAndParentPaths(t *testing.T) {
	workspace := Workspace{Root: t.TempDir()}
	manifestPath := filepath.Join(workspace.Root, MiezDir, "manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, []byte(`{"files":["../outside"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.LegacyManagedFiles(); err == nil {
		t.Fatal("LegacyManagedFiles accepted an unsafe path")
	}
}

func TestOpenRejectsFutureConfigVersion(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, ".miez", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte("version: 999\ntargets: [copilot]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(root); err == nil {
		t.Fatal("Open accepted a future config version")
	}
}

func TestEnsureTeamStateReturnsAnIndependentCopy(t *testing.T) {
	workspace := Workspace{
		Config: model.Config{
			TeamStates: map[string]model.TeamState{},
		},
	}
	team := model.Team{ID: "demo-team"}
	state := workspace.EnsureTeamState(team)
	state.SetWorkerSkill("builder", "writing", true)
	state.SetWorkerModel("builder", "gpt-5")

	stored := workspace.State(team.ID)
	if len(stored.WorkerSkills["builder"]) != 0 {
		t.Fatalf("stored worker skills = %#v, want unchanged", stored.WorkerSkills)
	}
	if len(stored.WorkerModels) != 0 {
		t.Fatalf("stored worker models = %#v, want unchanged", stored.WorkerModels)
	}
}

func TestPreparationRollbackRestoresRepositoryState(t *testing.T) {
	root := t.TempDir()
	ignorePath := filepath.Join(root, ".gitignore")
	if err := os.WriteFile(ignorePath, []byte("custom\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	workspace, err := New(root, []string{"copilot"}, "")
	if err != nil {
		t.Fatal(err)
	}
	preparation, err := workspace.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	if err := preparation.Rollback(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".miez")); !os.IsNotExist(err) {
		t.Fatalf("state directory after rollback: %v", err)
	}
	data, err := os.ReadFile(ignorePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "custom\n" {
		t.Fatalf(".gitignore after rollback = %q", data)
	}
}
