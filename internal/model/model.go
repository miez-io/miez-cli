// Package model defines the file-backed artifact and workspace contracts.
package model

const (
	ConfigVersion = 1
	WorkerAgent   = "agent"
	WorkerCommand = "command"
)

// Config is the local state written below .miez.
type Config struct {
	Version    int                  `yaml:"version"`
	Targets    []string             `yaml:"targets"`
	ActiveTeam string               `yaml:"active_team,omitempty"`
	TeamStates map[string]TeamState `yaml:"teams,omitempty"`
}

// TeamState stores user choices without modifying authored team files.
type TeamState struct {
	ActiveWorkflow  string              `yaml:"active_workflow,omitempty"`
	DisabledWorkers []string            `yaml:"disabled_workers,omitempty"`
	WorkerSkills    map[string][]string `yaml:"worker_skills,omitempty"`
	RemovedSkills   map[string][]string `yaml:"removed_skills,omitempty"`
	WorkerModels    map[string]string   `yaml:"worker_models,omitempty"`
}

// Team describes one installable artifact bundle. Models lets the team add
// new models or correct a mapping in BuiltinModels itself, without waiting
// for a new miez release. MCP declares the tool servers workers may depend
// on; a worker references an entry here by id in its own Tools field.
type Team struct {
	ID           string        `yaml:"id"`
	Version      string        `yaml:"version"`
	Name         string        `yaml:"name"`
	Description  string        `yaml:"description,omitempty"`
	Author       string        `yaml:"author,omitempty"`
	DefaultModel string        `yaml:"default_model,omitempty"`
	Models       []ModelOption `yaml:"models,omitempty"`
	MCP          []MCPServer   `yaml:"mcp,omitempty"`
	Skills       []Skill       `yaml:"skills,omitempty"`
	Workers      []Worker      `yaml:"workers"`
	Workflows    []Workflow    `yaml:"workflows"`
}

// Skill points at one generated skill Markdown artifact.
type Skill struct {
	ID   string `yaml:"id"`
	Path string `yaml:"path"`
}

// MCPServer is one available MCP tool server a worker may depend on. Env
// maps the environment variable name the tool process expects to a
// placeholder expression (one of the three accepted syntaxes: "<VAR>",
// "${VAR}", "${env:VAR}") naming where miez sources its value from.
type MCPServer struct {
	ID  string            `yaml:"id"`
	Env map[string]string `yaml:"env,omitempty"`
}

// Worker points at a command or agent Markdown artifact. Model is only
// meaningful for kind: agent workers. Tools references ids declared in the
// owning team's MCP list.
type Worker struct {
	ID          string   `yaml:"id"`
	Kind        string   `yaml:"kind"`
	Path        string   `yaml:"path"`
	Description string   `yaml:"description,omitempty"`
	Model       string   `yaml:"model,omitempty"`
	Skills      []string `yaml:"skills,omitempty"`
	Tools       []string `yaml:"tools,omitempty"`
}

// EffectiveModel resolves the model id an agent worker uses: a workspace
// override set via `miez worker model set` (state), else its own model
// field, else defaultModel (typically the owning team's DefaultModel), else
// the package default.
func (worker Worker) EffectiveModel(state TeamState, defaultModel string) string {
	if override := state.WorkerModels[worker.ID]; override != "" {
		return override
	}
	if worker.Model != "" {
		return worker.Model
	}
	if defaultModel != "" {
		return defaultModel
	}
	return DefaultModelID
}

// Workflow is a named ordered process made from worker phases.
type Workflow struct {
	ID     string  `yaml:"id"`
	Name   string  `yaml:"name"`
	Path   string  `yaml:"path"`
	Phases []Phase `yaml:"phases"`
}

// Phase groups the workers that participate in one ordered step.
type Phase struct {
	ID      string   `yaml:"id"`
	Workers []string `yaml:"workers"`
}

// DefaultTeamState returns an independent zeroed state for a team.
func DefaultTeamState(team Team) TeamState {
	workflow := ""
	if len(team.Workflows) > 0 {
		workflow = team.Workflows[0].ID
	}

	return TeamState{
		ActiveWorkflow: workflow,
		WorkerSkills:   map[string][]string{},
		RemovedSkills:  map[string][]string{},
		WorkerModels:   map[string]string{},
	}
}
