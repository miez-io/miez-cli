package team

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/manuel/miez-cli/internal/artifacts"
	"github.com/manuel/miez-cli/internal/compile"
	"github.com/manuel/miez-cli/internal/ghclient"
	"github.com/manuel/miez-cli/internal/lockfile"
	"github.com/manuel/miez-cli/internal/modules"
	"github.com/manuel/miez-cli/internal/validate"
	"github.com/manuel/miez-cli/internal/workspace"
)

// Use activates an already-installed team id without contacting GitHub or
// changing the installed source. Any failure preserves the previous active
// team and its generated output.
func (s *Service) Use(ctx context.Context, workspaceValue workspace.Workspace, teamID string) (artifacts.TeamDir, error) {
	if err := ctx.Err(); err != nil {
		return artifacts.TeamDir{}, err
	}
	if ghclient.IsGitHubURL(teamID) {
		return artifacts.TeamDir{}, errors.New("team use accepts an installed team id; use `miez team install <github-url>` for a remote team")
	}
	if !s.IsInstalled(teamID) {
		return artifacts.TeamDir{}, fmt.Errorf("team %q is not installed; use `miez team install <github-url>` first", teamID)
	}
	return s.activateInstalled(workspaceValue, teamID)
}

func (s *Service) activateInstalled(workspaceValue workspace.Workspace, teamID string) (artifacts.TeamDir, error) {
	team, err := s.Load(teamID)
	if err != nil {
		return artifacts.TeamDir{}, err
	}
	if err := validate.Team(team).Error(); err != nil {
		return artifacts.TeamDir{}, err
	}
	if err := validate.Package(team.Root, team.Team); err != nil {
		return artifacts.TeamDir{}, err
	}
	state := workspaceValue.EnsureTeamState(team.Team)
	if err := s.PersistMutation(workspaceValue, team, state, team.Team.ID); err != nil {
		return artifacts.TeamDir{}, err
	}
	return team, nil
}

func (s *Service) installFromGitHub(ctx context.Context, workspaceValue workspace.Workspace, url string) (artifacts.TeamDir, error) {
	ref, err := ghclient.ParseTeamURL(url)
	if err != nil {
		return artifacts.TeamDir{}, err
	}
	token, _, err := s.resolveCredential(ctx, ref.Owner)
	if err != nil {
		return artifacts.TeamDir{}, err
	}

	stageParent, err := os.MkdirTemp("", "miez-team-use-")
	if err != nil {
		return artifacts.TeamDir{}, fmt.Errorf("create team staging directory: %w", err)
	}
	defer os.RemoveAll(stageParent)

	stagedTeam, err := s.stageGitHubTeam(ctx, stageParent, ref, token)
	if err != nil {
		return artifacts.TeamDir{}, err
	}
	return s.installStaged(ctx, workspaceValue, url, stagedTeam)
}

