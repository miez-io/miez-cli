package team

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/manuel/miez-cli/internal/workspace"
)

func initWorkspace(t *testing.T, root string) workspace.Workspace {
	t.Helper()
	workspaceValue, err := workspace.Initialize(root, []string{"copilot"}, "")
	if err != nil {
		t.Fatal(err)
	}
	return workspaceValue
}

func TestInstallInstallsPublicRepoWithNoCredential(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "team-a", "main", "sha1", demoTeamFiles("team-a", "Team A"))
	root := t.TempDir()
	service := newTestService(root, fake, map[string]string{})

	teamID, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/team-a")
	if err != nil {
		t.Fatal(err)
	}
	if teamID != "team-a" {
		t.Fatalf("team id = %q", teamID)
	}
	if _, err := os.Stat(filepath.Join(root, ".miez", "miez_modules", "team-a", "miez.generated.yaml")); err != nil {
		t.Fatalf("team not installed under miez_modules: %v", err)
	}
	workspaceValue, err := workspace.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if workspaceValue.Config.ActiveTeam != "team-a" {
		t.Fatalf("active team = %q, want team-a", workspaceValue.Config.ActiveTeam)
	}
	if _, err := os.Stat(filepath.Join(root, ".miez", "miez.lock.yaml")); err != nil {
		t.Fatalf("lockfile not written: %v", err)
	}
}

func TestInstallAllowsTeamWithoutWorkflow(t *testing.T) {
	fake := newFakeGitHub(t)
	files := demoTeamFiles("workflowless-team", "Workflowless Team")
	delete(files, "workflowless-team/workflows/default.md")
	files["workflowless-team/miez.generated.yaml"] = strings.Replace(
		files["workflowless-team/miez.generated.yaml"],
		"workflows:\n  - id: default\n    name: Default\n    path: workflows/default.md\n    phases:\n      - id: build\n        workers: [builder]\n",
		"workflows: []\n",
		1,
	)
	fake.setTeam("acme", "workflowless-team", "main", "sha1", files)
	root := t.TempDir()
	service := newTestService(root, fake, map[string]string{})

	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/workflowless-team"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "agents", "builder.md")); err != nil {
		t.Fatalf("worker output is missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "instructions", "miez-workflow.instructions.md")); !os.IsNotExist(err) {
		t.Fatalf("workflow output exists for workflowless team: %v", err)
	}
}

func TestInstallRendersAgentDescriptionFromWorkerFrontmatter(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "agent-team", "main", "sha1", agentTeamFiles())
	root := t.TempDir()
	service := newTestService(root, fake, map[string]string{})

	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/agent-team"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".github", "agents", "architect.md"))
	if err != nil {
		t.Fatal(err)
	}
	if want := "description: Designs system architecture.\n"; !strings.Contains(string(data), want) {
		t.Fatalf("installed agent = %q, want %q", data, want)
	}
}

func TestInstallMigratesLegacyDistributionState(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "team-a", "main", "sha1", demoTeamFiles("team-a", "Team A"))
	root := t.TempDir()
	service := newTestService(root, fake, map[string]string{})
	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/team-a"); err != nil {
		t.Fatal(err)
	}

	if err := os.Rename(filepath.Join(root, ".miez", "miez_modules"), filepath.Join(root, "miez_modules")); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(root, ".miez", "miez.lock.yaml"), filepath.Join(root, "miez.lock.yaml")); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		filepath.Join(root, ".miez", "changes", "archive"),
		filepath.Join(root, ".miez", "runs"),
	} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".miez", "manifest.json"), []byte(`{"files":[".github/prompts/builder.prompt.md"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".miez", "compiled.json"), []byte(`{"active_team":"team-a"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/team-a"); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		filepath.Join(root, ".miez", "miez_modules", "team-a"),
		filepath.Join(root, ".miez", "miez.lock.yaml"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("migrated path %s is missing: %v", path, err)
		}
	}
	for _, path := range []string{
		filepath.Join(root, "miez_modules"),
		filepath.Join(root, "miez.lock.yaml"),
		filepath.Join(root, ".miez", "manifest.json"),
		filepath.Join(root, ".miez", "compiled.json"),
		filepath.Join(root, ".miez", "changes"),
		filepath.Join(root, ".miez", "runs"),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("legacy path %s remains: %v", path, err)
		}
	}
}

