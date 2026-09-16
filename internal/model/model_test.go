package model

import "testing"

func TestDefaultTeamStateSelectsFirstWorkflow(t *testing.T) {
	team := Team{
		Workflows: []Workflow{{ID: "default"}},
	}

	state := DefaultTeamState(team)

	if state.ActiveWorkflow != "default" {
		t.Fatalf("active workflow = %q, want %q", state.ActiveWorkflow, "default")
	}
	if state.WorkerSkills == nil {
		t.Fatal("worker skills map is nil")
	}
}

func TestDefaultTeamStateHandlesNoWorkflow(t *testing.T) {
	state := DefaultTeamState(Team{})

	if state.ActiveWorkflow != "" {
		t.Fatalf("active workflow = %q, want empty", state.ActiveWorkflow)
	}
}

func TestIsValidModelID(t *testing.T) {
	valid := []string{"gpt-5", "gpt-5.1", "gpt-5.2", "claude-sonnet-4.5", "claude-opus-4.5-thinking", "composer"}
	for _, value := range valid {
		if !IsValidModelID(value) {
			t.Errorf("IsValidModelID(%q) = false, want true", value)
		}
	}
	invalid := []string{"", "GPT-5", "gpt 5", "-gpt-5", "gpt-5-", "gpt..5"}
	for _, value := range invalid {
		if IsValidModelID(value) {
			t.Errorf("IsValidModelID(%q) = true, want false", value)
		}
	}
}

func TestWorkerEffectiveModelPrefersStateOverOwnOverTeamOverDefault(t *testing.T) {
	if got := (Worker{ID: "w", Model: "gpt-5.1"}).EffectiveModel(TeamState{}, "claude-sonnet-4.5"); got != "gpt-5.1" {
		t.Fatalf("EffectiveModel = %q, want worker override", got)
	}
	if got := (Worker{ID: "w"}).EffectiveModel(TeamState{}, "claude-sonnet-4.5"); got != "claude-sonnet-4.5" {
		t.Fatalf("EffectiveModel = %q, want team default", got)
	}
	if got := (Worker{ID: "w"}).EffectiveModel(TeamState{}, ""); got != DefaultModelID {
		t.Fatalf("EffectiveModel = %q, want package default %q", got, DefaultModelID)
	}
	state := TeamState{}
	state.SetWorkerModel("w", "composer")
	if got := (Worker{ID: "w", Model: "gpt-5.1"}).EffectiveModel(state, "claude-sonnet-4.5"); got != "composer" {
		t.Fatalf("EffectiveModel = %q, want the state override to win", got)
	}
}

func TestFindModel(t *testing.T) {
	if _, ok := FindModel(Team{}, "gpt-5"); !ok {
		t.Fatal("FindModel(gpt-5) = false, want true")
	}
	if _, ok := FindModel(Team{}, "not-a-model"); ok {
		t.Fatal("FindModel(not-a-model) = true, want false")
	}
}

func TestFindModelUsesTeamDeclaredModels(t *testing.T) {
	team := Team{Models: []ModelOption{
		{ID: "gpt-6", Copilot: "GPT-6 (copilot)"},
		{ID: "gpt-5", Copilot: "Custom GPT-5 (copilot)"},
	}}

	newModel, ok := FindModel(team, "gpt-6")
	if !ok {
		t.Fatal("FindModel did not find a team-declared model")
	}
	if newModel.Copilot != "GPT-6 (copilot)" {
		t.Fatalf("team-declared model = %#v", newModel)
	}

	overridden, ok := FindModel(team, "gpt-5")
	if !ok {
		t.Fatal("FindModel did not find the overridden built-in model")
	}
	if overridden.Copilot != "Custom GPT-5 (copilot)" {
		t.Fatalf("overridden model = %#v, want the team's mapping to win", overridden)
	}
}

func TestModelOptionNameForRejectsMissingCopilotName(t *testing.T) {
	blank := ModelOption{ID: "no-copilot-name"}
	if _, err := blank.NameFor(); err == nil {
		t.Fatal("NameFor succeeded for a model with no copilot name")
	}
	gpt5, ok := FindModel(Team{}, "gpt-5")
	if !ok {
		t.Fatal("FindModel(gpt-5) = false")
	}
	if _, err := gpt5.NameFor(); err != nil {
		t.Fatalf("NameFor() error = %v", err)
	}
}
