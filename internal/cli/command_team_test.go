package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/manuel/miez-cli/internal/artifacts"
	"github.com/manuel/miez-cli/internal/validate"
	"github.com/manuel/miez-cli/internal/workspace"
)

func TestBootstrapCreatesTeamThatCheckAccepts(t *testing.T) {
	root := t.TempDir()
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, root)

	if err := app.Execute(context.Background(), []string{"team", "bootstrap", "custom-team"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "custom-team", "miez.generated.yaml")); !os.IsNotExist(err) {
		t.Fatalf("bootstrap generated the team index before build: %v", err)
	}
	for _, relative := range []string{
		"README.md",
		".github/instructions/miez-team-package.instructions.md",
		".github/prompts/new-worker.prompt.md",
		".github/prompts/new-skill.prompt.md",
		".github/prompts/new-workflow.prompt.md",
		".github/skills/write-workers/SKILL.md",
		".github/skills/write-workers/assets/worker-template.md",
		".github/skills/review-team-package/SKILL.md",
		".github/skills/review-team-package/references/build-errors.md",
	} {
		if _, err := os.Stat(filepath.Join(root, "custom-team", filepath.FromSlash(relative))); err != nil {
			t.Fatalf("bootstrap support file %s is missing: %v", relative, err)
		}
	}
	readme, err := os.ReadFile(filepath.Join(root, "custom-team", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(readme), "kind: agent") || !strings.Contains(string(readme), "miez.generated.yaml") {
		t.Fatalf("bootstrap README does not explain the team contract: %q", readme)
	}
	if strings.Contains(string(readme), "{{TEAM_ID}}") || !strings.Contains(string(readme), "custom-team") {
		t.Fatalf("bootstrap README did not resolve the team id: %q", readme)
	}
	if err := app.Execute(context.Background(), []string{"team", "build", "custom-team"}); err != nil {
		t.Fatal(err)
	}
	generated, err := os.ReadFile(filepath.Join(root, "custom-team", "miez.generated.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, unwanted := range []string{".github/", "write-workers", "review-team-package"} {
		if strings.Contains(string(generated), unwanted) {
			t.Fatalf("generated catalog contains authoring support %q: %q", unwanted, generated)
		}
	}
	team, err := artifacts.LoadTeam(filepath.Join(root, "custom-team"))
	if err != nil {
		t.Fatal(err)
	}
	if err := validate.Team(team).Error(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".miez", "miez_modules")); !os.IsNotExist(err) {
		t.Fatalf("bootstrap wrote into miez_modules: %v", err)
	}
}