// installStaged validates and installs an already-downloaded stagedTeam,
// replacing the previously active team's source only after the new team
// and its generated output are ready. Any failure preserves the previously
// active team, its source, and its generated output unchanged.
func (s *Service) installStaged(ctx context.Context, workspaceValue workspace.Workspace, url string, stagedTeam staged) (artifacts.TeamDir, error) {
	if err := validate.Team(stagedTeam.Team).Error(); err != nil {
		return artifacts.TeamDir{}, err
	}
	if err := validate.Package(stagedTeam.Team.Root, stagedTeam.Team.Team); err != nil {
		return artifacts.TeamDir{}, err
	}
	if err := ctx.Err(); err != nil {
		return artifacts.TeamDir{}, err
	}
	previousPaths, err := s.previousManagedPaths(workspaceValue)
	if err != nil {
		return artifacts.TeamDir{}, err
	}
	previewState := workspaceValue.EnsureTeamState(stagedTeam.Team.Team)
	if err := s.resolveMCPSecrets(workspaceValue, stagedTeam.Team.Team, previewState); err != nil {
		return artifacts.TeamDir{}, err
	}

	newRoot, err := modules.TeamPath(s.Root, stagedTeam.Team.Team.ID)
	if err != nil {
		return artifacts.TeamDir{}, err
	}
	replaceExisting := false
	info, err := os.Lstat(newRoot)
	if err == nil {
		if workspaceValue.Config.ActiveTeam != stagedTeam.Team.Team.ID {
			return artifacts.TeamDir{}, fmt.Errorf("team %q is already installed; use `miez team use %s` or `miez team update %s`", stagedTeam.Team.Team.ID, stagedTeam.Team.Team.ID, stagedTeam.Team.Team.ID)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return artifacts.TeamDir{}, fmt.Errorf("team source %s must not be a symlink", newRoot)
		}
		if !info.IsDir() {
			return artifacts.TeamDir{}, fmt.Errorf("team source %s is not a directory", newRoot)
		}
		replaceExisting = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return artifacts.TeamDir{}, fmt.Errorf("inspect team destination: %w", err)
	}

	state := workspaceValue.EnsureTeamState(stagedTeam.Team.Team)
	plan, err := compile.Build(workspaceValue, stagedTeam.Team, state)
	if err != nil {
		return artifacts.TeamDir{}, err
	}
	transaction, err := compile.ApplyTransaction(workspaceValue, plan, previousPaths)
	if err != nil {
		return artifacts.TeamDir{}, err
	}

	backupParent := ""
	backupRoot := ""
	oldMoved := false
	rollbackSource := func(cause error) error {
		var rollbackErrors []error
		if !replaceExisting || oldMoved {
			if err := os.RemoveAll(newRoot); err != nil {
				rollbackErrors = append(rollbackErrors, err)
			}
		}
		if oldMoved {
			if err := os.Rename(backupRoot, newRoot); err != nil {
				rollbackErrors = append(rollbackErrors, err)
			}
		}
		if backupParent != "" {
			if err := os.RemoveAll(backupParent); err != nil {
				rollbackErrors = append(rollbackErrors, err)
			}
		}
		rollbackErrors = append(rollbackErrors, transaction.Rollback())
		return errors.Join(append([]error{cause}, rollbackErrors...)...)
	}

	if replaceExisting {
		backupParent, err = os.MkdirTemp(filepath.Dir(newRoot), ".miez-team-replace-")
		if err != nil {
			return artifacts.TeamDir{}, errors.Join(err, transaction.Rollback())
		}
		backupRoot = filepath.Join(backupParent, "old-team")
		if err := os.Rename(newRoot, backupRoot); err != nil {
			return artifacts.TeamDir{}, rollbackSource(fmt.Errorf("stage existing team %q: %w", stagedTeam.Team.Team.ID, err))
		}
		oldMoved = true
	}
	if err := os.MkdirAll(filepath.Dir(newRoot), 0o755); err != nil {
		return artifacts.TeamDir{}, rollbackSource(fmt.Errorf("create modules directory: %w", err))
	}
	if err := os.Mkdir(newRoot, 0o755); err != nil {
		return artifacts.TeamDir{}, rollbackSource(fmt.Errorf("create team destination: %w", err))
	}
	if err := artifacts.CopyDirectory(ctx, stagedTeam.Team.Root, newRoot); err != nil {
		return artifacts.TeamDir{}, rollbackSource(fmt.Errorf("install team %q: %w", stagedTeam.Team.Team.ID, err))
	}
	if err := ctx.Err(); err != nil {
		return artifacts.TeamDir{}, rollbackSource(err)
	}

	installedTeam, err := artifacts.LoadTeam(newRoot)
	if err != nil {
		return artifacts.TeamDir{}, rollbackSource(err)
	}
	entry, err := lockfile.BuildEntry(newRoot, url, stagedTeam.Ref.Branch, stagedTeam.Commit, installedTeam.Team.Version)
	if err != nil {
		return artifacts.TeamDir{}, rollbackSource(err)
	}
	lock, err := lockfile.Load(s.Root)
	if err != nil {
		return artifacts.TeamDir{}, rollbackSource(err)
	}
	if lock.Teams == nil {
		lock.Teams = map[string]lockfile.Entry{}
	}
	lock.Teams[installedTeam.Team.ID] = entry

	workspaceValue.Config.ActiveTeam = installedTeam.Team.ID
	workspaceValue.SetTeamState(installedTeam.Team.ID, state)
	if err := workspaceValue.Save(); err != nil {
		return artifacts.TeamDir{}, rollbackSource(err)
	}
	if err := lockfile.Save(s.Root, lock); err != nil {
		return artifacts.TeamDir{}, rollbackSource(err)
	}
	if err := transaction.Commit(); err != nil {
		return artifacts.TeamDir{}, err
	}
	if backupParent != "" {
		_ = os.RemoveAll(backupParent)
	}
	return installedTeam, nil
}
