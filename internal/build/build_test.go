package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/manuel/miez-cli/internal/artifacts"
)

func TestBuildGeneratesDeterministicCatalogFromFrontmatter(t *testing.T) {
	root := writePackage(t)

	first, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}
	if string(first.Data) != string(second.Data) {
		t.Fatal("repeated builds produced different generated index bytes")
	}
	if !strings.HasPrefix(string(first.Data), artifacts.GeneratedTeamIndexHeader) {
		t.Fatal("generated index is missing its do-not-edit header")
	}
	if first.Team.Workers[0].Description != "Builds implementation prompts." {
		t.Fatalf("worker description = %q", first.Team.Workers[0].Description)
	}
	if first.Team.Workers[0].Skills[0] != "writing" {
		t.Fatalf("worker skills = %#v", first.Team.Workers[0].Skills)
	}
	if first.Team.Workflows[0].Path != "workflows/default.md" {
		t.Fatalf("workflow path = %q", first.Team.Workflows[0].Path)
	}
	if !strings.Contains(string(first.Data), "path: workflows/default.md") {
		t.Fatal("generated catalog omitted workflow path")
	}
	if !strings.Contains(string(first.Data), "description: Builds implementation prompts.") {
		t.Fatal("generated catalog omitted worker description")
	}

	if err := Apply(root, first); err != nil {
		t.Fatal(err)
	}
	team, err := artifacts.LoadTeam(root)
	if err != nil {
		t.Fatal(err)
	}
	if team.Team.ID != "demo-team" || len(team.Team.Skills) != 1 {
		t.Fatalf("loaded team = %#v", team.Team)
	}
}

func TestBuildRejectsInvalidWorkflowReference(t *testing.T) {
	root := writePackage(t)
	workflowPath := filepath.Join(root, "workflows", "default.md")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(strings.Replace(string(data), "workers: [builder]", "workers: [missing]", 1))
	if err := os.WriteFile(workflowPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	existing := []byte("previous\n")
	if err := os.WriteFile(filepath.Join(root, artifacts.TeamIndexFileName), existing, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Build(root); err == nil {
		t.Fatal("Build accepted an unknown workflow worker")
	}
	got, err := os.ReadFile(filepath.Join(root, artifacts.TeamIndexFileName))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(existing) {
		t.Fatalf("existing generated index changed after failed build: %q", got)
	}
}

func TestBuildRejectsWorkerFrontmatterID(t *testing.T) {
	root := writePackage(t)
	workerPath := filepath.Join(root, "workers", "builder.md")
	data, err := os.ReadFile(workerPath)
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(strings.Replace(string(data), "---\nkind: command", "---\nid: starter\nkind: command", 1))
	if err := os.WriteFile(workerPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Build(root); err == nil ||
		!strings.Contains(err.Error(), `frontmatter field "id" is not allowed`) {
		t.Fatalf("Build error = %v, want stray frontmatter id rejection", err)
	}
}

func TestBuildRejectsSkillFrontmatterID(t *testing.T) {
	root := writePackage(t)
	skillPath := filepath.Join(root, "skills", "writing", "SKILL.md")
	if err := os.WriteFile(skillPath, []byte("---\nid: writing\n---\n# Writing\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Build(root); err == nil ||
		!strings.Contains(err.Error(), `frontmatter field "id" is not allowed`) {
		t.Fatalf("Build error = %v, want stray frontmatter id rejection", err)
	}
}

func writePackage(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"miez.yaml":               "id: demo-team\nversion: 1.0.0\nname: Demo team\nauthor: tests\n",
		"workers/builder.md":      "---\nkind: command\ndescription: Builds implementation prompts.\nskills: [writing]\n---\n# Builder\n",
		"skills/writing/SKILL.md": "---\n---\n# Writing\n",
		"workflows/default.md":    "---\nname: Default\nphases:\n  - id: build\n    workers: [builder]\n---\n# Default\n",
	}
	for relative, content := range files {
		path := filepath.Join(root, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
