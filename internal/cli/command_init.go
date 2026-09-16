package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/manuel/miez-cli/internal/workspace"
)

func (app *App) newTeamInstallCommand() *cobra.Command {
	var targets string

	command := &cobra.Command{
		Use:   "install <github-url>",
		Short: "install and activate a team from GitHub",
		Args:  exactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			ctx := command.Context()
			selectedTargets, err := app.selectInstallTargets(ctx, targets, command.Flags().Changed("targets"))
			if err != nil {
				return promptError(err)
			}
			teamID, err := app.teams.Install(ctx, selectedTargets, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(command.OutOrStdout(), "installed and active team: %s\n", teamID)
			return nil
		},
	}
	command.Flags().StringVar(&targets, "targets", "", "render target: copilot (prompted for on first install if omitted)")
	return command
}

// promptError maps an interactive prompt cancellation (Ctrl+C) to a process
// exit code, instead of the default failure code for a bad answer.
func promptError(err error) error {
	if errors.Is(err, errPromptCanceled) {
		return &ExitError{Code: 130, Err: err}
	}
	return err
}

func (app *App) selectTargets(ctx context.Context, value string, explicit bool) ([]string, error) {
	if !explicit && strings.TrimSpace(value) == "" {
		answer, err := app.prompt(ctx, "Render target (copilot)", "copilot")
		if err != nil {
			return nil, err
		}
		value = answer
	}
	values := splitCSV(value)
	if len(values) == 0 {
		return nil, errors.New("at least one target is required")
	}
	for _, target := range values {
		if target != "copilot" {
			return nil, fmt.Errorf("unsupported target %q; GitHub Copilot (\"copilot\") is the only supported render target", target)
		}
	}
	return values, nil
}

func (app *App) selectInstallTargets(ctx context.Context, value string, explicit bool) ([]string, error) {
	if explicit || strings.TrimSpace(value) != "" {
		return app.selectTargets(ctx, value, true)
	}
	workspaceValue, err := workspace.Open(app.Cwd)
	if err == nil {
		return append([]string{}, workspaceValue.Config.Targets...), nil
	}
	if !errors.Is(err, workspace.ErrNotInitialized) {
		return nil, err
	}
	return app.selectTargets(ctx, "", false)
}
