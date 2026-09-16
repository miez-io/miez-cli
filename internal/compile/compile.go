// Package compile renders validated team artifacts for GitHub Copilot, the
// only supported render target in this phase.
package compile

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/manuel/miez-cli/internal/artifacts"
	"github.com/manuel/miez-cli/internal/model"
	"github.com/manuel/miez-cli/internal/validate"
	"github.com/manuel/miez-cli/internal/workspace"
)

// Target is the only render target this phase supports.
const Target = "copilot"

// Plan is a complete set of managed files ready to apply.
type Plan struct {
	Files   map[string][]byte
	Summary Summary
}

// Summary describes a completed in-memory compilation plan.
type Summary struct {
	ActiveTeam string   `json:"active_team"`
	Targets    []string `json:"targets"`
	Workflow   string   `json:"workflow"`
	Files      []string `json:"files"`
}

// Build validates and renders a team without changing the workspace.
func Build(workspaceValue workspace.Workspace, team artifacts.TeamDir, state model.TeamState) (Plan, error) {
	result := validate.Team(team)
	if err := result.Error(); err != nil {
		return Plan{}, err
	}
	for _, target := range workspaceValue.Config.Targets {
		if strings.ToLower(strings.TrimSpace(target)) != Target {
			return Plan{}, fmt.Errorf("unsupported target %q; GitHub Copilot (%q) is the only supported render target", target, Target)
		}
	}

	files := map[string][]byte{}
	activeWorkers := team.Team.EnabledWorkerIDs(state)
	for _, worker := range team.Team.Workers {
		if _, enabled := activeWorkers[worker.ID]; !enabled {
			continue
		}
		markdown, err := team.ReadWorker(worker)
		if err != nil {
			return Plan{}, err
		}
		body := markdown.Body
		effectiveSkills := worker.EffectiveSkills(state)
		for _, skill := range effectiveSkills {
			skillFiles, err := team.ReadSkillFiles(skill)
			if err != nil {
				return Plan{}, err
			}
			for relativePath, data := range skillFiles {
				if err := addFile(files, skillOutputPath(skill, relativePath), data); err != nil {
					return Plan{}, err
				}
			}
		}
		if len(effectiveSkills) > 0 {
			body += renderSkillLinks(worker, effectiveSkills)
		}
		if worker.Kind == model.WorkerAgent {
			modelID := worker.EffectiveModel(state, team.Team.DefaultModel)
			option, ok := model.FindModel(team.Team, modelID)
			if !ok {
				return Plan{}, fmt.Errorf("worker %q: unsupported model %q; choose one of: %s", worker.ID, modelID, strings.Join(model.ModelIDs(team.Team), ", "))
			}
			modelName, err := option.NameFor()
			if err != nil {
				return Plan{}, fmt.Errorf("worker %q: %w", worker.ID, err)
			}
			body = fmt.Sprintf("---\nmodel: %s\n---\n\n%s", modelName, body)
		}
		if err := addFile(files, workerPath(worker), []byte(body+"\n")); err != nil {
			return Plan{}, err
		}
	}

	rules, err := renderRules(team)
	if err != nil {
		return Plan{}, err
	}
	if rules != "" {
		if err := addFile(files, rulesPath(), []byte(rules)); err != nil {
			return Plan{}, err
		}
	}

	workflow, err := state.SelectedWorkflow(team.Team)
	if err != nil {
		return Plan{}, err
	}
	if !workflow.HasActiveWorker(activeWorkers) {
		return Plan{}, fmt.Errorf("workflow %q has no enabled workers", workflow.ID)
	}
	workflowMarkdown, err := team.ReadWorkflow(workflow)
	if err != nil {
		return Plan{}, err
	}
	if err := addFile(files, workflowPath(), []byte(renderWorkflow(team.Team, workflow, workflowMarkdown.Body, activeWorkers))); err != nil {
		return Plan{}, err
	}

	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, filepath.ToSlash(path))
	}
	sort.Strings(paths)
	return Plan{
		Files: files,
		Summary: Summary{
			ActiveTeam: team.Team.ID,
			Targets:    append([]string{}, workspaceValue.Config.Targets...),
			Workflow:   workflow.ID,
			Files:      paths,
		},
	}, nil
}

// Apply replaces the generated files in plan without a previous path set.
// Lifecycle operations should use ApplyTransaction with the previous active
// team's managed paths so obsolete output can be removed safely.
func Apply(workspaceValue workspace.Workspace, plan Plan) error {
	transaction, err := ApplyTransaction(workspaceValue, plan, nil)
	if err != nil {
		return err
	}
	return transaction.Commit()
}

