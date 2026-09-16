// Package workspace owns miez's repository-local state and directories.
package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/manuel/miez-cli/internal/model"
	"gopkg.in/yaml.v3"
)

const (
	MiezDir    = ".miez"
	ConfigFile = "config.yaml"

	legacyManifestFile = "manifest.json"
	legacyCompiledFile = "compiled.json"
)

// ErrNotInitialized identifies a repository without a persisted miez
// workspace configuration.
var ErrNotInitialized = errors.New("workspace is not initialized")

// Workspace is a repository root with loaded miez state.
type Workspace struct {
	Root   string
	Config model.Config
}

// Open loads an initialized workspace from root.
func Open(root string) (Workspace, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Workspace{}, fmt.Errorf("resolve workspace: %w", err)
	}
	data, err := os.ReadFile(filepath.Join(root, MiezDir, ConfigFile))
	if errors.Is(err, os.ErrNotExist) {
		return Workspace{}, fmt.Errorf("%w; run `miez team install <github-url>` first", ErrNotInitialized)
	}
	if err != nil {
		return Workspace{}, fmt.Errorf("read workspace config: %w", err)
	}

	var config model.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Workspace{}, fmt.Errorf("parse workspace config: %w", err)
	}
	if config.Version == 0 {
		config.Version = model.ConfigVersion
	}
	if config.Version > model.ConfigVersion {
		return Workspace{}, fmt.Errorf("workspace config version %d is newer than supported version %d", config.Version, model.ConfigVersion)
	}
	if config.TeamStates == nil {
		config.TeamStates = map[string]model.TeamState{}
	}
	config.Targets = normalizeTargets(config.Targets)
	return Workspace{Root: root, Config: config}, nil
}

// New creates an unpersisted workspace configuration after checking that the
// repository has not already been initialized.
func New(root string, targets []string, activeTeam string) (Workspace, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Workspace{}, fmt.Errorf("resolve workspace: %w", err)
	}
	configPath := filepath.Join(root, MiezDir, ConfigFile)
	if _, err := os.Stat(configPath); err == nil {
		return Workspace{}, fmt.Errorf("workspace is already initialized at %s", configPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return Workspace{}, fmt.Errorf("check workspace config: %w", err)
	}

	return Workspace{
		Root: root,
		Config: model.Config{
			Version:    model.ConfigVersion,
			Targets:    normalizeTargets(targets),
			ActiveTeam: activeTeam,
			TeamStates: map[string]model.TeamState{},
		},
	}, nil
}

// Preparation tracks repository-local setup until initialization commits.
type Preparation struct {
	workspace     Workspace
	createdDirs   []string
	ignorePath    string
	ignoreExisted bool
	ignoreData    []byte
	finalized     bool
}

// Prepare creates repository-local state directories without persisting the
// configuration. Commit or Rollback the returned preparation after the rest of
// initialization succeeds or fails.
func (workspace Workspace) Prepare() (*Preparation, error) {
	preparation := &Preparation{
		workspace:   workspace,
		createdDirs: []string{},
		ignorePath:  filepath.Join(workspace.Root, ".gitignore"),
	}
	data, err := os.ReadFile(preparation.ignorePath)
	if err == nil {
		preparation.ignoreExisted = true
		preparation.ignoreData = append([]byte{}, data...)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read .gitignore: %w", err)
	}

	for _, directory := range workspace.stateDirectories() {
		_, statErr := os.Stat(directory)
		if errors.Is(statErr, os.ErrNotExist) {
			preparation.createdDirs = append(preparation.createdDirs, directory)
		} else if statErr != nil {
			return nil, fmt.Errorf("inspect %s: %w", directory, statErr)
		}
		if err := os.MkdirAll(directory, 0o755); err != nil {
			_ = preparation.Rollback()
			return nil, fmt.Errorf("create %s: %w", directory, err)
		}
	}
	if err := workspace.ensureIgnored(); err != nil {
		_ = preparation.Rollback()
		return nil, err
	}
	return preparation, nil
}

// Commit makes prepared repository state permanent.
func (preparation *Preparation) Commit() {
	preparation.finalized = true
}

// Rollback restores the repository state captured before Prepare.
func (preparation *Preparation) Rollback() error {
	if preparation.finalized {
		return nil
	}
	var rollbackErrors []error
	if preparation.ignoreExisted {
		if err := os.WriteFile(preparation.ignorePath, preparation.ignoreData, 0o644); err != nil {
			rollbackErrors = append(rollbackErrors, err)
		}
	} else if err := os.Remove(preparation.ignorePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		rollbackErrors = append(rollbackErrors, err)
	}
	for index := len(preparation.createdDirs) - 1; index >= 0; index-- {
		if err := os.Remove(preparation.createdDirs[index]); err != nil && !errors.Is(err, os.ErrNotExist) {
			rollbackErrors = append(rollbackErrors, err)
		}
	}
	preparation.finalized = true
	return errors.Join(rollbackErrors...)
}

// Initialize creates the repository-local state directories and config.
func Initialize(root string, targets []string, activeTeam string) (Workspace, error) {
	workspace, err := New(root, targets, activeTeam)
	if err != nil {
		return Workspace{}, err
	}
	preparation, err := workspace.Prepare()
	if err != nil {
		return Workspace{}, err
	}
	if err := workspace.Save(); err != nil {
		_ = preparation.Rollback()
		return Workspace{}, err
	}
	preparation.Commit()
	return workspace, nil
}

// Save writes config atomically without changing authored team files.
func (workspace Workspace) Save() error {
	data, err := yaml.Marshal(workspace.Config)
	if err != nil {
		return fmt.Errorf("encode workspace config: %w", err)
	}
	return writeAtomic(workspace.ConfigPath(), data, 0o644)
}

