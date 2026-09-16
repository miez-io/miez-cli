package model

import (
	"fmt"
	"sort"
)

// HasWorkflow reports whether team defines workflowID.
func (team Team) HasWorkflow(workflowID string) bool {
	for _, workflow := range team.Workflows {
		if workflow.ID == workflowID {
			return true
		}
	}
	return false
}

// Worker returns the worker with workerID, if team defines one.
func (team Team) Worker(workerID string) (Worker, bool) {
	for _, worker := range team.Workers {
		if worker.ID == workerID {
			return worker, true
		}
	}
	return Worker{}, false
}

// Skill returns the skill with skillID, if team defines one.
func (team Team) Skill(skillID string) (Skill, bool) {
	for _, skill := range team.Skills {
		if skill.ID == skillID {
			return skill, true
		}
	}
	return Skill{}, false
}

// EnabledWorkerIDs returns every team worker id not disabled by state.
func (team Team) EnabledWorkerIDs(state TeamState) map[string]struct{} {
	disabled := map[string]struct{}{}
	for _, workerID := range state.DisabledWorkers {
		disabled[workerID] = struct{}{}
	}
	enabled := map[string]struct{}{}
	for _, worker := range team.Workers {
		if _, blocked := disabled[worker.ID]; !blocked {
			enabled[worker.ID] = struct{}{}
		}
	}
	return enabled
}

// SelectedWorkflow resolves the workflow named by state, defaulting to
// team's first workflow when state has none selected yet.
func (state *TeamState) SelectedWorkflow(team Team) (Workflow, error) {
	workflowID := state.ActiveWorkflow
	if workflowID == "" && len(team.Workflows) > 0 {
		workflowID = team.Workflows[0].ID
	}
	for _, workflow := range team.Workflows {
		if workflow.ID == workflowID {
			return workflow, nil
		}
	}
	return Workflow{}, fmt.Errorf("workflow %q is not defined by team %q", workflowID, team.ID)
}

// ActiveWorkflowUsesWorker reports whether the selected workflow assigns
// workerID to any phase.
func (state *TeamState) ActiveWorkflowUsesWorker(team Team, workerID string) bool {
	workflow, err := state.SelectedWorkflow(team)
	if err != nil {
		return false
	}
	for _, phase := range workflow.Phases {
		for _, worker := range phase.Workers {
			if worker == workerID {
				return true
			}
		}
	}
	return false
}

// WorkflowWorkerIDs returns the enabled workers assigned to the selected
// workflow's phases.
func (state *TeamState) WorkflowWorkerIDs(team Team) map[string]struct{} {
	workflow, err := state.SelectedWorkflow(team)
	if err != nil {
		return map[string]struct{}{}
	}
	disabled := map[string]struct{}{}
	for _, workerID := range state.DisabledWorkers {
		disabled[workerID] = struct{}{}
	}
	active := map[string]struct{}{}
	for _, phase := range workflow.Phases {
		for _, worker := range phase.Workers {
			if _, isDisabled := disabled[worker]; !isDisabled {
				active[worker] = struct{}{}
			}
		}
	}
	return active
}

// SetWorkerDisabled adds or removes workerID from the workflow's disabled set.
func (state *TeamState) SetWorkerDisabled(workerID string, disabled bool) {
	state.DisabledWorkers = toggleValue(state.DisabledWorkers, workerID, disabled)
}

// SetWorkerSkill assigns or removes skillID for workerID, recording an
// explicit removal even when the team's default skills include it.
func (state *TeamState) SetWorkerSkill(workerID, skillID string, enabled bool) {
	if state.WorkerSkills == nil {
		state.WorkerSkills = map[string][]string{}
	}
	if state.RemovedSkills == nil {
		state.RemovedSkills = map[string][]string{}
	}
	state.WorkerSkills[workerID] = toggleValue(state.WorkerSkills[workerID], skillID, enabled)
	state.RemovedSkills[workerID] = toggleValue(state.RemovedSkills[workerID], skillID, !enabled)
}

// SetWorkerModel overrides workerID's effective model (see
// Worker.EffectiveModel), letting a user switch an agent's model per
// workspace, at any time, without editing the generated team index.
func (state *TeamState) SetWorkerModel(workerID, modelID string) {
	if state.WorkerModels == nil {
		state.WorkerModels = map[string]string{}
	}
	state.WorkerModels[workerID] = modelID
}

// EffectiveSkills returns worker's default skills plus state overrides, minus
// explicit removals, without duplicates.
func (worker Worker) EffectiveSkills(state TeamState) []string {
	removed := map[string]struct{}{}
	for _, skill := range state.RemovedSkills[worker.ID] {
		removed[skill] = struct{}{}
	}
	seen := map[string]struct{}{}
	skills := []string{}
	for _, skill := range append(append([]string{}, worker.Skills...), state.WorkerSkills[worker.ID]...) {
		if _, isRemoved := removed[skill]; isRemoved {
			continue
		}
		if _, exists := seen[skill]; exists {
			continue
		}
		seen[skill] = struct{}{}
		skills = append(skills, skill)
	}
	return skills
}

// HasActiveWorker reports whether any phase in workflow assigns a worker
// present in active.
func (workflow Workflow) HasActiveWorker(active map[string]struct{}) bool {
	for _, phase := range workflow.Phases {
		for _, worker := range phase.Workers {
			if _, enabled := active[worker]; enabled {
				return true
			}
		}
	}
	return false
}

func toggleValue(values []string, value string, enabled bool) []string {
	result := []string{}
	for _, current := range values {
		if current != value {
			result = append(result, current)
		}
	}
	if enabled {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
