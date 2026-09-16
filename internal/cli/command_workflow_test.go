package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkflowIsAlwaysRendered(t *testing.T) {
	root := t.TempDir()
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, root)
	if err := installFixtureTeam(t, app); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "instructions", "miez-workflow.instructions.md")); err != nil {
		t.Fatalf("workflow instruction missing: %v", err)
	}
}

func TestWorkflowUseCompletionUsesInstalledCatalog(t *testing.T) {
	root := t.TempDir()
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, root)
	if err := installFixtureTeam(t, app); err != nil {
		t.Fatal(err)
	}
	workflowUse, _, err := app.NewCommand().Find([]string{"workflow", "use"})
	if err != nil {
		t.Fatal(err)
	}
	candidates, _ := workflowUse.ValidArgsFunction(workflowUse, []string{}, "")
	if len(candidates) != 1 || candidates[0] != "default" {
		t.Fatalf("workflow candidates = %#v", candidates)
	}
}

func TestWorkflowEnableAndDisableCommandsAreRemoved(t *testing.T) {
	root := t.TempDir()
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, root)
	workflow, _, err := app.NewCommand().Find([]string{"workflow"})
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range workflow.Commands() {
		if command.Name() == "enable" || command.Name() == "disable" {
			t.Fatalf("obsolete workflow command %q still exists", command.Name())
		}
	}
}

func TestWorkflowWorkerDisableRejectsDisablingEveryWorker(t *testing.T) {
	root := t.TempDir()
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, root)
	if err := installFixtureTeam(t, app); err != nil {
		t.Fatal(err)
	}
	for _, worker := range []string{"spec-engineer", "architect", "planner", "user-story", "implementer", "reviewer"} {
		if err := app.Execute(context.Background(), []string{"workflow", "worker", "disable", worker}); err != nil {
			t.Fatal(err)
		}
	}
	if err := app.Execute(context.Background(), []string{"workflow", "worker", "disable", "finisher"}); err == nil {
		t.Fatal("workflow worker disable emptied the active workflow")
	}
}
