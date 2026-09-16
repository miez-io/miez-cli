package ghclient_test

import (
	"testing"

	"github.com/manuel/miez-cli/internal/ghclient"
)

func TestParseTeamURL(t *testing.T) {
	cases := []struct {
		url        string
		owner      string
		repository string
		branch     string
		teamPath   string
	}{
		{"https://github.com/owner/repo", "owner", "repo", "main", ""},
		{"https://github.com/owner/repo/tree/develop", "owner", "repo", "develop", ""},
		{"https://github.com/owner/repo/projects/team-a", "owner", "repo", "main", "projects/team-a"},
		{"https://github.com/owner/repo/tree/develop/projects/team-a", "owner", "repo", "develop", "projects/team-a"},
		{"https://github.com/owner/repo.git", "owner", "repo", "main", ""},
	}
	for _, testCase := range cases {
		ref, err := ghclient.ParseTeamURL(testCase.url)
		if err != nil {
			t.Fatalf("ParseTeamURL(%q) error = %v", testCase.url, err)
		}
		if ref.Owner != testCase.owner || ref.Repository != testCase.repository || ref.Branch != testCase.branch || ref.TeamPath != testCase.teamPath {
			t.Fatalf("ParseTeamURL(%q) = %#v, want owner=%q repo=%q branch=%q path=%q",
				testCase.url, ref, testCase.owner, testCase.repository, testCase.branch, testCase.teamPath)
		}
	}
}

func TestParseTeamURLRejectsTreeWithoutBranch(t *testing.T) {
	if _, err := ghclient.ParseTeamURL("https://github.com/owner/repo/tree"); err == nil {
		t.Fatal("ParseTeamURL accepted tree/ without a branch")
	}
}

func TestParseTeamURLRejectsQueryAndFragment(t *testing.T) {
	if _, err := ghclient.ParseTeamURL("https://github.com/owner/repo?x=1"); err == nil {
		t.Fatal("ParseTeamURL accepted a query string")
	}
	if _, err := ghclient.ParseTeamURL("https://github.com/owner/repo#frag"); err == nil {
		t.Fatal("ParseTeamURL accepted a fragment")
	}
}

func TestIsGitHubURL(t *testing.T) {
	for _, value := range []string{"https://github.com/owner/repo", "http://www.github.com/owner/repo"} {
		if !ghclient.IsGitHubURL(value) {
			t.Errorf("IsGitHubURL(%q) = false", value)
		}
	}
	for _, value := range []string{"not-a-url", "https://gitlab.com/owner/repo", "owner-repo"} {
		if ghclient.IsGitHubURL(value) {
			t.Errorf("IsGitHubURL(%q) = true", value)
		}
	}
}
