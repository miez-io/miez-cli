package team

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/manuel/miez-cli/internal/artifacts"
	"github.com/manuel/miez-cli/internal/ghclient"
	"github.com/manuel/miez-cli/internal/lockfile"
	"github.com/manuel/miez-cli/internal/workspace"
)

// UpdatePlan describes the file and version changes an update would apply.
type UpdatePlan struct {
	TeamID       string
	RepoURL      string
	Ref          string
	OldCommit    string
	NewCommit    string
	UpToDate     bool
	AddedFiles   []string
	RemovedFiles []string
	ChangedFiles []string

	stagedTeam    staged
	stageParent   string
	repositoryURL string
}

// Close releases any staged download PlanUpdate produced. Safe to call on a
// zero-value or already-applied plan.
func (plan UpdatePlan) Close() {
	if plan.stageParent != "" {
		_ = os.RemoveAll(plan.stageParent)
	}
}

// PlanUpdate resolves teamIDOrURL to an installed team's lockfile entry. When
// the reference is empty, it uses the workspace's active team. It then
// re-resolves the latest matching ref (one lightweight API call). Only when
// that ref differs from the locked commit does it download the tarball, so
// the common "already up to date" case never re-downloads anything. Call
// Close on the returned plan once done with it, whether or not it is applied.
func (s *Service) PlanUpdate(ctx context.Context, teamIDOrURL string) (UpdatePlan, error) {
	if strings.TrimSpace(teamIDOrURL) == "" {
		workspaceValue, err := workspace.Open(s.Root)
		if err != nil {
			return UpdatePlan{}, fmt.Errorf("resolve active team for update: %w", err)
		}
		if workspaceValue.Config.ActiveTeam == "" {
			return UpdatePlan{}, fmt.Errorf("no active team; run `miez team install <github-url>` or `miez team use <team-id>` first")
		}
		teamIDOrURL = workspaceValue.Config.ActiveTeam
	}
	teamID, entry, err := s.resolveInstalledEntry(teamIDOrURL)
	if err != nil {
		return UpdatePlan{}, err
	}
	ref, err := ghclient.ParseTeamURL(entry.RepoURL)
	if err != nil {
		return UpdatePlan{}, err
	}
	ref.Branch = entry.Ref
	token, _, err := s.resolveCredential(ctx, ref.Owner)
	if err != nil {
		return UpdatePlan{}, err
	}
	newCommit, err := s.GH.ResolveCommit(ctx, ref.Owner, ref.Repository, ref.Branch, token)
	if err != nil {
		return UpdatePlan{}, err
	}
	base := UpdatePlan{TeamID: teamID, RepoURL: entry.RepoURL, Ref: entry.Ref, OldCommit: entry.Commit, NewCommit: newCommit}
	if newCommit == entry.Commit {
		base.UpToDate = true
		return base, nil
	}

	stageParent, err := os.MkdirTemp("", "miez-team-update-")
	if err != nil {
		return UpdatePlan{}, fmt.Errorf("create team staging directory: %w", err)
	}
	stagedTeam, err := s.stageGitHubTeam(ctx, stageParent, ref, token)
	if err != nil {
		_ = os.RemoveAll(stageParent)
		return UpdatePlan{}, err
	}

	newFiles, err := filesWithHashes(stagedTeam.Team.Root)
	if err != nil {
		_ = os.RemoveAll(stageParent)
		return UpdatePlan{}, err
	}
	added, removed, changed := diffFiles(entry.FileHashes, newFiles)
	base.AddedFiles, base.RemovedFiles, base.ChangedFiles = added, removed, changed
	base.stagedTeam = stagedTeam
	base.stageParent = stageParent
	base.repositoryURL = entry.RepoURL
	return base, nil
}

// ApplyUpdate installs plan's already-staged download over the team's
// current install and rewrites its lockfile entry. It is a no-op when plan
// reports UpToDate.
func (s *Service) ApplyUpdate(ctx context.Context, workspaceValue workspace.Workspace, plan UpdatePlan) (artifacts.TeamDir, error) {
	if plan.UpToDate {
		return s.Load(plan.TeamID)
	}
	return s.installStaged(ctx, workspaceValue, plan.repositoryURL, plan.stagedTeam)
}

// resolveInstalledEntry resolves a literal URL or installed team id to its
// lockfile entry.
func (s *Service) resolveInstalledEntry(teamIDOrURL string) (string, lockfile.Entry, error) {
	lock, err := lockfile.Load(s.Root)
	if err != nil {
		return "", lockfile.Entry{}, err
	}
	teamID := teamIDOrURL
	if ghclient.IsGitHubURL(teamIDOrURL) {
		for id, entry := range lock.Teams {
			if entry.RepoURL == teamIDOrURL {
				teamID = id
				break
			}
		}
	}
	entry, exists := lock.Teams[teamID]
	if !exists {
		return "", lockfile.Entry{}, fmt.Errorf("team %q is not installed; run `miez team install <github-url>` first", teamIDOrURL)
	}
	return teamID, entry, nil
}

func diffFiles(oldHashes map[string]string, newFiles map[string]string) (added, removed, changed []string) {
	for path, hash := range newFiles {
		oldHash, existed := oldHashes[path]
		if !existed {
			added = append(added, path)
		} else if oldHash != hash {
			changed = append(changed, path)
		}
	}
	for path := range oldHashes {
		if _, exists := newFiles[path]; !exists {
			removed = append(removed, path)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(changed)
	return added, removed, changed
}

func filesWithHashes(root string) (map[string]string, error) {
	files, err := artifacts.WalkFiles(root)
	if err != nil {
		return nil, err
	}
	hashes := make(map[string]string, len(files))
	for _, relative := range files {
		hash, err := lockfile.HashFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			return nil, err
		}
		hashes[relative] = hash
	}
	return hashes, nil
}
