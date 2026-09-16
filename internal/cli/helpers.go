package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/manuel/miez-cli/internal/artifacts"
	"github.com/manuel/miez-cli/internal/model"
	"github.com/manuel/miez-cli/internal/workspace"
)

// activeTeam loads and validates the workspace's active team and its state.
func (app *App) activeTeam() (workspace.Workspace, artifacts.TeamDir, model.TeamState, error) {
	workspaceValue, err := workspace.Open(app.Cwd)
	if err != nil {
		return workspace.Workspace{}, artifacts.TeamDir{}, model.TeamState{}, err
	}
	team, state, err := app.teams.Resolve(workspaceValue)
	if err != nil {
		return workspace.Workspace{}, artifacts.TeamDir{}, model.TeamState{}, err
	}
	return workspaceValue, team, state, nil
}

// errPromptCanceled is returned by prompt when ctx is canceled, for example
// by Ctrl+C, while waiting for interactive input.
var errPromptCanceled = errors.New("canceled while waiting for input")

// prompt writes label and a default to app.Out, then reads one line from
// app.In, falling back to fallback when the line is blank or unreadable. It
// returns errPromptCanceled immediately if ctx is canceled, instead of
// blocking forever on a stdin read that cannot itself observe cancellation.
func (app *App) prompt(ctx context.Context, label, fallback string) (string, error) {
	fmt.Fprintf(app.Out, "%s [%s]: ", label, fallback)
	if app.input == nil {
		app.input = bufio.NewReader(app.In)
	}
	result := make(chan string, 1)
	go func() {
		line, err := app.input.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			result <- fallback
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			line = fallback
		}
		result <- line
	}()
	select {
	case <-ctx.Done():
		return "", errPromptCanceled
	case line := <-result:
		return line, nil
	}
}

// splitCSV lower-cases, trims, and de-duplicates a comma-separated list.
func splitCSV(value string) []string {
	seen := map[string]struct{}{}
	values := []string{}
	for _, part := range strings.Split(value, ",") {
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "" {
			continue
		}
		if _, exists := seen[part]; exists {
			continue
		}
		seen[part] = struct{}{}
		values = append(values, part)
	}
	return values
}