// Transaction keeps the previous generated output available until the caller
// commits the surrounding configuration mutation.
type Transaction struct {
	workspace workspace.Workspace
	staging   string
	backups   []backupFile
	installed []string
	finalized bool
}

type backupFile struct {
	target string
	backup string
}

// ApplyTransaction stages and installs a complete generated output set. The
// previousPaths contain paths managed by the previous effective team view.
// The legacy .miez/manifest.json is also accepted for one-time cleanup, but
// no new compiler metadata is written. The caller must call Commit after
// related state has been persisted, or Rollback when that surrounding
// operation fails.
func ApplyTransaction(workspaceValue workspace.Workspace, plan Plan, previousPaths []string) (*Transaction, error) {
	legacyPaths, err := workspaceValue.LegacyManagedFiles()
	if err != nil {
		return nil, err
	}
	previousPaths = append(previousPaths, legacyPaths...)
	oldPaths := map[string]struct{}{}
	for _, relativePath := range uniqueSortedPaths(previousPaths) {
		path, err := safeWorkspacePath(workspaceValue.Root, relativePath)
		if err != nil {
			return nil, err
		}
		oldPaths[filepath.ToSlash(relativePath)] = struct{}{}
		if err := ensureNoSymlink(workspaceValue.Root, path); err != nil {
			return nil, err
		}
	}

	paths := make([]string, 0, len(plan.Files))
	for relativePath := range plan.Files {
		paths = append(paths, filepath.ToSlash(relativePath))
	}
	sort.Strings(paths)
	if err := validatePlanPaths(paths); err != nil {
		return nil, err
	}
	for _, relativePath := range paths {
		path, err := safeWorkspacePath(workspaceValue.Root, relativePath)
		if err != nil {
			return nil, err
		}
		if err := ensureNoSymlink(workspaceValue.Root, path); err != nil {
			return nil, err
		}
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("inspect generated file %s: %w", relativePath, err)
		}
		if _, managed := oldPaths[relativePath]; !managed {
			return nil, fmt.Errorf("refusing to overwrite unmanaged generated file %s", relativePath)
		}
		if info.IsDir() {
			return nil, fmt.Errorf("managed generated path %s is a directory", relativePath)
		}
	}

	staging, err := os.MkdirTemp(workspaceValue.Root, ".miez-compile-")
	if err != nil {
		return nil, fmt.Errorf("create compilation staging directory: %w", err)
	}
	transaction := &Transaction{workspace: workspaceValue, staging: staging}
	cleanupOnError := func(cause error) (*Transaction, error) {
		return nil, errors.Join(cause, transaction.Rollback())
	}
	for _, relativePath := range paths {
		if err := writeStagedFile(staging, relativePath, plan.Files[relativePath]); err != nil {
			return cleanupOnError(err)
		}
	}

	replacePaths := append([]string{}, previousPaths...)
	replacePaths = append(replacePaths, paths...)
	replacePaths = append(replacePaths, workspaceValue.LegacyMetadataPaths()...)
	replacePaths = uniqueSortedPaths(replacePaths)
	for _, relativePath := range replacePaths {
		path, err := safeWorkspacePath(workspaceValue.Root, relativePath)
		if err != nil {
			return cleanupOnError(err)
		}
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return cleanupOnError(fmt.Errorf("inspect existing generated file %s: %w", relativePath, err))
		}
		if info.IsDir() {
			return cleanupOnError(fmt.Errorf("existing generated path %s is a directory", relativePath))
		}
		backupPath := filepath.Join(staging, "backup", filepath.FromSlash(relativePath))
		if err := os.MkdirAll(filepath.Dir(backupPath), 0o755); err != nil {
			return cleanupOnError(fmt.Errorf("create backup directory for %s: %w", relativePath, err))
		}
		if err := os.Rename(path, backupPath); err != nil {
			return cleanupOnError(fmt.Errorf("stage existing generated file %s: %w", relativePath, err))
		}
		transaction.backups = append(transaction.backups, backupFile{target: path, backup: backupPath})
	}

	for _, relativePath := range paths {
		source := filepath.Join(staging, filepath.FromSlash(relativePath))
		target, err := safeWorkspacePath(workspaceValue.Root, relativePath)
		if err != nil {
			return cleanupOnError(err)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return cleanupOnError(fmt.Errorf("create generated directory for %s: %w", relativePath, err))
		}
		if err := os.Rename(source, target); err != nil {
			return cleanupOnError(fmt.Errorf("install generated file %s: %w", relativePath, err))
		}
		transaction.installed = append(transaction.installed, target)
	}
	return transaction, nil
}

