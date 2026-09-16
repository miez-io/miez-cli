// Package validate checks the miez-cli team artifact contract.
package validate

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/manuel/miez-cli/internal/artifacts"
	"github.com/manuel/miez-cli/internal/manifest"
	"github.com/manuel/miez-cli/internal/model"
	"github.com/manuel/miez-cli/internal/modules"
)

// Issue is one actionable blocking validation problem.
type Issue struct {
	Path    string
	Message string
}

// Package verifies that a generated team index still agrees with the
// package metadata it was built from.
func Package(root string, team model.Team) error {
	packageManifest, err := manifest.LoadTeamPackage(root)
	if err != nil {
		return err
	}
	metadataMatches := packageManifest.ID == team.ID &&
		packageManifest.Version == team.Version &&
		packageManifest.Name == team.Name &&
		packageManifest.Description == team.Description &&
		packageManifest.Author == team.Author &&
		packageManifest.DefaultModel == team.DefaultModel
	if !metadataMatches {
		return fmt.Errorf("generated %s metadata does not match %s", artifacts.TeamIndexFileName, manifest.FileName)
	}
	modelsMatch := len(packageManifest.Models) == len(team.Models) &&
		(len(packageManifest.Models) == 0 || reflect.DeepEqual(packageManifest.Models, team.Models))
	if !modelsMatch {
		return fmt.Errorf("generated %s models do not match %s", artifacts.TeamIndexFileName, manifest.FileName)
	}
	mcpMatches := len(packageManifest.MCP) == len(team.MCP) &&
		(len(packageManifest.MCP) == 0 || reflect.DeepEqual(packageManifest.MCP, team.MCP))
	if !mcpMatches {
		return fmt.Errorf("generated %s mcp declarations do not match %s", artifacts.TeamIndexFileName, manifest.FileName)
	}
	return nil
}

func (issue Issue) Error() string {
	return fmt.Sprintf("%s: %s", issue.Path, issue.Message)
}

// Result contains all validation issues instead of stopping at the first one.
type Result struct {
	Team   artifacts.TeamDir
	Issues []Issue
}

// Error renders the result as one error for command-line callers.
func (result Result) Error() error {
	if len(result.Issues) == 0 {
		return nil
	}
	lines := make([]string, 0, len(result.Issues))
	for _, issue := range result.Issues {
		lines = append(lines, issue.Error())
	}
	return fmt.Errorf("team validation failed:\n- %s", strings.Join(lines, "\n- "))
}

