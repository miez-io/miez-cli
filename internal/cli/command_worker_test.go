package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/manuel/miez-cli/internal/workspace"
)

func TestRemovingAuthoredSkillRemovesCompiledLink(t *testing.T) {
	root := t.TempDir()
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, root)
	if err := installFixtureTeam(t, app); err != nil {
		t.Fatal(err)
	}
	workerPath := filepath.Join(root, ".github", "agents", "spec-engineer.md")
	data, err := os.ReadFile(workerPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "- [spec-writing](../skills/spec-writing/SKILL.md)") {
		t.Fatalf("skill link missing before removal: %q", data)
	}
	if err := app.Execute(context.Background(), []string{"worker", "skill", "remove", "spec-engineer", "spec-writing"}); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(workerPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "spec-writing") {
		t.Fatal("removed authored skill remains in compiled prompt")
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "skills", "spec-writing", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("orphaned skill file remains after removal: %v", err)
	}
}

func TestRemovingUnassignedSkillDoesNotChangeState(t *testing.T) {
	root := t.TempDir()
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, root)
	if err := installFixtureTeam(t, app); err != nil {
		t.Fatal(err)
	}
	if err := app.Execute(context.Background(), []string{"worker", "skill", "remove", "spec-engineer", "architecture"}); err == nil {
		t.Fatal("removing an unassigned skill succeeded")
	}
	workspaceValue, err := workspace.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	state := workspaceValue.State("spec-driven-team")
	if len(state.RemovedSkills["spec-engineer"]) != 0 {
		t.Fatalf("removed skills = %#v", state.RemovedSkills["spec-engineer"])
	}
}

func TestWorkerModelListShowsSupportedModels(t *testing.T) {
	output := &bytes.Buffer{}
	app := newTestApp(t, output, &bytes.Buffer{}, t.TempDir())
	if err := app.Execute(context.Background(), []string{"worker", "model", "list"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "gpt-5\n") {
		t.Fatalf("output = %q, want it to list gpt-5", output.String())
	}
	if strings.ContainsAny(output.String(), "\t") {
		t.Fatalf("output = %q, want ids only without a display-name column", output.String())
	}
}
