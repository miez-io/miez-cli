package compile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/manuel/miez-cli/internal/artifacts"
	"github.com/manuel/miez-cli/internal/model"
	"github.com/manuel/miez-cli/internal/workspace"
)

func TestBuildRendersCopilotOutputAndLinkedSkills(t *testing.T) {
	root := t.TempDir()
	workspaceValue, err := workspace.Initialize(root, []string{"copilot"}, "demo-team")
	if err != nil {
		t.Fatal(err)
	}
	teamRoot := filepath.Join(root, ".miez", "miez_modules", "demo-team")
	writeCompileTeam(t, teamRoot)
	team, err := artifacts.LoadTeam(teamRoot)
	if err != nil {
		t.Fatal(err)
	}
	state := model.DefaultTeamState(team.Team)
	state.WorkerSkills["builder"] = []string{"extra"}

	plan, err := Build(workspaceValue, team, state)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := plan.Files[".github/agents/builder.md"]; !ok {
		t.Fatal("missing Copilot agent")
	}
	if taskOutput, ok := plan.Files[".github/prompts/task-analyze.prompt.md"]; !ok {
		t.Fatal("missing task prompt")
	} else if !strings.HasPrefix(string(taskOutput), "---\ndescription: Analyze the solution.\n---\n\n# Analyze\n") {
		t.Fatalf("task prompt = %q", taskOutput)
	}
	workerOutput := string(plan.Files[".github/agents/builder.md"])
	if !strings.Contains(workerOutput, "model: GPT-5 (copilot)\n") ||
		!strings.Contains(workerOutput, "description: Builds implementation prompts.\n") {
		t.Fatalf("prompt description frontmatter = %q", workerOutput)
	}
	if !strings.Contains(workerOutput, "## Skills") {
		t.Fatal("skills section is missing")
	}
	if !strings.Contains(workerOutput, "- [extra](../skills/extra/SKILL.md)") {
		t.Fatalf("skill link is missing: %q", workerOutput)
	}
	if strings.Contains(workerOutput, "# Extra") {
		t.Fatal("skill body was injected into worker output")
	}
	if _, ok := plan.Files[".github/skills/extra/SKILL.md"]; !ok {
		t.Fatal("missing Copilot skill")
	}
	if string(plan.Files[".github/skills/extra/assets/template.md"]) != "asset template\n" {
		t.Fatalf("missing skill asset: %q", plan.Files[".github/skills/extra/assets/template.md"])
	}
	if string(plan.Files[".github/skills/extra/references/guide.md"]) != "reference guide\n" {
		t.Fatalf("missing skill reference: %q", plan.Files[".github/skills/extra/references/guide.md"])
	}
	if !strings.Contains(string(plan.Files[".github/instructions/miez-workflow.instructions.md"]), "applyTo: \"**\"") {
		t.Fatal("workflow instruction is not always-on")
	}

	if err := Apply(workspaceValue, plan); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "agents", "builder.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "prompts", "task-analyze.prompt.md")); err != nil {
		t.Fatal(err)
	}
	for relative, expected := range map[string]string{
		".github/skills/extra/assets/template.md":  "asset template\n",
		".github/skills/extra/references/guide.md": "reference guide\n",
	} {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatalf("read installed skill file %s: %v", relative, err)
		}
		if string(data) != expected {
			t.Fatalf("installed skill file %s = %q, want %q", relative, data, expected)
		}
	}
	managedPaths, err := ManagedPaths(team, state)
	if err != nil {
		t.Fatal(err)
	}
	if !hasPath(managedPaths, ".github/skills/extra/assets/template.md") ||
		!hasPath(managedPaths, ".github/skills/extra/references/guide.md") {
		t.Fatalf("managed paths = %#v", managedPaths)
	}
	if !hasPath(managedPaths, ".github/prompts/task-analyze.prompt.md") {
		t.Fatalf("managed paths omitted task prompt: %#v", managedPaths)
	}
}

func TestBuildRejectsUnsupportedTargetBeforeWriting(t *testing.T) {
	root := t.TempDir()
	workspaceValue, err := workspace.Initialize(root, []string{"not-a-provider"}, "demo-team")
	if err != nil {
		t.Fatal(err)
	}
	teamRoot := filepath.Join(root, ".miez", "miez_modules", "demo-team")
	writeCompileTeam(t, teamRoot)
	team, err := artifacts.LoadTeam(teamRoot)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Build(workspaceValue, team, model.DefaultTeamState(team.Team)); err == nil {
		t.Fatal("Build succeeded for unsupported target")
	}
	if _, err := os.Stat(filepath.Join(root, ".github")); !os.IsNotExist(err) {
		t.Fatalf("generated target directory exists after failed build: %v", err)
	}
}

