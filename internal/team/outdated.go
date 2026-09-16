package team

import (
	"context"
	"fmt"
	"sort"

	"github.com/manuel/miez-cli/internal/ghclient"
	"github.com/manuel/miez-cli/internal/lockfile"
)

// OutdatedReport compares one installed team's locked ref against the
// latest matching ref available upstream, without changing anything.
type OutdatedReport struct {
	TeamID        string
	Ref           string
	CurrentCommit string
	LatestCommit  string
	UpToDate      bool
}

// Outdated reports on teamID, or on every installed team when teamID is
// empty. It never writes to .miez/miez_modules/, .miez/miez.lock.yaml, or any rendered
// output: it only resolves each team's ref to its latest commit.
func (s *Service) Outdated(ctx context.Context, teamID string) ([]OutdatedReport, error) {
	lock, err := lockfile.Load(s.Root)
	if err != nil {
		return nil, err
	}
	teamIDs := []string{}
	if teamID != "" {
		if _, ok := lock.Teams[teamID]; !ok {
			return nil, fmt.Errorf("team %q is not installed", teamID)
		}
		teamIDs = append(teamIDs, teamID)
	} else {
		for id := range lock.Teams {
			teamIDs = append(teamIDs, id)
		}
		sort.Strings(teamIDs)
	}

	reports := make([]OutdatedReport, 0, len(teamIDs))
	for _, id := range teamIDs {
		entry := lock.Teams[id]
		ref, err := ghclient.ParseTeamURL(entry.RepoURL)
		if err != nil {
			return nil, err
		}
		ref.Branch = entry.Ref
		token, _, err := s.resolveCredential(ctx, ref.Owner)
		if err != nil {
			return nil, err
		}
		latest, err := s.GH.ResolveCommit(ctx, ref.Owner, ref.Repository, ref.Branch, token)
		if err != nil {
			return nil, err
		}
		reports = append(reports, OutdatedReport{
			TeamID:        id,
			Ref:           entry.Ref,
			CurrentCommit: entry.Commit,
			LatestCommit:  latest,
			UpToDate:      latest == entry.Commit,
		})
	}
	return reports, nil
}
