package team

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/manuel/miez-cli/internal/artifacts"
	"github.com/manuel/miez-cli/internal/ghclient"
	"github.com/manuel/miez-cli/internal/manifest"
	"github.com/manuel/miez-cli/internal/model"
)

// staged is one downloaded team, resolved to an exact commit.
type staged struct {
	Team   artifacts.TeamDir
	Ref    ghclient.Ref
	Commit string
}

// stageGitHubTeam resolves ref's branch to a commit, downloads that exact
// commit's tarball below stageParent, and loads the generated team index it
// contains (at ref.TeamPath, or the archive's only index when TeamPath is empty).
func (s *Service) stageGitHubTeam(ctx context.Context, stageParent string, ref ghclient.Ref, token string) (staged, error) {
	commit, err := s.GH.ResolveCommit(ctx, ref.Owner, ref.Repository, ref.Branch, token)
	if err != nil {
		return staged{}, err
	}
	downloadRoot := filepath.Join(stageParent, "download")
	if err := s.GH.DownloadTarball(ctx, ref.Owner, ref.Repository, commit, token, downloadRoot); err != nil {
		return staged{}, err
	}

	var teamRoot string
	if ref.TeamPath == "" {
		teamRoot, err = findTeamRoot(downloadRoot)
	} else {
		teamRoot, err = findTeamAtPath(downloadRoot, ref.TeamPath)
	}
	if err != nil {
		return staged{}, err
	}

	team, err := artifacts.LoadTeam(teamRoot)
	if err != nil {
		return staged{}, err
	}
	if _, err := manifest.LoadTeamPackage(teamRoot); err != nil {
		return staged{}, err
	}
	if !model.IsValidIdentifier(team.Team.ID) {
		return staged{}, fmt.Errorf("downloaded team id %q must be kebab-case", team.Team.ID)
	}
	canonicalRoot := filepath.Join(stageParent, team.Team.ID)
	if err := os.Rename(teamRoot, canonicalRoot); err != nil {
		return staged{}, fmt.Errorf("canonicalize downloaded team directory: %w", err)
	}
	team.Root = canonicalRoot
	return staged{Team: team, Ref: ref, Commit: commit}, nil
}