func TestTeamInstallRetainsAuthoringSupportInModuleAndLockfile(t *testing.T) {
	root := t.TempDir()
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, root)
	remoteRoot := t.TempDir()
	writeSimpleTeam(t, remoteRoot, "support-team")
	for relative, content := range map[string]string{
		"README.md": "team authoring guide\n",
		".github/instructions/miez-team-package.instructions.md": "package contract rules\n",
		".github/prompts/new-worker.prompt.md":                   "new worker helper\n",
		".github/skills/write-workers/assets/worker-template.md": "worker template\n",
	} {
		path := filepath.Join(remoteRoot, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	useFixtureGitHub(t, app, remoteRoot, "support-team")

	if err := app.Execute(context.Background(), []string{"team", "install", "https://github.com/example/support-team", "--targets", "copilot"}); err != nil {
		t.Fatal(err)
	}
	moduleRoot := filepath.Join(root, ".miez", "miez_modules", "support-team")
	for relative, want := range map[string]string{
		"README.md": "team authoring guide\n",
		".github/instructions/miez-team-package.instructions.md": "package contract rules\n",
		".github/prompts/new-worker.prompt.md":                   "new worker helper\n",
		".github/skills/write-workers/assets/worker-template.md": "worker template\n",
	} {
		data, err := os.ReadFile(filepath.Join(moduleRoot, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatalf("installed support file %s is missing: %v", relative, err)
		}
		if string(data) != want {
			t.Fatalf("installed support file %s = %q, want %q", relative, data, want)
		}
	}
	lockfile, err := os.ReadFile(filepath.Join(root, ".miez", "miez.lock.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, relative := range []string{
		"README.md",
		".github/instructions/miez-team-package.instructions.md",
		".github/prompts/new-worker.prompt.md",
		".github/skills/write-workers/assets/worker-template.md",
	} {
		if !strings.Contains(string(lockfile), relative) {
			t.Fatalf("lockfile does not include support file %s: %q", relative, lockfile)
		}
	}
}

func TestTeamInstallRetainsPreviousTeamForLocalUse(t *testing.T) {
	root := t.TempDir()
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, root)
	if err := installFixtureTeam(t, app); err != nil {
		t.Fatal(err)
	}

	remoteRoot := t.TempDir()
	writeSimpleTeam(t, remoteRoot, "remote-team")
	useFixtureGitHub(t, app, remoteRoot, "remote-team")
	if err := app.Execute(context.Background(), []string{"team", "install", "https://github.com/example/remote-team", "--targets", "copilot"}); err != nil {
		t.Fatal(err)
	}

	workspaceValue, err := workspace.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if workspaceValue.Config.ActiveTeam != "remote-team" {
		t.Fatalf("active team = %q", workspaceValue.Config.ActiveTeam)
	}
	if _, err := os.Stat(filepath.Join(root, ".miez", "miez_modules", "spec-driven-team")); err != nil {
		t.Fatalf("old team was not retained for local activation: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".miez", "miez_modules", "remote-team")); err != nil {
		t.Fatalf("new team missing: %v", err)
	}
	if err := app.Execute(context.Background(), []string{"team", "use", "spec-driven-team"}); err != nil {
		t.Fatal(err)
	}
	reopened, err := workspace.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.Config.ActiveTeam != "spec-driven-team" {
		t.Fatalf("active team after local use = %q", reopened.Config.ActiveTeam)
	}
}

func TestTeamUpdateOutdatedAndAudit(t *testing.T) {
	root := t.TempDir()
	output := &bytes.Buffer{}
	app := newTestApp(t, output, &bytes.Buffer{}, root)
	if err := installFixtureTeam(t, app); err != nil {
		t.Fatal(err)
	}

	output.Reset()
	if err := app.Execute(context.Background(), []string{"team", "outdated"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "up to date") {
		t.Fatalf("outdated output = %q", output.String())
	}

	output.Reset()
	if err := app.Execute(context.Background(), []string{"team", "audit"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "clean") {
		t.Fatalf("audit output = %q", output.String())
	}

	output.Reset()
	if err := app.Execute(context.Background(), []string{"team", "update", "--dry-run"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "already up to date") {
		t.Fatalf("update output = %q", output.String())
	}
}

func TestTeamAuditCiFlagExitsNonZeroOnDrift(t *testing.T) {
	root := t.TempDir()
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, root)
	if err := installFixtureTeam(t, app); err != nil {
		t.Fatal(err)
	}
	tamperedPath := filepath.Join(root, ".miez", "miez_modules", "spec-driven-team", "miez.generated.yaml")
	if err := os.WriteFile(tamperedPath, []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := app.Execute(context.Background(), []string{"team", "audit", "--ci"})
	if err == nil {
		t.Fatal("team audit --ci succeeded despite drift")
	}
	var exitError *ExitError
	if !errors.As(err, &exitError) || exitError.Code == 0 {
		t.Fatalf("error = %v, want a non-zero ExitError", err)
	}
}

func writeSimpleTeam(t *testing.T, root, id string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "workers"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := "id: " + id + "\nversion: 1.0.0\nname: " + id + "\nauthor: test\nworkers:\n  - id: starter\n    kind: command\n    path: workers/starter.md\nworkflows:\n  - id: default\n    name: Default\n    path: workflows/default.md\n    phases:\n      - id: work\n        workers: [starter]\n"
	if err := os.WriteFile(filepath.Join(root, "miez.generated.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "miez.yaml"), []byte("id: "+id+"\nversion: 1.0.0\nname: "+id+"\nauthor: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "workers", "starter.md"), []byte("---\nid: starter\n---\n# Starter\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "workflows"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "workflows", "default.md"), []byte("---\nid: default\nname: Default\nphases:\n  - id: work\n    workers: [starter]\n---\n# Default\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
