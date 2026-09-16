package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/manuel/miez-cli/internal/compile"
	"github.com/manuel/miez-cli/internal/validate"
	"github.com/manuel/miez-cli/internal/workspace"
)

func (app *App) newCheckCommand() *cobra.Command {
	var teamID string
	command := &cobra.Command{
		Use:   "check",
		Short: "validate installed team artifacts",
		Args:  noArgs,
		RunE: func(command *cobra.Command, args []string) error {
			if workspaceValue, err := workspace.Open(app.Cwd); err == nil {
				for _, target := range workspaceValue.Config.Targets {
					if strings.ToLower(strings.TrimSpace(target)) != compile.Target {
						return fmt.Errorf("unsupported target %q; GitHub Copilot (%q) is the only supported render target", target, compile.Target)
					}
				}
			}
			teams, err := app.teams.Check(teamID)
			if err != nil {
				return err
			}
			if len(teams) == 0 {
				return errors.New("no installed teams found")
			}
			for _, team := range teams {
				if err := validate.Package(team.Root, team.Team); err != nil {
					return err
				}
				if err := validate.Team(team).Error(); err != nil {
					return err
				}
				fmt.Fprintf(command.OutOrStdout(), "ok %s (%d workflow(s))\n", team.Team.ID, len(team.Team.Workflows))
			}
			return nil
		},
	}
	command.Flags().StringVar(&teamID, "team", "", "validate only this team")
	return command
}