// EnsureTeamState returns a state entry with defaults for team.
func (workspace *Workspace) EnsureTeamState(team model.Team) model.TeamState {
	if workspace.Config.TeamStates == nil {
		workspace.Config.TeamStates = map[string]model.TeamState{}
	}
	state, ok := workspace.Config.TeamStates[team.ID]
	if !ok {
		state = model.DefaultTeamState(team)
	}
	state = cloneTeamState(state)
	workspace.Config.TeamStates[team.ID] = state
	return cloneTeamState(state)
}

// SetTeamState persists a team-scoped state value in memory before Save.
func (workspace *Workspace) SetTeamState(teamID string, state model.TeamState) {
	if workspace.Config.TeamStates == nil {
		workspace.Config.TeamStates = map[string]model.TeamState{}
	}
	workspace.Config.TeamStates[teamID] = cloneTeamState(state)
}

// State returns a copy of state for a team, or the zero state if absent.
func (workspace Workspace) State(teamID string) model.TeamState {
	return cloneTeamState(workspace.Config.TeamStates[teamID])
}

func cloneTeamState(state model.TeamState) model.TeamState {
	return model.TeamState{
		ActiveWorkflow:  state.ActiveWorkflow,
		DisabledWorkers: append([]string{}, state.DisabledWorkers...),
		WorkerSkills:    cloneStringSliceMap(state.WorkerSkills),
		RemovedSkills:   cloneStringSliceMap(state.RemovedSkills),
		WorkerModels:    cloneStringMap(state.WorkerModels),
	}
}

func cloneStringSliceMap(values map[string][]string) map[string][]string {
	cloned := map[string][]string{}
	for key, value := range values {
		cloned[key] = append([]string{}, value...)
	}
	return cloned
}

func cloneStringMap(values map[string]string) map[string]string {
	cloned := map[string]string{}
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

// ConfigPath returns the local YAML state path.
func (workspace Workspace) ConfigPath() string {
	return filepath.Join(workspace.Root, MiezDir, ConfigFile)
}

// McpLocalPath returns the untracked, workspace-local file that holds
// user-supplied MCP tool configuration values (never committed).
func (workspace Workspace) McpLocalPath() string {
	return filepath.Join(workspace.Root, MiezDir, "mcp.local.yaml")
}

// LegacyManagedFiles reads the old compiler manifest for one-time cleanup of
// generated output created before distribution state moved into .miez. New
// renders never write this file.
func (workspace Workspace) LegacyManagedFiles() ([]string, error) {
	path := filepath.Join(workspace.Root, MiezDir, legacyManifestFile)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read legacy managed manifest: %w", err)
	}
	var manifest struct {
		Files []string `json:"files"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse legacy managed manifest: %w", err)
	}
	manifest.Files, err = normalizePaths(manifest.Files)
	if err != nil {
		return nil, err
	}
	return manifest.Files, nil
}

// LegacyMetadataPaths returns the old compiler metadata paths that a render
// transaction may remove after it has safely backed them up.
func (workspace Workspace) LegacyMetadataPaths() []string {
	return []string{
		filepath.ToSlash(filepath.Join(MiezDir, legacyManifestFile)),
		filepath.ToSlash(filepath.Join(MiezDir, legacyCompiledFile)),
	}
}

func (workspace Workspace) stateDirectories() []string {
	return []string{
		filepath.Join(workspace.Root, MiezDir),
	}
}

// ensureIgnored adds the repository-local paths whose values are not committed
// to .gitignore: MCP values and downloaded team source.
func (workspace Workspace) ensureIgnored() error {
	path := filepath.Join(workspace.Root, ".gitignore")
	wanted := []string{".miez/mcp.local.yaml", ".miez/miez_modules/"}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return os.WriteFile(path, []byte(strings.Join(wanted, "\n")+"\n"), 0o644)
	}
	if err != nil {
		return fmt.Errorf("read .gitignore: %w", err)
	}
	existing := map[string]struct{}{}
	for _, line := range strings.Split(string(data), "\n") {
		existing[strings.TrimSpace(line)] = struct{}{}
	}
	missing := []string{}
	for _, entry := range wanted {
		if _, ok := existing[entry]; !ok {
			missing = append(missing, entry)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	separator := ""
	if len(data) > 0 && data[len(data)-1] != '\n' {
		separator = "\n"
	}
	return os.WriteFile(path, append(data, []byte(separator+strings.Join(missing, "\n")+"\n")...), 0o644)
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".miez-tmp-")
	if err != nil {
		return fmt.Errorf("create temporary file for %s: %w", path, err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if err := temporary.Chmod(mode); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set temporary mode for %s: %w", path, err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary file for %s: %w", path, err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary file for %s: %w", path, err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return nil
}

func normalizeTargets(values []string) []string {
	seen := map[string]struct{}{}
	result := []string{}
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func normalizePaths(values []string) ([]string, error) {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return nil, errors.New("managed manifest contains an empty path")
		}
		cleanPath := filepath.Clean(filepath.FromSlash(value))
		if !filepath.IsLocal(cleanPath) || cleanPath == "." {
			return nil, fmt.Errorf("managed manifest path %q escapes the workspace", value)
		}
		clean := filepath.ToSlash(cleanPath)
		if clean != filepath.ToSlash(value) {
			return nil, fmt.Errorf("managed manifest path %q is not normalized", value)
		}
		if _, exists := seen[clean]; exists {
			return nil, fmt.Errorf("managed manifest contains duplicate path %q", value)
		}
		seen[clean] = struct{}{}
		result = append(result, clean)
	}
	sort.Strings(result)
	if result == nil {
		result = []string{}
	}
	return result, nil
}
