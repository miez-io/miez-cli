package model

import (
	"reflect"
	"testing"
)

func testTeam() Team {
	return Team{
		ID: "demo-team",
		Workers: []Worker{
			{ID: "builder", Skills: []string{"writing"}},
			{ID: "reviewer"},
		},
		Workflows: []Workflow{
			{
				ID: "default",
				Phases: []Phase{
					{ID: "build", Workers: []string{"builder"}},
					{ID: "review", Workers: []string{"reviewer"}},
				},
			},
			{ID: "alternate"},
		},
	}
}

func TestTeamHasWorkflow(t *testing.T) {
	team := testTeam()

	if !team.HasWorkflow("default") {
		t.Error("HasWorkflow(default) = false, want true")
	}
	if team.HasWorkflow("missing") {
		t.Error("HasWorkflow(missing) = true, want false")
	}
}

func TestTeamWorker(t *testing.T) {
	team := testTeam()

	worker, ok := team.Worker("builder")
	if !ok || worker.ID != "builder" {
		t.Fatalf("Worker(builder) = %+v, %v", worker, ok)
	}
	if _, ok := team.Worker("missing"); ok {
		t.Error("Worker(missing) = true, want false")
	}
}

func TestTeamEnabledWorkerIDs(t *testing.T) {
	team := testTeam()
	state := TeamState{DisabledWorkers: []string{"reviewer"}}

	enabled := team.EnabledWorkerIDs(state)

	if _, ok := enabled["builder"]; !ok {
		t.Error("builder should be enabled")
	}
	if _, ok := enabled["reviewer"]; ok {
		t.Error("reviewer should be disabled")
	}
}

func TestTeamStateSelectedWorkflowDefaultsToFirst(t *testing.T) {
	team := testTeam()
	state := TeamState{}

	workflow, err := state.SelectedWorkflow(team)
	if err != nil {
		t.Fatal(err)
	}
	if workflow.ID != "default" {
		t.Fatalf("workflow id = %q, want %q", workflow.ID, "default")
	}
}

func TestTeamStateSelectedWorkflowHonorsActiveWorkflow(t *testing.T) {
	team := testTeam()
	state := TeamState{ActiveWorkflow: "alternate"}

	workflow, err := state.SelectedWorkflow(team)
	if err != nil {
		t.Fatal(err)
	}
	if workflow.ID != "alternate" {
		t.Fatalf("workflow id = %q, want %q", workflow.ID, "alternate")
	}
}

func TestTeamStateSelectedWorkflowRejectsUnknownID(t *testing.T) {
	team := testTeam()
	state := TeamState{ActiveWorkflow: "missing"}

	if _, err := state.SelectedWorkflow(team); err == nil {
		t.Fatal("SelectedWorkflow succeeded for an unknown workflow id")
	}
}

func TestTeamStateActiveWorkflowUsesWorker(t *testing.T) {
	team := testTeam()
	state := TeamState{}

	if !state.ActiveWorkflowUsesWorker(team, "builder") {
		t.Error("expected default workflow to use builder")
	}
	if state.ActiveWorkflowUsesWorker(team, "unknown-worker") {
		t.Error("expected default workflow to not use unknown-worker")
	}
}

func TestTeamStateWorkflowWorkerIDsExcludesDisabled(t *testing.T) {
	team := testTeam()
	state := TeamState{DisabledWorkers: []string{"reviewer"}}

	active := state.WorkflowWorkerIDs(team)

	if _, ok := active["builder"]; !ok {
		t.Error("builder should be active")
	}
	if _, ok := active["reviewer"]; ok {
		t.Error("reviewer should be excluded")
	}
}

func TestTeamStateSetWorkerDisabledToggles(t *testing.T) {
	state := TeamState{}

	state.SetWorkerDisabled("reviewer", true)
	if !reflect.DeepEqual(state.DisabledWorkers, []string{"reviewer"}) {
		t.Fatalf("disabled workers = %#v", state.DisabledWorkers)
	}

	state.SetWorkerDisabled("reviewer", false)
	if len(state.DisabledWorkers) != 0 {
		t.Fatalf("disabled workers = %#v, want empty", state.DisabledWorkers)
	}
}

func TestTeamStateSetWorkerSkillAddAndRemove(t *testing.T) {
	state := TeamState{}

	state.SetWorkerSkill("builder", "extra", true)
	if !reflect.DeepEqual(state.WorkerSkills["builder"], []string{"extra"}) {
		t.Fatalf("worker skills = %#v", state.WorkerSkills["builder"])
	}
	if len(state.RemovedSkills["builder"]) != 0 {
		t.Fatalf("removed skills = %#v, want empty", state.RemovedSkills["builder"])
	}

	state.SetWorkerSkill("builder", "extra", false)
	if len(state.WorkerSkills["builder"]) != 0 {
		t.Fatalf("worker skills = %#v, want empty", state.WorkerSkills["builder"])
	}
	if !reflect.DeepEqual(state.RemovedSkills["builder"], []string{"extra"}) {
		t.Fatalf("removed skills = %#v", state.RemovedSkills["builder"])
	}
}

func TestTeamStateSetWorkerModel(t *testing.T) {
	state := TeamState{}

	state.SetWorkerModel("architect", "claude-opus-4.5")
	if state.WorkerModels["architect"] != "claude-opus-4.5" {
		t.Fatalf("worker models = %#v", state.WorkerModels)
	}

	state.SetWorkerModel("architect", "gpt-5.1")
	if state.WorkerModels["architect"] != "gpt-5.1" {
		t.Fatalf("worker models = %#v, want the later value to win", state.WorkerModels)
	}
}

func TestWorkerEffectiveSkillsHonorsRemoval(t *testing.T) {
	worker := Worker{ID: "builder", Skills: []string{"writing"}}
	state := TeamState{}
	state.SetWorkerSkill("builder", "writing", false)

	skills := worker.EffectiveSkills(state)

	if len(skills) != 0 {
		t.Fatalf("skills = %#v, want empty after removal", skills)
	}
}

func TestWorkerEffectiveSkillsMergesStateAdditions(t *testing.T) {
	worker := Worker{ID: "builder", Skills: []string{"writing"}}
	state := TeamState{}
	state.SetWorkerSkill("builder", "review", true)

	skills := worker.EffectiveSkills(state)

	want := []string{"writing", "review"}
	if !reflect.DeepEqual(skills, want) {
		t.Fatalf("skills = %#v, want %#v", skills, want)
	}
}

func TestWorkflowHasActiveWorker(t *testing.T) {
	workflow := Workflow{Phases: []Phase{{Workers: []string{"builder"}}}}

	if !workflow.HasActiveWorker(map[string]struct{}{"builder": {}}) {
		t.Error("expected workflow to have an active worker")
	}
	if workflow.HasActiveWorker(map[string]struct{}{"other": {}}) {
		t.Error("expected workflow to have no active worker")
	}
}