func TestApplyRejectsUnmanagedGeneratedFile(t *testing.T) {
	root := t.TempDir()
	workspaceValue, err := workspace.Initialize(root, []string{"copilot"}, "demo-team")
	if err != nil {
		t.Fatal(err)
	}
	teamRoot := filepath.Join(root, ".miez", "miez_modules", "demo-team")
	writeCompileTeam(t, teamRoot)
	team, err := artifacts.LoadTeam(teamRoot)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := Build(workspaceValue, team, model.DefaultTeamState(team.Team))
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, ".github", "agents", "builder.md")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("user-owned\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Apply(workspaceValue, plan); err == nil {
		t.Fatal("Apply overwrote an unmanaged generated file")
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "user-owned\n" {
		t.Fatalf("unmanaged file = %q", data)
	}
}

func TestApplyTransactionRollbackRestoresPreviousOutput(t *testing.T) {
	root := t.TempDir()
	workspaceValue, err := workspace.Initialize(root, []string{"copilot"}, "demo-team")
	if err != nil {
		t.Fatal(err)
	}
	first := Plan{
		Files: map[string][]byte{".github/prompts/demo.prompt.md": []byte("first\n")},
		Summary: Summary{
			ActiveTeam: "demo-team",
			Targets:    []string{"copilot"},
		},
	}
	if err := Apply(workspaceValue, first); err != nil {
		t.Fatal(err)
	}
	second := Plan{
		Files: map[string][]byte{".github/prompts/demo.prompt.md": []byte("second\n")},
		Summary: Summary{
			ActiveTeam: "demo-team",
			Targets:    []string{"copilot"},
		},
	}
	transaction, err := ApplyTransaction(workspaceValue, second, []string{".github/prompts/demo.prompt.md"})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".github", "prompts", "demo.prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "second\n" {
		t.Fatalf("staged output = %q", data)
	}
	if err := transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(filepath.Join(root, ".github", "prompts", "demo.prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "first\n" {
		t.Fatalf("rolled-back output = %q", data)
	}
}

func TestBuildHonorsRemovedSkills(t *testing.T) {
	root := t.TempDir()
	workspaceValue, err := workspace.Initialize(root, []string{"copilot"}, "demo-team")
	if err != nil {
		t.Fatal(err)
	}
	teamRoot := filepath.Join(root, ".miez", "miez_modules", "demo-team")
	writeCompileTeam(t, teamRoot)
	team, err := artifacts.LoadTeam(teamRoot)
	if err != nil {
		t.Fatal(err)
	}
	team.Team.Workers[0].Skills = []string{"extra"}
	state := model.DefaultTeamState(team.Team)
	state.RemovedSkills["builder"] = []string{"extra"}

	plan, err := Build(workspaceValue, team, state)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(plan.Files[".github/agents/builder.md"]), "Assigned skill: extra") {
		t.Fatal("removed skill was rendered")
	}
}

func TestBuildRendersAgentModelFrontmatter(t *testing.T) {
	root := t.TempDir()
	workspaceValue, err := workspace.Initialize(root, []string{"copilot"}, "agent-team")
	if err != nil {
		t.Fatal(err)
	}
	teamRoot := filepath.Join(root, ".miez", "miez_modules", "agent-team")
	writeCompileAgentTeam(t, teamRoot, "claude-sonnet-4.5")
	team, err := artifacts.LoadTeam(teamRoot)
	if err != nil {
		t.Fatal(err)
	}

	plan, err := Build(workspaceValue, team, model.DefaultTeamState(team.Team))
	if err != nil {
		t.Fatal(err)
	}
	workerOutput := string(plan.Files[".github/agents/architect.md"])
	if !strings.Contains(workerOutput, "model: Claude Sonnet 4.5 (copilot)\n") ||
		!strings.Contains(workerOutput, "description: Designs system architecture.\n") {
		t.Fatalf("agent frontmatter = %q", plan.Files[".github/agents/architect.md"])
	}
	if !strings.Contains(workerOutput, "description: Designs system architecture.\n") {
		t.Fatal("agent description is missing")
	}
}

func TestBuildPreservesNativeAgentModelSelector(t *testing.T) {
	root := t.TempDir()
	workspaceValue, err := workspace.Initialize(root, []string{"copilot"}, "agent-team")
	if err != nil {
		t.Fatal(err)
	}
	teamRoot := filepath.Join(root, ".miez", "miez_modules", "agent-team")
	writeCompileAgentTeam(t, teamRoot, "GPT-5.6 Luna (copilot)")
	team, err := artifacts.LoadTeam(teamRoot)
	if err != nil {
		t.Fatal(err)
	}

	plan, err := Build(workspaceValue, team, model.DefaultTeamState(team.Team))
	if err != nil {
		t.Fatal(err)
	}
	workerOutput := string(plan.Files[".github/agents/architect.md"])
	if !strings.Contains(workerOutput, "model: GPT-5.6 Luna (copilot)\n") ||
		!strings.Contains(workerOutput, "description: Designs system architecture.\n") {
		t.Fatalf("agent frontmatter = %q, want native selector preserved", workerOutput)
	}
}

