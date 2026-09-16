package secrets_test

import (
	"path/filepath"
	"testing"

	"github.com/manuel/miez-cli/internal/model"
	"github.com/manuel/miez-cli/internal/secrets"
)

func TestParsePlaceholderAcceptsAllThreeSyntaxes(t *testing.T) {
	cases := map[string]string{
		"<GITHUB_TOKEN>":         "GITHUB_TOKEN",
		"${GITHUB_TOKEN}":        "GITHUB_TOKEN",
		"${env:GITHUB_TOKEN}":    "GITHUB_TOKEN",
		"${env:API_KEY_PRIMARY}": "API_KEY_PRIMARY",
	}
	for input, want := range cases {
		got, ok := secrets.ParsePlaceholder(input)
		if !ok || got != want {
			t.Errorf("ParsePlaceholder(%q) = (%q, %v), want (%q, true)", input, got, ok, want)
		}
	}
}

func TestParsePlaceholderRejectsNonPlaceholder(t *testing.T) {
	for _, input := range []string{"GITHUB_TOKEN", "", "<>", "${}", "plain-text"} {
		if _, ok := secrets.ParsePlaceholder(input); ok {
			t.Errorf("ParsePlaceholder(%q) accepted a non-placeholder value", input)
		}
	}
}

func TestIsMaskedMatchesTokenAndKeyCaseInsensitive(t *testing.T) {
	for _, name := range []string{"GITHUB_TOKEN", "api_key", "Secret_Key", "ACCESS_TOKEN"} {
		if !secrets.IsMasked(name) {
			t.Errorf("IsMasked(%q) = false, want true", name)
		}
	}
	if secrets.IsMasked("GITHUB_ORG") {
		t.Error("IsMasked(GITHUB_ORG) = true, want false")
	}
}

func TestRequiredForCollectsSortedRequirements(t *testing.T) {
	team := model.Team{MCP: []model.MCPServer{
		{ID: "github", Env: map[string]string{"GITHUB_TOKEN": "${env:GITHUB_TOKEN}", "GITHUB_ORG": "<GITHUB_ORG>"}},
	}}
	requirements, err := secrets.RequiredFor(team, []string{"github"})
	if err != nil {
		t.Fatal(err)
	}
	if len(requirements) != 2 {
		t.Fatalf("requirements = %#v, want 2", requirements)
	}
	if requirements[0].EnvKey != "GITHUB_ORG" || requirements[1].EnvKey != "GITHUB_TOKEN" {
		t.Fatalf("requirements not sorted: %#v", requirements)
	}
}

func TestRequiredForRejectsUnknownMcpID(t *testing.T) {
	if _, err := secrets.RequiredFor(model.Team{}, []string{"missing"}); err == nil {
		t.Fatal("RequiredFor accepted an unknown mcp id")
	}
}

func TestRequiredForRejectsNonPlaceholderEnvValue(t *testing.T) {
	team := model.Team{MCP: []model.MCPServer{
		{ID: "github", Env: map[string]string{"GITHUB_TOKEN": "not-a-placeholder"}},
	}}
	if _, err := secrets.RequiredFor(team, []string{"github"}); err == nil {
		t.Fatal("RequiredFor accepted a non-placeholder env value")
	}
}

func TestResolveUsesOverrideOrEnvWithoutPrompting(t *testing.T) {
	requirements := []secrets.Requirement{{ServerID: "github", EnvKey: "GITHUB_TOKEN", Variable: "GITHUB_TOKEN"}}
	resolver := secrets.Resolver{
		Overrides: map[string]string{"GITHUB_TOKEN": "from-override"},
		LookupEnv: func(string) (string, bool) { return "", false },
	}
	resolved, err := secrets.Resolve(requirements, resolver, true, func(string, bool) (string, error) {
		t.Fatal("prompt called despite a resolvable override")
		return "", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if resolved["GITHUB_TOKEN"] != "from-override" {
		t.Fatalf("resolved = %#v", resolved)
	}
}

func TestResolvePromptsInteractivelyWhenMissing(t *testing.T) {
	requirements := []secrets.Requirement{{ServerID: "github", EnvKey: "GITHUB_TOKEN", Variable: "GITHUB_TOKEN"}}
	resolver := secrets.Resolver{LookupEnv: func(string) (string, bool) { return "", false }}
	var promptedMasked bool
	resolved, err := secrets.Resolve(requirements, resolver, true, func(variable string, masked bool) (string, error) {
		promptedMasked = masked
		return "typed-value", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !promptedMasked {
		t.Fatal("prompt was not told to mask a token variable")
	}
	if resolved["GITHUB_TOKEN"] != "typed-value" {
		t.Fatalf("resolved = %#v", resolved)
	}
}

func TestResolveFailsClearlyWhenNonInteractiveAndMissing(t *testing.T) {
	requirements := []secrets.Requirement{{ServerID: "github", EnvKey: "GITHUB_TOKEN", Variable: "GITHUB_TOKEN"}}
	resolver := secrets.Resolver{LookupEnv: func(string) (string, bool) { return "", false }}
	_, err := secrets.Resolve(requirements, resolver, false, func(string, bool) (string, error) {
		t.Fatal("prompt called in non-interactive mode")
		return "", nil
	})
	if err == nil {
		t.Fatal("Resolve succeeded despite a missing required value in non-interactive mode")
	}
}

func TestStoreMergeRoundTripsAndNeverTouchesCommittedFiles(t *testing.T) {
	root := t.TempDir()
	store := secrets.Store{Path: filepath.Join(root, ".miez", "mcp.local.yaml")}
	if err := store.Merge(map[string]string{"GITHUB_TOKEN": "abc"}); err != nil {
		t.Fatal(err)
	}
	values, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if values["GITHUB_TOKEN"] != "abc" {
		t.Fatalf("values = %#v", values)
	}
	if err := store.Merge(map[string]string{"GITHUB_ORG": "acme"}); err != nil {
		t.Fatal(err)
	}
	values, err = store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if values["GITHUB_TOKEN"] != "abc" || values["GITHUB_ORG"] != "acme" {
		t.Fatalf("values after merge = %#v", values)
	}
}

func TestStoreLoadMissingReturnsEmpty(t *testing.T) {
	store := secrets.Store{Path: filepath.Join(t.TempDir(), "mcp.local.yaml")}
	values, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 0 {
		t.Fatalf("values = %#v, want empty", values)
	}
}
