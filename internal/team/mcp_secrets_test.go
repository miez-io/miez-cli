package team

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/manuel/miez-cli/internal/secrets"
)

func mcpTeamFiles(id string) map[string]string {
	return map[string]string{
		id + "/miez.yaml":            "id: " + id + "\nversion: 1.0.0\nname: MCP team\nauthor: test\nmcp:\n  - id: github\n    env:\n      GITHUB_TOKEN: \"${env:GITHUB_TOKEN}\"\n",
		id + "/miez.generated.yaml":  "id: " + id + "\nversion: 1.0.0\nname: MCP team\nauthor: test\nmcp:\n  - id: github\n    env:\n      GITHUB_TOKEN: \"${env:GITHUB_TOKEN}\"\nworkers:\n  - id: builder\n    kind: agent\n    path: workers/builder.md\n    tools: [github]\nworkflows:\n  - id: default\n    name: Default\n    path: workflows/default.md\n    phases:\n      - id: build\n        workers: [builder]\n",
		id + "/workers/builder.md":   "---\nkind: agent\ntools: [github]\n---\n# Builder\n",
		id + "/workflows/default.md": "---\nname: Default\nphases:\n  - id: build\n    workers: [builder]\n---\n# Default\n",
	}
}

func TestUseFailsClearlyWhenMCPSecretMissingNonInteractive(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "mcp-team", "main", "sha1", mcpTeamFiles("mcp-team"))
	root := t.TempDir()
	initWorkspace(t, root)
	service := newTestService(root, fake, map[string]string{})

	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/mcp-team"); err == nil {
		t.Fatal("Use succeeded despite a missing required MCP secret")
	}
	if _, err := os.Stat(filepath.Join(root, ".miez", "miez_modules")); !os.IsNotExist(err) {
		t.Fatalf("miez_modules created despite failure: %v", err)
	}
}

func TestUseResolvesMCPSecretFromEnvWithoutPrompting(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "mcp-team", "main", "sha1", mcpTeamFiles("mcp-team"))
	root := t.TempDir()
	workspaceValue := initWorkspace(t, root)
	service := newTestService(root, fake, map[string]string{})
	service.Interactive = true
	service.PromptSecret = func(string, bool) (string, error) {
		t.Fatal("prompt called despite a resolvable env value")
		return "", nil
	}

	// GITHUB_TOKEN resolves via the process environment lookup passed to
	// resolveMCPSecrets (os.LookupEnv), not via credentials; set it for
	// real for this one test.
	t.Setenv("GITHUB_TOKEN", "from-process-env")

	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/mcp-team"); err != nil {
		t.Fatal(err)
	}
	store := secrets.Store{Path: workspaceValue.McpLocalPath()}
	values, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if values["GITHUB_TOKEN"] != "from-process-env" {
		t.Fatalf("stored values = %#v", values)
	}
}

func TestUsePromptsForMissingMCPSecretWhenInteractive(t *testing.T) {
	fake := newFakeGitHub(t)
	fake.setTeam("acme", "mcp-team", "main", "sha1", mcpTeamFiles("mcp-team"))
	root := t.TempDir()
	workspaceValue := initWorkspace(t, root)
	service := newTestService(root, fake, map[string]string{})
	service.Interactive = true
	var promptedMasked bool
	service.PromptSecret = func(variable string, masked bool) (string, error) {
		promptedMasked = masked
		return "typed-token", nil
	}

	if _, err := service.Install(context.Background(), []string{"copilot"}, "https://github.com/acme/mcp-team"); err != nil {
		t.Fatal(err)
	}
	if !promptedMasked {
		t.Fatal("prompt was not told to mask GITHUB_TOKEN")
	}
	store := secrets.Store{Path: workspaceValue.McpLocalPath()}
	values, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if values["GITHUB_TOKEN"] != "typed-token" {
		t.Fatalf("stored values = %#v", values)
	}
}
