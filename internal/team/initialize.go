package team

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/manuel/miez-cli/internal/ghclient"
	"github.com/manuel/miez-cli/internal/workspace"
)

// Install installs and activates the GitHub team at githubURL. A missing
// workspace is prepared as part of the first successful installation; an
// existing workspace keeps its persisted target configuration. Any failure
// leaves the previous active team and generated output usable.
func (s *Service) Install(ctx context.Context, targets []string, githubURL string) (string, error) {
	if strings.TrimSpace(githubURL) == "" {
		return "", errors.New("a GitHub team URL is required")
	}
	if _, err := ghclient.ParseTeamURL(githubURL); err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	workspaceValue, err := workspace.Open(s.Root)
	existingWorkspace := err == nil
	if err != nil && !errors.Is(err, workspace.ErrNotInitialized) {
		return "", err
	}
	if existingWorkspace {
		if len(targets) > 0 && !sameTargets(workspaceValue.Config.Targets, targets) {
			return "", fmt.Errorf("requested targets %v do not match initialized workspace targets %v", targets, workspaceValue.Config.Targets)
		}
	} else {
		if len(targets) == 0 {
			return "", errors.New("a render target is required for the first team installation")
		}
		workspaceValue, err = workspace.New(s.Root, targets, "")
		if err != nil {
			return "", err
		}
	}
	preparation, err := workspaceValue.Prepare()
	if err != nil {
		return "", err
	}
	migration, err := s.prepareLegacyDistribution()
	if err != nil {
		_ = preparation.Rollback()
		return "", err
	}
	committed := false
	defer func() {
		if !committed {
			_ = migration.Rollback()
			_ = preparation.Rollback()
		}
	}()

	team, err := s.installFromGitHub(ctx, workspaceValue, githubURL)
	if err != nil {
		return "", err
	}
	removeEmptyLegacyState(s.Root)
	migration.Commit()
	preparation.Commit()
	committed = true
	return team.Team.ID, nil
}

func sameTargets(left, right []string) bool {
	return strings.Join(normalizedTargets(left), "\x00") == strings.Join(normalizedTargets(right), "\x00")
}

func normalizedTargets(values []string) []string {
	seen := map[string]struct{}{}
	normalized := []string{}
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	sort.Strings(normalized)
	return normalized
}
