package team

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/manuel/miez-cli/internal/workspace"
)

func TestPlanUpdateReportsUpToDateWithoutDownloading(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "team-a", "main", "sha1", demoTeamFiles("team-a", "Team A"))
	root := t.TempDir()
	initWorkspace(t, root)
	service := newTestService(root, fake, map[string]string{})
	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/team-a"); err != nil {
		t.Fatal(err)
	}

	plan, err := service.PlanUpdate(context.Background(), "team-a")
	if err != nil {
		t.Fatal(err)
	}
	defer plan.Close()
	if !plan.UpToDate {
		t.Fatalf("plan = %#v, want UpToDate", plan)
	}
}

func TestPlanUpdateDefaultsToActiveTeam(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "team-a", "main", "sha1", demoTeamFiles("team-a", "Team A"))
	root := t.TempDir()
	service := newTestService(root, fake, map[string]string{})
	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/team-a"); err != nil {
		t.Fatal(err)
	}

	plan, err := service.PlanUpdate(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer plan.Close()
	if plan.TeamID != "team-a" {
		t.Fatalf("TeamID = %q, want active team team-a", plan.TeamID)
	}
	if !plan.UpToDate {
		t.Fatal("plan is not up to date")
	}
}

func TestPlanUpdateWithoutActiveTeamFailsLocally(t *testing.T) {
	root := t.TempDir()
	initWorkspace(t, root)
	service := newTestService(root, newFakeGitHub(t), map[string]string{})

	if _, err := service.PlanUpdate(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "no active team") {
		t.Fatalf("PlanUpdate error = %v, want no active team", err)
	}
}

func TestPlanAndApplyUpdateInstallsNewCommit(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "team-a", "main", "sha1", demoTeamFiles("team-a", "Team A"))
	root := t.TempDir()
	initWorkspace(t, root)
	service := newTestService(root, fake, map[string]string{})
	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/team-a"); err != nil {
		t.Fatal(err)
	}

	files := demoTeamFiles("team-a", "Team A")
	files["team-a/workers/builder.md"] = "---\n---\n# Builder v2\n"
	files["team-a/workers/extra.md"] = "---\n---\n# Extra\n"
	fake.setTeam("acme", "team-a", "main", "sha2", files)

	plan, err := service.PlanUpdate(context.Background(), "team-a")
	if err != nil {
		t.Fatal(err)
	}
	defer plan.Close()
	if plan.UpToDate {
		t.Fatal("plan reports UpToDate despite a new commit")
	}
	if plan.NewCommit != "sha2" {
		t.Fatalf("NewCommit = %q, want sha2", plan.NewCommit)
	}
	if len(plan.AddedFiles) != 1 || plan.AddedFiles[0] != "workers/extra.md" {
		t.Fatalf("AddedFiles = %#v", plan.AddedFiles)
	}
	if len(plan.ChangedFiles) != 1 || plan.ChangedFiles[0] != "workers/builder.md" {
		t.Fatalf("ChangedFiles = %#v", plan.ChangedFiles)
	}

	reopened, err := workspace.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApplyUpdate(context.Background(), reopened, plan); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".miez", "miez_modules", "team-a", "workers", "extra.md")); err != nil {
		t.Fatalf("update did not install the new file: %v", err)
	}
}

func TestApplyUpdateIsNoOpWhenUpToDate(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "team-a", "main", "sha1", demoTeamFiles("team-a", "Team A"))
	root := t.TempDir()
	initWorkspace(t, root)
	service := newTestService(root, fake, map[string]string{})
	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/team-a"); err != nil {
		t.Fatal(err)
	}
	lockBefore, err := os.ReadFile(filepath.Join(root, ".miez", "miez.lock.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	plan, err := service.PlanUpdate(context.Background(), "team-a")
	if err != nil {
		t.Fatal(err)
	}
	defer plan.Close()
	reopened, err := workspace.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApplyUpdate(context.Background(), reopened, plan); err != nil {
		t.Fatal(err)
	}
	lockAfter, err := os.ReadFile(filepath.Join(root, ".miez", "miez.lock.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(lockBefore) != string(lockAfter) {
		t.Fatal("ApplyUpdate wrote the lockfile despite being up to date")
	}
}