// ManagedPaths returns the generated Copilot paths for an effective team
// view. Lifecycle mutations use it to remove output no longer owned by that
// view without a persistent compiler manifest.
func ManagedPaths(team artifacts.TeamDir, state model.TeamState) ([]string, error) {
	paths := []string{}
	activeWorkers := team.Team.EnabledWorkerIDs(state)
	for _, worker := range team.Team.Workers {
		if _, enabled := activeWorkers[worker.ID]; !enabled {
			continue
		}
		paths = append(paths, workerPath(worker))
		for _, skill := range worker.EffectiveSkills(state) {
			skillFiles, err := team.ReadSkillFiles(skill)
			if err != nil {
				return nil, err
			}
			for relativePath := range skillFiles {
				paths = append(paths, skillOutputPath(skill, relativePath))
			}
		}
	}
	rules, err := renderRules(team)
	if err != nil {
		return nil, err
	}
	if rules != "" {
		paths = append(paths, rulesPath())
	}
	if len(team.Team.Workflows) > 0 {
		paths = append(paths, workflowPath())
	}
	return uniqueSortedPaths(paths), nil
}

// Commit discards the previous generated output after the surrounding state is
// known to be persisted successfully.
func (transaction *Transaction) Commit() error {
	if transaction.finalized {
		return nil
	}
	transaction.finalized = true
	return os.RemoveAll(transaction.staging)
}

// Rollback restores the previous generated output and removes staged files.
func (transaction *Transaction) Rollback() error {
	if transaction.finalized {
		return nil
	}
	var rollbackErrors []error
	for index := len(transaction.installed) - 1; index >= 0; index-- {
		if err := os.Remove(transaction.installed[index]); err != nil && !errors.Is(err, os.ErrNotExist) {
			rollbackErrors = append(rollbackErrors, err)
		}
	}
	for index := len(transaction.backups) - 1; index >= 0; index-- {
		backup := transaction.backups[index]
		if err := os.MkdirAll(filepath.Dir(backup.target), 0o755); err != nil {
			rollbackErrors = append(rollbackErrors, err)
			continue
		}
		if err := os.Rename(backup.backup, backup.target); err != nil {
			rollbackErrors = append(rollbackErrors, err)
		}
	}
	if err := os.RemoveAll(transaction.staging); err != nil {
		rollbackErrors = append(rollbackErrors, err)
	}
	transaction.finalized = true
	return errors.Join(rollbackErrors...)
}

func workerPath(worker model.Worker) string {
	id := worker.ID
	if worker.Kind == model.WorkerAgent {
		return filepath.ToSlash(filepath.Join(".github", "agents", id+".md"))
	}
	return filepath.ToSlash(filepath.Join(".github", "prompts", id+".prompt.md"))
}

func workflowPath() string {
	return filepath.ToSlash(filepath.Join(".github", "instructions", "miez-workflow.instructions.md"))
}

// rulesPath is a second, independent always-on instruction, distinct from
// workflowPath: it renders whenever the team declares rules/ content.
func rulesPath() string {
	return filepath.ToSlash(filepath.Join(".github", "instructions", "miez-rules.instructions.md"))
}

func skillOutputPath(skillID, relativePath string) string {
	return filepath.ToSlash(filepath.Join(".github", "skills", skillID, filepath.FromSlash(relativePath)))
}

func renderSkillLinks(worker model.Worker, skills []string) string {
	var builder strings.Builder
	builder.WriteString("\n\n## Skills\n\n")
	builder.WriteString("Additionally to your persona, read every linked skill carefully, follow its instructions, and keep its specific know-how in context.\n\n")
	for _, skill := range skills {
		builder.WriteString(fmt.Sprintf("- [%s](%s)\n", skill, workerSkillPath(worker, skill)))
	}
	return builder.String()
}

func workerSkillPath(worker model.Worker, skillID string) string {
	return filepath.ToSlash(filepath.Join("..", "skills", skillID, "SKILL.md"))
}

func addFile(files map[string][]byte, path string, data []byte) error {
	path = filepath.ToSlash(path)
	if existing, exists := files[path]; exists && !bytes.Equal(existing, data) {
		return fmt.Errorf("generated path collision at %s", path)
	}
	files[path] = data
	return nil
}

