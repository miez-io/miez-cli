// Package team implements the deterministic lifecycle operations for
// installing, activating, updating, auditing, and mutating miez teams.
//
// This package owns no interactive I/O beyond confirmation callbacks; the
// cli package owns prompting and output formatting and delegates every
// domain operation here.
package team

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/manuel/miez-cli/internal/artifacts"
	"github.com/manuel/miez-cli/internal/compile"
	"github.com/manuel/miez-cli/internal/credentials"
	"github.com/manuel/miez-cli/internal/ghclient"
	"github.com/manuel/miez-cli/internal/model"
	"github.com/manuel/miez-cli/internal/modules"
	"github.com/manuel/miez-cli/internal/secrets"
	"github.com/manuel/miez-cli/internal/validate"
	"github.com/manuel/miez-cli/internal/workspace"
)

// GitHubHost is the only credential-resolution host this phase supports.
const GitHubHost = "github.com"

// Service performs team installation, activation, and mutation for one
// repository root.
type Service struct {
	// Root is the repository root that owns .miez/, miez.yaml, and the
	// .miez/miez.lock.yaml and .miez/miez_modules/ distribution state.
	Root string
	// GH downloads and resolves GitHub repository content. Tests may point
	// it at a fake API server.
	GH *ghclient.Client
	// Credentials resolves a GitHub bearer token for module access. Tests
	// may replace it with a fake environment/CLI.
	Credentials *credentials.Resolver
	// Log, when set, receives a diagnostic line whenever a credential
	// resolves, naming the chain tier used (wired to a --verbose flag).
	Log func(message string)
	// Interactive reports whether a missing MCP secret may be prompted for.
	// False by default, so a Service used without CLI wiring never blocks
	// on stdin.
	Interactive bool
	// PromptSecret prompts for one missing MCP secret value. Required
	// whenever Interactive is true.
	PromptSecret secrets.Prompt
}

// NewService creates a Service rooted at root with real GitHub and
// credential backends.
func NewService(root string) *Service {
	return &Service{Root: root, GH: ghclient.New(), Credentials: credentials.NewResolver()}
}

// ModulesPath returns the raw downloaded team source root below Root.
func (s *Service) ModulesPath() string {
	return modules.Root(s.Root)
}

// List returns every installed team in load order.
func (s *Service) List() ([]artifacts.TeamDir, error) {
	return artifacts.LoadAll(s.ModulesPath())
}

// Load reads and shape-checks one installed team by id.
func (s *Service) Load(teamID string) (artifacts.TeamDir, error) {
	if strings.TrimSpace(teamID) == "" {
		return artifacts.TeamDir{}, errors.New("team id is required")
	}
	path, err := modules.TeamPath(s.Root, teamID)
	if err != nil {
		return artifacts.TeamDir{}, err
	}
	return artifacts.LoadTeam(path)
}

// IsInstalled reports whether teamID resolves to an installed team.
func (s *Service) IsInstalled(teamID string) bool {
	_, err := s.Load(teamID)
	return err == nil
}

// Check validates one team by id, or every installed team when teamID is
// empty.
func (s *Service) Check(teamID string) ([]artifacts.TeamDir, error) {
	if teamID != "" {
		team, err := s.Load(teamID)
		if err != nil {
			return nil, err
		}
		return []artifacts.TeamDir{team}, nil
	}
	return s.List()
}

// Resolve loads and validates the workspace's active team and its state.
func (s *Service) Resolve(workspaceValue workspace.Workspace) (artifacts.TeamDir, model.TeamState, error) {
	if workspaceValue.Config.ActiveTeam == "" {
		return artifacts.TeamDir{}, model.TeamState{}, errors.New("no active team; run `miez team install <github-url>` or `miez team use <team-id>` first")
	}
	team, err := s.Load(workspaceValue.Config.ActiveTeam)
	if err != nil {
		return artifacts.TeamDir{}, model.TeamState{}, err
	}
	if err := validate.Team(team).Error(); err != nil {
		return artifacts.TeamDir{}, model.TeamState{}, err
	}
	if err := validate.Package(team.Root, team.Team); err != nil {
		return artifacts.TeamDir{}, model.TeamState{}, err
	}
	state := workspaceValue.EnsureTeamState(team.Team)
	return team, state, nil
}

// PersistMutation compiles team, stages the generated output, updates
// workspace state, and commits or rolls back the whole change as one unit.
func (s *Service) PersistMutation(workspaceValue workspace.Workspace, team artifacts.TeamDir, state model.TeamState, activeTeamID string) error {
	previousPaths, err := s.previousManagedPaths(workspaceValue)
	if err != nil {
		return err
	}
	plan, err := compile.Build(workspaceValue, team, state)
	if err != nil {
		return err
	}
	transaction, err := compile.ApplyTransaction(workspaceValue, plan, previousPaths)
	if err != nil {
		return err
	}
	if activeTeamID != "" {
		workspaceValue.Config.ActiveTeam = activeTeamID
	}
	workspaceValue.SetTeamState(team.Team.ID, state)
	if err := workspaceValue.Save(); err != nil {
		return errors.Join(err, transaction.Rollback())
	}
	return transaction.Commit()
}

