package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestNewCommandExposesFiveApplicationRoots(t *testing.T) {
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, t.TempDir())
	root := app.NewCommand()

	roots := map[string]bool{}
	for _, command := range root.Commands() {
		roots[command.Name()] = true
	}
	for _, name := range []string{"team", "workflow", "worker", "check"} {
		if !roots[name] {
			t.Errorf("missing root command %q", name)
		}
	}
	if roots["init"] {
		t.Error("root init command is still exposed")
	}
}

func TestTeamHelpListsAllSixSubcommands(t *testing.T) {
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, t.TempDir())
	root := app.NewCommand()
	for _, command := range root.Commands() {
		if command.Name() != "team" {
			continue
		}
		names := map[string]bool{}
		for _, sub := range command.Commands() {
			names[sub.Name()] = true
		}
		for _, want := range []string{"install", "list", "use", "bootstrap", "update", "outdated", "audit"} {
			if !names[want] {
				t.Errorf("team command missing subcommand %q", want)
			}
		}
		return
	}
	t.Fatal("team root command not found")
}

func TestUsageErrorsReturnExitCodeTwo(t *testing.T) {
	app := newTestApp(t, &bytes.Buffer{}, &bytes.Buffer{}, t.TempDir())

	err := app.Execute(context.Background(), []string{"--unknown"})
	var exitError *ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("error type = %T, want *ExitError", err)
	}
	if exitError.Code != 2 {
		t.Fatalf("exit code = %d, want 2", exitError.Code)
	}
}

func newTestApp(t *testing.T, out, errOut *bytes.Buffer, cwd string) *App {
	t.Helper()
	app, err := New(out, errOut, cwd)
	if err != nil {
		t.Fatal(err)
	}
	return app
}