func TestBuildCopiesProviderWorkerFrontmatterAndStripsMiezFields(t *testing.T) {
	root := t.TempDir()
	workspaceValue, err := workspace.Initialize(root, []string{"copilot"}, "agent-team")
	if err != nil {
		t.Fatal(err)
	}
	teamRoot := filepath.Join(root, ".miez", "miez_modules", "agent-team")
	writeCompileAgentTeam(t, teamRoot, "gpt-5.1")
	workerPath := filepath.Join(teamRoot, "commands", "architect.md")
	worker := `---
name: Source Architect
description: Source description wins.
model: gpt-5.1
reasoning-effort: xhigh
kind: agent
skills: [source-skill]
tools: [github]
---
# Architect
`
	if err := os.WriteFile(workerPath, []byte(worker), 0o644); err != nil {
		t.Fatal(err)
	}
	team, err := artifacts.LoadTeam(teamRoot)
	if err != nil {
		t.Fatal(err)
	}

	plan, err := Build(workspaceValue, team, model.DefaultTeamState(team.Team))
	if err != nil {
		t.Fatal(err)
	}
	output := string(plan.Files[".github/agents/architect.md"])
	for _, want := range []string{
		"name: Source Architect\n",
		"description: Source description wins.\n",
		"model: GPT-5.1 (copilot)\n",
		"reasoning-effort: xhigh\n",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("agent output = %q, missing provider frontmatter %q", output, want)
		}
	}
	for _, unwanted := range []string{
		"kind: agent\n",
		"skills: [source-skill]\n",
		"tools: [github]\n",
	} {
		if strings.Contains(output, unwanted) {
			t.Fatalf("agent output = %q, retained miez field %q", output, unwanted)
		}
	}
}