// renderRules concatenates every Markdown file below team's rules/
// directory, sorted by path, into one always-on instruction body. It
// returns an empty string when the team declares no rules/ content.
func renderRules(team artifacts.TeamDir) (string, error) {
	rulesRoot := filepath.Join(team.Root, "rules")
	info, err := os.Stat(rulesRoot)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("inspect rules directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("rules path %s is not a directory", rulesRoot)
	}

	entries, err := os.ReadDir(rulesRoot)
	if err != nil {
		return "", fmt.Errorf("read rules directory: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".md" {
			continue
		}
		names = append(names, entry.Name())
	}
	if len(names) == 0 {
		return "", nil
	}
	sort.Strings(names)

	var builder strings.Builder
	builder.WriteString("---\napplyTo: \"**\"\n---\n\n# miez team rules\n\n")
	builder.WriteString("This instruction is generated by miez from the active team's rules/ folder. ")
	builder.WriteString("It is independent of workflow selection.\n")
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(rulesRoot, name))
		if err != nil {
			return "", fmt.Errorf("read rules file %s: %w", name, err)
		}
		builder.WriteString("\n")
		builder.Write(bytes.TrimSpace(data))
		builder.WriteString("\n")
	}
	return builder.String(), nil
}

func renderWorkflow(team model.Team, workflow model.Workflow, body string, active map[string]struct{}) string {
	var builder strings.Builder
	builder.WriteString("---\napplyTo: \"**\"\n---\n\n")
	builder.WriteString("# miez workflow routing\n\n")
	builder.WriteString("This instruction is generated by miez. Follow the active workflow phases in order.\n\n")
	builder.WriteString(fmt.Sprintf("Team: %s\nWorkflow: %s\n\n", team.Name, workflow.Name))
	if trimmedBody := strings.TrimSpace(body); trimmedBody != "" {
		builder.WriteString(trimmedBody)
		builder.WriteString("\n\n")
	}
	builder.WriteString("## Active workflow phases\n\n")
	phaseNumber := 0
	for _, phase := range workflow.Phases {
		workers := []string{}
		for _, worker := range phase.Workers {
			if _, enabled := active[worker]; enabled {
				workers = append(workers, worker)
			}
		}
		if len(workers) == 0 {
			continue
		}
		phaseNumber++
		builder.WriteString(fmt.Sprintf("%d. **%s**: %s\n", phaseNumber, phase.ID, strings.Join(workers, ", ")))
	}
	return builder.String()
}

func safeWorkspacePath(root, relativePath string) (string, error) {
	clean, err := safeArtifactPath(relativePath)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, clean), nil
}

func writeStagedFile(staging, relativePath string, data []byte) error {
	clean, err := safeArtifactPath(relativePath)
	if err != nil {
		return err
	}
	path := filepath.Join(staging, clean)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create staging directory for %s: %w", relativePath, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write staged file %s: %w", relativePath, err)
	}
	return nil
}

func uniqueSortedPaths(values []string) []string {
	seen := map[string]struct{}{}
	paths := make([]string, 0, len(values))
	for _, value := range values {
		value = filepath.ToSlash(value)
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		paths = append(paths, value)
	}
	sort.Strings(paths)
	return paths
}

func validatePlanPaths(paths []string) error {
	for _, path := range paths {
		if _, err := safeArtifactPath(path); err != nil {
			return err
		}
	}
	return nil
}

func ensureNoSymlink(root, target string) error {
	root = filepath.Clean(root)
	info, err := os.Lstat(root)
	if err != nil {
		return fmt.Errorf("inspect workspace root: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("workspace root %s must not be a symlink", root)
	}
	relative, err := filepath.Rel(root, target)
	if err != nil || !filepath.IsLocal(relative) {
		return fmt.Errorf("generated path %s escapes workspace", target)
	}
	current := root
	parts := strings.Split(relative, string(filepath.Separator))
	for index, part := range parts {
		if part == "." || part == "" {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("inspect generated path %s: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("generated path %s must not traverse a symlink", current)
		}
		if index < len(parts)-1 && !info.IsDir() {
			return fmt.Errorf("generated path parent %s is not a directory", current)
		}
	}
	return nil
}

func safeArtifactPath(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", errors.New("generated path must not be empty")
	}
	clean := filepath.Clean(filepath.FromSlash(value))
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("generated path %q escapes workspace", value)
	}
	return clean, nil
}
