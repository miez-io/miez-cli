package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/manuel/miez-cli/internal/artifacts"
)

func TestTeamAcceptsValidReferences(t *testing.T) {
	root := writeTeam(t, "valid-team", `
id: valid-team
version: 1.0.0
name: Valid team
workers:
  - id: builder
    kind: command
    path: commands/builder.md
    skills: [writing]
workflows:
  - id: default
    name: Default
    phases:
      - id: build
        workers: [builder]
`)

	team, err := artifacts.LoadTeam(root)
	if err != nil {
		t.Fatal(err)
	}
	result := Team(team)
	if err := result.Error(); err != nil {
		t.Fatal(err)
	}
}

func TestTeamAcceptsAgentWorkerWithValidModel(t *testing.T) {
	root := writeTeam(t, "agent-team", `
id: agent-team
version: 1.0.0
name: Agent team
workers:
  - id: builder
    kind: agent
    path: commands/builder.md
    model: claude-sonnet-4.5
workflows:
  - id: default
    name: Default
    phases:
      - id: build
        workers: [builder]
`)

	team, err := artifacts.LoadTeam(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Team(team).Error(); err != nil {
		t.Fatal(err)
	}
}

func TestTeamAcceptsAgentWorkerWithTeamDeclaredModel(t *testing.T) {
	root := writeTeam(t, "custom-model-team", `
id: custom-model-team
version: 1.0.0
name: Custom model team
models:
  - id: gpt-6
    copilot: GPT-6 (copilot)
workers:
  - id: builder
    kind: agent
    path: commands/builder.md
    model: gpt-6
workflows:
  - id: default
    name: Default
    phases:
      - id: build
        workers: [builder]
`)

	team, err := artifacts.LoadTeam(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Team(team).Error(); err != nil {
		t.Fatal(err)
	}
}

func TestTeamRejectsIncompleteTeamDeclaredModel(t *testing.T) {
	root := writeTeam(t, "incomplete-model-team", `
id: incomplete-model-team
version: 1.0.0
name: Incomplete model team
models:
  - id: gpt-6
workers:
  - id: builder
    kind: command
    path: commands/builder.md
workflows:
  - id: default
    name: Default
    phases:
      - id: build
        workers: [builder]
`)

	team, err := artifacts.LoadTeam(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Team(team).Error(); err == nil || !strings.Contains(err.Error(), "copilot name") {
		t.Fatalf("error = %v, want incomplete-model validation error", err)
	}
}

func TestTeamAcceptsAgentWorkerUsingDefaultModel(t *testing.T) {
	root := writeTeam(t, "default-model-team", `
id: default-model-team
version: 1.0.0
name: Default model team
default_model: gemini-3-pro
workers:
  - id: builder
    kind: agent
    path: commands/builder.md
workflows:
  - id: default
    name: Default
    phases:
      - id: build
        workers: [builder]
`)

	team, err := artifacts.LoadTeam(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Team(team).Error(); err != nil {
		t.Fatal(err)
	}
}

func TestTeamRejectsUnsupportedModel(t *testing.T) {
	root := writeTeam(t, "bad-model-team", `
id: bad-model-team
version: 1.0.0
name: Bad model team
workers:
  - id: builder
    kind: agent
    path: commands/builder.md
    model: gpt5.5
workflows:
  - id: default
    name: Default
    phases:
      - id: build
        workers: [builder]
`)

	team, err := artifacts.LoadTeam(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Team(team).Error(); err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("error = %v, want unsupported-model validation error", err)
	}
}

func TestTeamRejectsUnsupportedDefaultModel(t *testing.T) {
	root := writeTeam(t, "bad-default-model-team", `
id: bad-default-model-team
version: 1.0.0
name: Bad default model team
default_model: gpt5.5
workers:
  - id: builder
    kind: command
    path: commands/builder.md
workflows:
  - id: default
    name: Default
    phases:
      - id: build
        workers: [builder]
`)

	team, err := artifacts.LoadTeam(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Team(team).Error(); err == nil || !strings.Contains(err.Error(), "default_model") {
		t.Fatalf("error = %v, want unsupported default_model validation error", err)
	}
}

func TestTeamRejectsModelOnCommandWorker(t *testing.T) {
	root := writeTeam(t, "command-model-team", `
id: command-model-team
version: 1.0.0
name: Command model team
workers:
  - id: builder
    kind: command
    path: commands/builder.md
    model: gpt-5
workflows:
  - id: default
    name: Default
    phases:
      - id: build
        workers: [builder]
`)

	team, err := artifacts.LoadTeam(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Team(team).Error(); err == nil || !strings.Contains(err.Error(), "only valid for kind: agent") {
		t.Fatalf("error = %v, want model-on-command validation error", err)
	}
}

func TestTeamAcceptsWorkerToolReferencingDeclaredMcp(t *testing.T) {
	root := writeTeam(t, "mcp-team", `
id: mcp-team
version: 1.0.0
name: MCP team
mcp:
  - id: github
    env:
      GITHUB_TOKEN: "${env:GITHUB_TOKEN}"
workers:
  - id: builder
    kind: command
    path: commands/builder.md
    tools: [github]
workflows:
  - id: default
    name: Default
    phases:
      - id: build
        workers: [builder]
`)

	team, err := artifacts.LoadTeam(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Team(team).Error(); err != nil {
		t.Fatal(err)
	}
}

func TestTeamRejectsWorkerToolReferencingUnknownMcp(t *testing.T) {
	root := writeTeam(t, "bad-mcp-team", `
id: bad-mcp-team
version: 1.0.0
name: Bad MCP team
workers:
  - id: builder
    kind: command
    path: commands/builder.md
    tools: [unknown-tool]
workflows:
  - id: default
    name: Default
    phases:
      - id: build
        workers: [builder]
`)

	team, err := artifacts.LoadTeam(root)
	if err != nil {
		t.Fatal(err)
	}
	err = Team(team).Error()
	if err == nil || !strings.Contains(err.Error(), "builder") || !strings.Contains(err.Error(), "unknown-tool") {
		t.Fatalf("error = %v, want it to name the worker and the unresolved id", err)
	}
}

func TestTeamReportsMissingWorkerAndSkill(t *testing.T) {
	root := writeTeam(t, "broken-team", `
id: broken-team
version: 1.0.0
name: Broken team
workers:
  - id: builder
    kind: command
    path: commands/missing.md
    skills: [missing-skill]
workflows:
  - id: default
    name: Default
    phases:
      - id: build
        workers: [missing-worker]
`)

	team, err := artifacts.LoadTeam(root)
	if err != nil {
		t.Fatal(err)
	}
	result := Team(team)
	if len(result.Issues) < 3 {
		t.Fatalf("issues = %d, want at least 3", len(result.Issues))
	}
}

func TestTeamReportsDuplicatePhase(t *testing.T) {
	root := writeTeam(t, "duplicate-team", `
id: duplicate-team
version: 1.0.0
name: Duplicate team
workers:
  - id: builder
    kind: command
    path: commands/builder.md
workflows:
  - id: default
    name: Default
    phases:
      - id: build
        workers: [builder]
      - id: build
        workers: [builder]
`)

	team, err := artifacts.LoadTeam(root)
	if err != nil {
		t.Fatal(err)
	}
	result := Team(team)
	if len(result.Issues) != 1 || !strings.Contains(result.Issues[0].Message, "duplicate phase") {
		t.Fatalf("issues = %#v", result.Issues)
	}
}

func TestTeamRejectsEmptyWorkflow(t *testing.T) {
	root := writeTeam(t, "empty-workflow-team", `
id: empty-workflow-team
version: 1.0.0
name: Empty workflow team
workers:
  - id: builder
    kind: command
    path: commands/builder.md
workflows:
  - id: default
    name: Default
    phases: []
`)

	team, err := artifacts.LoadTeam(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Team(team).Error(); err == nil || !strings.Contains(err.Error(), "at least one phase") {
		t.Fatalf("error = %v, want empty-workflow validation error", err)
	}
}

func writeTeam(t *testing.T, name, manifest string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(filepath.Join(root, "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "skills", "writing"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest = strings.ReplaceAll(manifest, "    name: Default\n", "    name: Default\n    path: workflows/default.md\n")
	if strings.Contains(manifest, "skills: [writing]") {
		manifest = strings.Replace(manifest, "workers:\n", "skills:\n  - id: writing\n    path: skills/writing/SKILL.md\nworkers:\n", 1)
	}
	if err := os.WriteFile(filepath.Join(root, artifacts.TeamIndexFileName), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "commands", "builder.md"), []byte("---\nid: builder\n---\n# Builder\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "skills", "writing", "SKILL.md"), []byte("---\nid: writing\n---\n# Writing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "workflows"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "workflows", "default.md"), []byte("---\nid: default\nname: Default\nphases:\n  - id: build\n    workers: [builder]\n---\n# Default\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}