func (s *Service) previousManagedPaths(workspaceValue workspace.Workspace) ([]string, error) {
	paths, err := workspaceValue.LegacyManagedFiles()
	if err != nil {
		return nil, err
	}
	activeTeamID := workspaceValue.Config.ActiveTeam
	if activeTeamID == "" {
		return paths, nil
	}
	activeTeam, err := s.Load(activeTeamID)
	if err != nil {
		return nil, err
	}
	managed, err := compile.ManagedPaths(activeTeam, workspaceValue.State(activeTeamID))
	if err != nil {
		return nil, err
	}
	return append(paths, managed...), nil
}

// resolveCredential resolves a GitHub credential for owner, returning an
// empty token (not an error) when nothing in the chain resolves.
func (s *Service) resolveCredential(ctx context.Context, owner string) (string, string, error) {
	credential, ok, err := s.Credentials.Resolve(ctx, owner, GitHubHost)
	if err != nil {
		return "", "", err
	}
	if !ok {
		return "", "", nil
	}
	if s.Log != nil {
		s.Log(fmt.Sprintf("github credential resolved via %s", credential.Source))
	}
	return credential.Token, credential.Source, nil
}

// resolveMCPSecrets gathers every MCP env placeholder required by teamValue's
// currently enabled workers, resolves each against the local secrets store
// or the process environment, prompts interactively when s.Interactive
// allows it, and persists newly resolved values back to the local store.
// It runs before any filesystem mutation, so a failure here leaves nothing
// to roll back.
func (s *Service) resolveMCPSecrets(workspaceValue workspace.Workspace, teamValue model.Team, state model.TeamState) error {
	active := teamValue.EnabledWorkerIDs(state)
	toolIDs := []string{}
	seen := map[string]struct{}{}
	for _, worker := range teamValue.Workers {
		if _, enabled := active[worker.ID]; !enabled {
			continue
		}
		for _, tool := range worker.Tools {
			if _, exists := seen[tool]; exists {
				continue
			}
			seen[tool] = struct{}{}
			toolIDs = append(toolIDs, tool)
		}
	}
	if len(toolIDs) == 0 {
		return nil
	}
	sort.Strings(toolIDs)
	requirements, err := secrets.RequiredFor(teamValue, toolIDs)
	if err != nil {
		return err
	}
	if len(requirements) == 0 {
		return nil
	}
	store := secrets.Store{Path: workspaceValue.McpLocalPath()}
	overrides, err := store.Load()
	if err != nil {
		return err
	}
	resolver := secrets.Resolver{Overrides: overrides, LookupEnv: os.LookupEnv}
	interactive := s.Interactive && s.PromptSecret != nil
	resolved, err := secrets.Resolve(requirements, resolver, interactive, s.PromptSecret)
	if err != nil {
		return err
	}
	return store.Merge(resolved)
}

// findTeamRoot locates the single generated team index below an extracted
// archive when no in-repository team path was given.
func findTeamRoot(root string) (string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == artifacts.TeamIndexFileName {
			matches = append(matches, filepath.Dir(path))
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("inspect downloaded team: %w", err)
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("downloaded repository must contain exactly one %s, found %d; use a /tree/<branch>/<path> URL to select one team", artifacts.TeamIndexFileName, len(matches))
	}
	return matches[0], nil
}

// findTeamAtPath resolves the team at teamPath inside the archive extracted
// below root, without requiring it to be the only generated team index in the
// repository. This is what lets one repository host multiple teams.
func findTeamAtPath(root, teamPath string) (string, error) {
	topLevel, err := ghclient.SingleTopLevelDir(root)
	if err != nil {
		return "", err
	}
	clean := filepath.Clean(filepath.FromSlash(teamPath))
	if !filepath.IsLocal(clean) || clean == "." {
		return "", fmt.Errorf("GitHub team path %q escapes the repository", teamPath)
	}
	candidate := filepath.Join(topLevel, clean)
	if info, err := os.Stat(candidate); err != nil || !info.IsDir() {
		return "", fmt.Errorf("path %q not found in repository", teamPath)
	}
	if _, err := os.Stat(filepath.Join(candidate, artifacts.TeamIndexFileName)); err != nil {
		return "", fmt.Errorf("no %s found at %q in repository", artifacts.TeamIndexFileName, teamPath)
	}
	return candidate, nil
}
