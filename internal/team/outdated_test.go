package team

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/manuel/miez-cli/internal/lockfile"
)

func TestOutdatedWritesNothing(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "team-a", "main", "sha1", demoTeamFiles("team-a", "Team A"))
	root := t.TempDir()
	initWorkspace(t, root)
	service := newTestService(root, fake, map[string]string{})
	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/team-a"); err != nil {
		t.Fatal(err)
	}
	fake.setTeam("acme", "team-a", "main", "sha2", demoTeamFiles("team-a", "Team A"))

	before := snapshotTree(t, root)
	if _, err := service.Outdated(context.Background(), "team-a"); err != nil {
		t.Fatal(err)
	}
	after := snapshotTree(t, root)
	if before != after {
		t.Fatalf("Outdated changed on-disk state:\nbefore=%s\nafter=%s", before, after)
	}
}

func snapshotTree(t *testing.T, root string) string {
	t.Helper()
	var entries []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		info, statErr := entry.Info()
		if statErr != nil {
			return statErr
		}
		entries = append(entries, relative+":"+info.Mode().String())
		if !entry.IsDir() {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			entries[len(entries)-1] += ":" + string(data)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	result := ""
	for _, entry := range entries {
		result += entry + "\n"
	}
	return result
}

func TestOutdatedReportsWithoutChangingAnything(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "team-a", "main", "sha1", demoTeamFiles("team-a", "Team A"))
	root := t.TempDir()
	initWorkspace(t, root)
	service := newTestService(root, fake, map[string]string{})
	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/team-a"); err != nil {
		t.Fatal(err)
	}

	reports, err := service.Outdated(context.Background(), "team-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 || !reports[0].UpToDate {
		t.Fatalf("reports = %#v, want one up-to-date report", reports)
	}

	fake.setTeam("acme", "team-a", "main", "sha2", demoTeamFiles("team-a", "Team A"))
	reports, err = service.Outdated(context.Background(), "team-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 || reports[0].UpToDate || reports[0].LatestCommit != "sha2" {
		t.Fatalf("reports = %#v, want an outdated report pointing at sha2", reports)
	}
}

func TestOutdatedNamedTeamNotInstalledFails(t *testing.T) {
	root := t.TempDir()
	initWorkspace(t, root)
	service := newTestService(root, newFakeGitHub(t), map[string]string{})

	if _, err := service.Outdated(context.Background(), "not-installed"); err == nil {
		t.Fatal("Outdated succeeded for a team that is not installed")
	}
}

func TestOutdatedWithNoTeamNameReportsEveryInstalledTeam(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "team-a", "main", "sha1", demoTeamFiles("team-a", "Team A"))
	fake.setTeam("acme", "team-b", "main", "sha1", demoTeamFiles("team-b", "Team B"))
	root := t.TempDir()
	initWorkspace(t, root)
	lock := lockfile.Lockfile{Teams: map[string]lockfile.Entry{
		"team-a": {RepoURL: "https://github.com/acme/team-a", Ref: "main", Commit: "sha1"},
		"team-b": {RepoURL: "https://github.com/acme/team-b", Ref: "main", Commit: "sha1"},
	}}
	if err := lockfile.Save(root, lock); err != nil {
		t.Fatal(err)
	}
	service := newTestService(root, fake, map[string]string{})

	reports, err := service.Outdated(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 2 {
		t.Fatalf("reports = %#v, want 2 entries", reports)
	}
}