func TestInstallFailsClearlyForPrivateRepoWithNoCredential(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.requireToken = "secret"
	fake.setTeam("acme", "private-team", "main", "sha1", demoTeamFiles("private-team", "Private Team"))
	root := t.TempDir()
	initWorkspace(t, root)
	service := newTestService(root, fake, map[string]string{})

	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/private-team"); err == nil {
		t.Fatal("Use succeeded against a private repo with no credential")
	}
	if _, err := os.Stat(filepath.Join(root, ".miez", "miez_modules")); !os.IsNotExist(err) {
		t.Fatalf("miez_modules created despite failure: %v", err)
	}
	reopened, err := workspace.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.Config.ActiveTeam != "" {
		t.Fatalf("active team = %q, want unchanged (empty)", reopened.Config.ActiveTeam)
	}
}

func TestInstallPrivateRepoWithCredentialSucceeds(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.requireToken = "secret"
	fake.setTeam("acme", "private-team", "main", "sha1", demoTeamFiles("private-team", "Private Team"))
	root := t.TempDir()
	initWorkspace(t, root)
	service := newTestService(root, fake, map[string]string{"GITHUB_MIEZ_PAT": "secret"})

	teamID, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/private-team")
	if err != nil {
		t.Fatal(err)
	}
	if teamID != "private-team" {
		t.Fatalf("team id = %q", teamID)
	}
}

func TestUseUnknownTeamIDFailsWithoutRemoteLookup(t *testing.T) {
	root := t.TempDir()
	workspaceValue := initWorkspace(t, root)
	service := newTestService(root, newFakeGitHub(t), map[string]string{})

	if _, err := service.Use(context.Background(), workspaceValue, "not-installed"); err == nil {
		t.Fatal("Use succeeded for an uninstalled team id")
	}
}

func TestUseActivatesInstalledTeamAndRetainsOtherModules(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "team-a", "main", "sha1", demoTeamFiles("team-a", "Team A"))
	fake.setTeam("acme", "team-b", "main", "sha1", demoTeamFiles("team-b", "Team B"))
	root := t.TempDir()
	workspaceValue := initWorkspace(t, root)
	service := newTestService(root, fake, map[string]string{})

	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/team-a"); err != nil {
		t.Fatal(err)
	}
	workspaceValue, err := workspace.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Install(context.Background(), nil, "https://github.com/acme/team-b"); err != nil {
		t.Fatal(err)
	}
	workspaceValue, err = workspace.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Use(context.Background(), workspaceValue, "team-a"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(root, ".github", "agents", "builder.md")); err != nil {
		t.Fatalf("new team's rendered output missing: %v", err)
	}
	reopened, err := workspace.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.Config.ActiveTeam != "team-a" {
		t.Fatalf("active team = %q, want team-a", reopened.Config.ActiveTeam)
	}
	for _, teamID := range []string{"team-a", "team-b"} {
		if _, err := os.Stat(filepath.Join(root, ".miez", "miez_modules", teamID)); err != nil {
			t.Fatalf("installed team %s missing after local activation: %v", teamID, err)
		}
	}
}

func TestUseRejectsGitHubURL(t *testing.T) {
	root := t.TempDir()
	workspaceValue := initWorkspace(t, root)
	service := newTestService(root, newFakeGitHub(t), map[string]string{})

	if _, err := service.Use(context.Background(), workspaceValue, "https://github.com/acme/team-a"); err == nil {
		t.Fatal("Use accepted a GitHub URL")
	}
}

func TestInstallRejectsTargetChangeInExistingWorkspace(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "team-a", "main", "sha1", demoTeamFiles("team-a", "Team A"))
	root := t.TempDir()
	initWorkspace(t, root)
	service := newTestService(root, fake, map[string]string{})

	if _, err := service.Install(context.Background(), []string{"cursor"}, "https://github.com/acme/team-a"); err == nil {
		t.Fatal("Install accepted a target change in an existing workspace")
	}
	if _, err := os.Stat(filepath.Join(root, ".miez", "miez_modules")); !os.IsNotExist(err) {
		t.Fatalf("miez_modules exists after rejected target change: %v", err)
	}
}
