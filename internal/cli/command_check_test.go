package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckCommandValidatesSingleTeam(t *testing.T) {
	root := t.TempDir()
	output := &bytes.Buffer{}
	app := newTestApp(t, output, &bytes.Buffer{}, root)
	if err := installFixtureTeam(t, app); err != nil {
		t.Fatal(err)
	}

	output.Reset()
	if err := app.Execute(context.Background(), []string{"check", "--team", "spec-driven-team"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "ok spec-driven-team") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestCheckCommandRejectsUnknownTeam(t *testing.T) {
	root := t.TempDir()
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, root)

	if err := app.Execute(context.Background(), []string{"check", "--team", "missing-team"}); err == nil {
		t.Fatal("check succeeded for an unknown team")
	}
}

func TestCheckCommandFailsWithNoInstalledTeams(t *testing.T) {
	root := t.TempDir()
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, root)

	if err := app.Execute(context.Background(), []string{"check"}); err == nil {
		t.Fatal("check succeeded with no installed teams")
	}
}

func TestCheckCommandRejectsNonCopilotWorkspaceTarget(t *testing.T) {
	root := t.TempDir()
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, root)
	if err := installFixtureTeam(t, app); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, ".miez", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.ReplaceAll(string(data), "copilot", "cursor")
	if err := os.WriteFile(configPath, []byte(tampered), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := app.Execute(context.Background(), []string{"check"}); err == nil {
		t.Fatal("check accepted a non-copilot workspace target")
	}
}
