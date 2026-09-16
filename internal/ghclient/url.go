package ghclient

import (
	"fmt"
	"net/url"
	"strings"
)

// Ref is a parsed GitHub team reference: a repository, a branch (defaulting
// to "main"), and an optional in-repository team path, letting one
// repository host multiple teams the way an APM package hosts multiple
// projects.
type Ref struct {
	Owner      string
	Repository string
	Branch     string
	TeamPath   string
}

// IsGitHubURL reports whether value is an HTTP(S) URL addressed to GitHub.
func IsGitHubURL(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	validScheme := parsed.Scheme == "https" || parsed.Scheme == "http"
	validHost := parsed.Hostname() == "github.com" || parsed.Hostname() == "www.github.com"
	return validScheme && validHost
}

// ParseTeamURL validates repositoryURL and splits it into a Ref. Supported
// shapes: owner/repo, owner/repo/tree/<branch>, owner/repo/<path...>, and
// owner/repo/tree/<branch>/<path...>.
func ParseTeamURL(repositoryURL string) (Ref, error) {
	parsed, parseErr := url.Parse(strings.TrimSpace(repositoryURL))
	if parseErr != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return Ref{}, fmt.Errorf("GitHub team URL must be an http(s) URL")
	}
	if parsed.Hostname() != "github.com" && parsed.Hostname() != "www.github.com" {
		return Ref{}, fmt.Errorf("GitHub team URL must use github.com")
	}
	if parsed.Port() != "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return Ref{}, fmt.Errorf("GitHub team URL must not contain credentials, a port, a query, or a fragment")
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return Ref{}, fmt.Errorf("GitHub team URL must look like https://github.com/owner/repository")
	}
	owner, repository := parts[0], strings.TrimSuffix(parts[1], ".git")
	branch := "main"
	rest := parts[2:]
	if len(rest) > 0 && rest[0] == "tree" {
		if len(rest) < 2 || rest[1] == "" {
			return Ref{}, fmt.Errorf("GitHub team URL must include a branch after tree/")
		}
		branch = rest[1]
		rest = rest[2:]
	}
	return Ref{Owner: owner, Repository: repository, Branch: branch, TeamPath: strings.Join(rest, "/")}, nil
}