// Team validates the manifest and all referenced artifacts.
func Team(team artifacts.TeamDir) Result {
	result := Result{Team: team, Issues: []Issue{}}
	manifestPath := filepath.Join(team.Root, artifacts.TeamIndexFileName)

	checkIdentifier(&result, manifestPath, team.Team.ID, "team id")
	// Teams installed under .miez/miez_modules/<team-id> are looked up by
	// directory name, so the id and directory must agree. A standalone,
	// flat authoring repository has no such lookup and may live in an
	// arbitrarily named directory (e.g. the repository's own clone folder).
	if filepath.Base(filepath.Dir(team.Root)) == modules.Dir && team.Team.ID != filepath.Base(team.Root) {
		result.Issues = append(result.Issues, Issue{
			Path:    manifestPath,
			Message: fmt.Sprintf("id %q must match directory name %q", team.Team.ID, filepath.Base(team.Root)),
		})
	}
	if strings.TrimSpace(team.Team.Version) == "" {
		result.Issues = append(result.Issues, Issue{Path: manifestPath, Message: "version is required"})
	}
	if strings.TrimSpace(team.Team.Name) == "" {
		result.Issues = append(result.Issues, Issue{Path: manifestPath, Message: "name is required"})
	}
	if len(team.Team.Workflows) == 0 {
		result.Issues = append(result.Issues, Issue{Path: manifestPath, Message: "at least one workflow is required"})
	}

	declaredSkills := map[string]struct{}{}
	for index, skill := range team.Team.Skills {
		path := fmt.Sprintf("%s skills[%d]", manifestPath, index)
		checkIdentifier(&result, path, skill.ID, "skill id")
		if _, exists := declaredSkills[skill.ID]; exists {
			result.Issues = append(result.Issues, Issue{Path: path, Message: fmt.Sprintf("duplicate skill id %q", skill.ID)})
		}
		declaredSkills[skill.ID] = struct{}{}
		if strings.TrimSpace(skill.Path) == "" {
			result.Issues = append(result.Issues, Issue{Path: path, Message: "path is required"})
			continue
		}
		if strings.ToLower(filepath.Ext(skill.Path)) != ".md" {
			result.Issues = append(result.Issues, Issue{Path: path, Message: "path must reference a Markdown file"})
		}
		if _, err := team.ReadSkill(skill.ID); err != nil {
			result.Issues = append(result.Issues, Issue{Path: path, Message: err.Error()})
		}
	}

	declaredTasks := map[string]struct{}{}
	for index, task := range team.Team.Tasks {
		path := fmt.Sprintf("%s tasks[%d]", manifestPath, index)
		checkIdentifier(&result, path, task.ID, "task id")
		if _, exists := declaredTasks[task.ID]; exists {
			result.Issues = append(result.Issues, Issue{Path: path, Message: fmt.Sprintf("duplicate task id %q", task.ID)})
		}
		declaredTasks[task.ID] = struct{}{}
		if strings.TrimSpace(task.Path) == "" {
			result.Issues = append(result.Issues, Issue{Path: path, Message: "path is required"})
			continue
		}
		if strings.ToLower(filepath.Ext(task.Path)) != ".md" {
			result.Issues = append(result.Issues, Issue{Path: path, Message: "path must reference a Markdown file"})
		}
		if _, err := team.ReadTask(task); err != nil {
			result.Issues = append(result.Issues, Issue{Path: path, Message: err.Error()})
		}
	}

	declaredModels := map[string]struct{}{}
	for index, option := range team.Team.Models {
		path := fmt.Sprintf("%s models[%d]", manifestPath, index)
		if !model.IsValidModelID(option.ID) {
			result.Issues = append(result.Issues, Issue{Path: path, Message: fmt.Sprintf("model id %q is not valid", option.ID)})
		}
		if _, exists := declaredModels[option.ID]; exists {
			result.Issues = append(result.Issues, Issue{Path: path, Message: fmt.Sprintf("duplicate model id %q", option.ID)})
		}
		declaredModels[option.ID] = struct{}{}
		if strings.TrimSpace(option.Copilot) == "" {
			result.Issues = append(result.Issues, Issue{Path: path, Message: "model must set a copilot name"})
		}
	}
	if defaultModel := strings.TrimSpace(team.Team.DefaultModel); defaultModel != "" {
		if _, ok := model.FindModel(team.Team, defaultModel); !ok {
			result.Issues = append(result.Issues, Issue{
				Path:    manifestPath,
				Message: fmt.Sprintf("default_model %q is not supported; choose one of: %s", defaultModel, strings.Join(model.ModelIDs(team.Team), ", ")),
			})
		}
	}

	declaredTools := map[string]struct{}{}
	for index, server := range team.Team.MCP {
		path := fmt.Sprintf("%s mcp[%d]", manifestPath, index)
		checkIdentifier(&result, path, server.ID, "mcp id")
		if _, exists := declaredTools[server.ID]; exists {
			result.Issues = append(result.Issues, Issue{Path: path, Message: fmt.Sprintf("duplicate mcp id %q", server.ID)})
		}
		declaredTools[server.ID] = struct{}{}
	}

	workers := map[string]model.Worker{}
	for index, worker := range team.Team.Workers {
		path := fmt.Sprintf("%s workers[%d]", manifestPath, index)
		checkIdentifier(&result, path, worker.ID, "worker id")
		if _, exists := workers[worker.ID]; exists {
			result.Issues = append(result.Issues, Issue{Path: path, Message: fmt.Sprintf("duplicate worker id %q", worker.ID)})
		}
		workers[worker.ID] = worker
		if worker.Kind != model.WorkerAgent {
			result.Issues = append(result.Issues, Issue{Path: path, Message: fmt.Sprintf("kind %q is not supported; workers must use kind: agent", worker.Kind)})
		}
		effectiveModel := worker.EffectiveModel(model.TeamState{}, team.Team.DefaultModel)
		if _, ok := model.FindModel(team.Team, effectiveModel); !ok {
			result.Issues = append(result.Issues, Issue{
				Path:    path,
				Message: fmt.Sprintf("model %q is not supported; choose one of: %s", effectiveModel, strings.Join(model.ModelIDs(team.Team), ", ")),
			})
		}
		for _, toolID := range worker.Tools {
			if _, ok := declaredTools[toolID]; !ok {
				result.Issues = append(result.Issues, Issue{
					Path:    path,
					Message: fmt.Sprintf("worker %q references unknown mcp id %q", worker.ID, toolID),
				})
			}
		}
		if strings.TrimSpace(worker.Path) == "" {
			result.Issues = append(result.Issues, Issue{Path: path, Message: "path is required"})
			continue
		}
		if strings.ToLower(filepath.Ext(worker.Path)) != ".md" {
			result.Issues = append(result.Issues, Issue{Path: path, Message: "path must reference a Markdown file"})
		}
		if _, err := team.ReadWorker(worker); err != nil {
			result.Issues = append(result.Issues, Issue{Path: path, Message: err.Error()})
		}
		for _, skill := range worker.Skills {
			if !model.IsValidIdentifier(skill) {
				result.Issues = append(result.Issues, Issue{Path: path, Message: fmt.Sprintf("invalid skill id %q", skill)})
				continue
			}
			if _, err := team.ReadSkill(skill); err != nil {
				result.Issues = append(result.Issues, Issue{Path: path, Message: err.Error()})
			}
			if _, declared := declaredSkills[skill]; !declared {
				result.Issues = append(result.Issues, Issue{Path: path, Message: fmt.Sprintf("worker %q references undeclared skill %q", worker.ID, skill)})
			}
		}
	}

	workflows := map[string]struct{}{}
	for index, workflow := range team.Team.Workflows {
		path := fmt.Sprintf("%s workflows[%d]", manifestPath, index)
		checkIdentifier(&result, path, workflow.ID, "workflow id")
		if _, exists := workflows[workflow.ID]; exists {
			result.Issues = append(result.Issues, Issue{Path: path, Message: fmt.Sprintf("duplicate workflow id %q", workflow.ID)})
		}
		workflows[workflow.ID] = struct{}{}
		if strings.TrimSpace(workflow.Name) == "" {
			result.Issues = append(result.Issues, Issue{Path: path, Message: "name is required"})
		}
		if strings.TrimSpace(workflow.Path) == "" {
			result.Issues = append(result.Issues, Issue{Path: path, Message: "path is required"})
		} else if strings.ToLower(filepath.Ext(workflow.Path)) != ".md" {
			result.Issues = append(result.Issues, Issue{Path: path, Message: "path must reference a Markdown file"})
		} else if _, err := team.ReadWorkflow(workflow); err != nil {
			result.Issues = append(result.Issues, Issue{Path: path, Message: err.Error()})
		}
		if len(workflow.Phases) == 0 {
			result.Issues = append(result.Issues, Issue{Path: path, Message: "at least one phase is required"})
		}
		phaseIDs := map[string]struct{}{}
		for phaseIndex, phase := range workflow.Phases {
			phasePath := fmt.Sprintf("%s phases[%d]", path, phaseIndex)
			checkIdentifier(&result, phasePath, phase.ID, "phase id")
			if len(phase.Workers) == 0 {
				result.Issues = append(result.Issues, Issue{Path: phasePath, Message: "at least one worker is required"})
			}
			if _, exists := phaseIDs[phase.ID]; exists {
				result.Issues = append(result.Issues, Issue{Path: phasePath, Message: fmt.Sprintf("duplicate phase id %q", phase.ID)})
			}
			phaseIDs[phase.ID] = struct{}{}
			for _, workerID := range phase.Workers {
				if _, exists := workers[workerID]; !exists {
					result.Issues = append(result.Issues, Issue{
						Path:    phasePath,
						Message: fmt.Sprintf("unknown worker %q", workerID),
					})
				}
			}
		}
	}

	return result
}

func checkIdentifier(result *Result, path, value, label string) {
	if !model.IsValidIdentifier(value) {
		result.Issues = append(result.Issues, Issue{Path: path, Message: fmt.Sprintf("%s %q is not kebab-case", label, value)})
	}
}
