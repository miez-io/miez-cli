package team

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAuditReportsCleanWhenMatching(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "team-a", "main", "sha1", demoTeamFiles("team-a", "Team A"))
	root := t.TempDir()
	initWorkspace(t, root)
	service := newTestService(root, fake, map[string]string{})
	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/team-a"); err != nil {
		t.Fatal(err)
	}

	reports, err := service.Audit("team-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 || !reports[0].Clean() {
		t.Fatalf("reports = %#v, want one clean report", reports)
	}
}

func TestAuditDetectsModifiedMissingAndExtra(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "team-a", "main", "sha1", demoTeamFiles("team-a", "Team A"))
	root := t.TempDir()
	initWorkspace(t, root)
	service := newTestService(root, fake, map[string]string{})
	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/team-a"); err != nil {
		t.Fatal(err)
	}

	teamRoot := filepath.Join(root, ".miez", "miez_modules", "team-a")
	if err := os.WriteFile(filepath.Join(teamRoot, "miez.generated.yaml"), []byte("id: team-a\nversion: 9.9.9\nname: Tampered\nworkers: []\nworkflows: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(teamRoot, "workers", "builder.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(teamRoot, "workers", "undeclared.md"), []byte("extra"), 0o644); err != nil {
		t.Fatal(err)
	}

	reports, err := service.Audit("team-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 {
		t.Fatalf("reports = %#v", reports)
	}
	findings := map[string]string{}
	for _, finding := range reports[0].Findings {
		findings[finding.Path] = finding.Status
	}
	if findings["miez.generated.yaml"] != StatusModified {
		t.Errorf("generated index status = %q, want modified", findings["miez.generated.yaml"])
	}
	if findings["workers/builder.md"] != StatusMissing {
		t.Errorf("workers/builder.md status = %q, want missing", findings["workers/builder.md"])
	}
	if findings["workers/undeclared.md"] != StatusExtra {
		t.Errorf("workers/undeclared.md status = %q, want extra", findings["workers/undeclared.md"])
	}
	if reports[0].Clean() {
		t.Fatal("report reports clean despite drift")
	}
}

func TestAuditUnknownTeamFails(t *testing.T) {
	root := t.TempDir()
	initWorkspace(t, root)
	service := newTestService(root, newFakeGitHub(t), map[string]string{})
	if _, err := service.Audit("not-installed"); err == nil {
		t.Fatal("Audit succeeded for a team that is not installed")
	}
}