func TestBuildRendersTeamDeclaredModel(t *testing.T) {
	root := t.TempDir()
	workspaceValue, err := workspace.Initialize(root, []string{"copilot"}, "agent-team")
	if err != nil {
		t.Fatal(err)
	}
	teamRoot := filepath.Join(root, ".miez", "miez_modules", "agent-team")
	if err := os.MkdirAll(filepath.Join(teamRoot, "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `
id: agent-team
version: 1.0.0
name: Agent team
models:
  - id: gpt-6
    copilot: GPT-6 (copilot)
workers:
  - id: architect
    kind: agent
    path: commands/architect.md
    model: gpt-6
workflows: [{id: default, name: Default, path: workflows/default.md, phases: [{id: build, workers: [architect]}]}]
`
	if err := os.WriteFile(filepath.Join(teamRoot, artifacts.TeamIndexFileName), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(teamRoot, "commands", "architect.md"), []byte("---\n---\n# Architect\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(teamRoot, "workflows"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(teamRoot, "workflows", "default.md"), []byte("---\nname: Default\nphases:\n  - id: build\n    workers: [architect]\n---\n# Default\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	team, err := artifacts.LoadTeam(teamRoot)
	if err != nil {
		t.Fatal(err)
	}

	plan, err := Build(workspaceValue, team, model.DefaultTeamState(team.Team))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(plan.Files[".github/agents/architect.md"]), "---\nmodel: GPT-6 (copilot)\n---\n\n") {
		t.Fatalf("agent frontmatter = %q, want the team-declared model rendered", plan.Files[".github/agents/architect.md"])
	}
}

func TestBuildHonorsWorkerModelStateOverride(t *testing.T) {
	root := t.TempDir()
	workspaceValue, err := workspace.Initialize(root, []string{"copilot"}, "agent-team")
	if err != nil {
		t.Fatal(err)
	}
	teamRoot := filepath.Join(root, ".miez", "miez_modules", "agent-team")
	writeCompileAgentTeam(t, teamRoot, "claude-sonnet-4.5")
	team, err := artifacts.LoadTeam(teamRoot)
	if err != nil {
		t.Fatal(err)
	}
	state := model.DefaultTeamState(team.Team)
	state.SetWorkerModel("architect", "gpt-5.1")

	plan, err := Build(workspaceValue, team, state)
	if err != nil {
		t.Fatal(err)
	}
	workerOutput := string(plan.Files[".github/agents/architect.md"])
	if !strings.Contains(workerOutput, "model: GPT-5.1 (copilot)\n") ||
		!strings.Contains(workerOutput, "description: Designs system architecture.\n") {
		t.Fatalf("agent frontmatter = %q, want the state override to win over the generated index's model", workerOutput)
	}
}

func TestBuildRendersRulesAndWorkflowTogether(t *testing.T) {
	root := t.TempDir()
	workspaceValue, err := workspace.Initialize(root, []string{"copilot"}, "demo-team")
	if err != nil {
		t.Fatal(err)
	}
	teamRoot := filepath.Join(root, ".miez", "miez_modules", "demo-team")
	writeCompileTeam(t, teamRoot)
	if err := os.MkdirAll(filepath.Join(teamRoot, "rules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(teamRoot, "rules", "general.md"), []byte("Always answer in ASCII.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	team, err := artifacts.LoadTeam(teamRoot)
	if err != nil {
		t.Fatal(err)
	}

	plan, err := Build(workspaceValue, team, model.DefaultTeamState(team.Team))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(plan.Files[".github/instructions/miez-rules.instructions.md"]), "Always answer in ASCII.") {
		t.Fatal("rules were not rendered while workflow is disabled")
	}
	if _, ok := plan.Files[".github/instructions/miez-rules.instructions.md"]; !ok {
		t.Fatal("rules missing while workflow is active")
	}
	if _, ok := plan.Files[".github/instructions/miez-workflow.instructions.md"]; !ok {
		t.Fatal("workflow routing missing while active")
	}
}

func TestBuildOmitsRulesWhenTeamDeclaresNone(t *testing.T) {
	root := t.TempDir()
	workspaceValue, err := workspace.Initialize(root, []string{"copilot"}, "demo-team")
	if err != nil {
		t.Fatal(err)
	}
	teamRoot := filepath.Join(root, ".miez", "miez_modules", "demo-team")
	writeCompileTeam(t, teamRoot)
	team, err := artifacts.LoadTeam(teamRoot)
	if err != nil {
		t.Fatal(err)
	}

	// No rules directory: workflow remains active.
	plan, err := Build(workspaceValue, team, model.DefaultTeamState(team.Team))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := plan.Files[".github/instructions/miez-rules.instructions.md"]; ok {
		t.Fatal("rules instruction rendered with no rules/ directory")
	}
	if _, ok := plan.Files[".github/instructions/miez-workflow.instructions.md"]; !ok {
		t.Fatal("workflow routing missing with no rules directory")
	}
}

func hasPath(paths []string, wanted string) bool {
	for _, path := range paths {
		if path == wanted {
			return true
		}
	}
	return false
}

func writeCompileAgentTeam(t *testing.T, root, modelID string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := fmt.Sprintf(`
id: agent-team
version: 1.0.0
name: Agent team
workers:
  - id: architect
    kind: agent
    path: commands/architect.md
    model: %s
workflows: [{id: default, name: Default, path: workflows/default.md, phases: [{id: build, workers: [architect]}]}]
`, modelID)
	manifest = strings.Replace(manifest, "    path: commands/architect.md\n", "    path: commands/architect.md\n    description: Designs system architecture.\n", 1)
	if err := os.WriteFile(filepath.Join(root, artifacts.TeamIndexFileName), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "commands", "architect.md"), []byte("---\n---\n# Architect\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "workflows"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "workflows", "default.md"), []byte("---\nname: Default\nphases:\n  - id: build\n    workers: [architect]\n---\n# Default\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeCompileTeam(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "skills", "extra"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `
id: demo-team
version: 1.0.0
name: Demo team
skills:
  - id: extra
    path: skills/extra/SKILL.md
tasks:
  - id: analyze
    path: tasks/analyze.md
    description: Analyze the solution.
workers:
  - id: builder
    kind: agent
    path: commands/builder.md
    skills: []
workflows: [{id: default, name: Default, path: workflows/default.md, phases: [{id: build, workers: [builder]}]}]
`
	manifest = strings.Replace(manifest, "    path: commands/builder.md\n", "    path: commands/builder.md\n    description: Builds implementation prompts.\n", 1)
	if err := os.WriteFile(filepath.Join(root, artifacts.TeamIndexFileName), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "commands", "builder.md"), []byte("---\n---\n# Builder\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "skills", "extra", "SKILL.md"), []byte("---\n---\n# Extra\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "tasks", "analyze.md"), []byte("---\ndescription: Analyze the solution.\n---\n# Analyze\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for relative, content := range map[string]string{
		"assets/template.md":  "asset template\n",
		"references/guide.md": "reference guide\n",
	} {
		path := filepath.Join(root, "skills", "extra", filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "workflows"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "workflows", "default.md"), []byte("---\nname: Default\nphases:\n  - id: build\n    workers: [builder]\n---\n# Default\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
