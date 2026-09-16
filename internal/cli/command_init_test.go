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

func TestTeamInstallInstallsAndActivatesGitHubTeam(t *testing.T) {
	root := t.TempDir()
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, root)

	if err := installFixtureTeam(t, app); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".miez", "miez_modules", "spec-driven-team", "miez.generated.yaml")); err != nil {
		t.Fatalf("team not installed under miez_modules: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "prompts", "spec-engineer.prompt.md")); err != nil {
		t.Fatalf("rendered output missing: %v", err)
	}
}

func TestTeamInstallRejectsNonCopilotTarget(t *testing.T) {
	root := t.TempDir()
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, root)
	useFixtureGitHub(t, app, "testdata/spec-driven-team", "spec-driven-team")

	err := app.Execute(context.Background(), []string{"team", "install", "https://github.com/example/spec-driven-team", "--targets", "cursor"})
	if err == nil {
		t.Fatal("team install accepted a non-copilot target")
	}
}

func TestTeamInstallRequiresGitHubURL(t *testing.T) {
	root := t.TempDir()
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, root)

	err := app.Execute(context.Background(), []string{"team", "install", "--targets", "copilot"})
	if err == nil {
		t.Fatal("team install succeeded with no GitHub URL")
	}
}

func TestTeamInstallPromptsForTargetOnFreshWorkspace(t *testing.T) {
	root := t.TempDir()
	output := &bytes.Buffer{}
	app := newTestApp(t, output, &bytes.Buffer{}, root)
	app.In = strings.NewReader("\n")
	useFixtureGitHub(t, app, "testdata/spec-driven-team", "spec-driven-team")

	if err := app.Execute(context.Background(), []string{"team", "install", "https://github.com/example/spec-driven-team"}); err != nil {
		t.Fatal(err)
	}
	workspaceValue, err := workspace.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(workspaceValue.Config.Targets) != 1 || workspaceValue.Config.Targets[0] != "copilot" {
		t.Fatalf("targets = %#v, want [copilot]", workspaceValue.Config.Targets)
	}
	if !strings.Contains(output.String(), "Render target") {
		t.Fatalf("install output did not include target prompt: %q", output.String())
	}
}
