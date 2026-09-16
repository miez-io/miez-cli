package cli

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/manuel/miez-cli/internal/workspace"
)

func (app *App) newTeamCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "team",
		Short: "list, install, update, audit, or bootstrap teams",
	}
	command.AddCommand(
		app.newTeamInstallCommand(),
		app.newTeamListCommand(),
		app.newTeamUseCommand(),
		app.newTeamBootstrapCommand(),
		app.newTeamBuildCommand(),
		app.newTeamUpdateCommand(),
		app.newTeamOutdatedCommand(),
		app.newTeamAuditCommand(),
	)
	return command
}

func (app *App) newTeamBuildCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "build [path]",
		Short: "validate a team package and generate the team index",
		Args:  maxArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			root := app.Cwd
			if len(args) == 1 {
				root = args[0]
				if !filepath.IsAbs(root) {
					root = filepath.Join(app.Cwd, root)
				}
			}
			plan, err := app.teams.BuildPackage(root)
			if err != nil {
				return err
			}
			fmt.Fprintf(command.OutOrStdout(), "built team %q version %s\n", plan.Team.ID, plan.Team.Version)
			return nil
		},
	}
}

func (app *App) newTeamListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list installed teams",
		Args:  noArgs,
		RunE: func(command *cobra.Command, args []string) error {
			teams, err := app.teams.List()
			if err != nil {
				return err
			}
			active := ""
			if value, err := workspace.Open(app.Cwd); err == nil {
				active = value.Config.ActiveTeam
			}
			if len(teams) == 0 {
				fmt.Fprintln(command.OutOrStdout(), "no installed teams")
				return nil
			}
			sort.Slice(teams, func(left, right int) bool { return teams[left].Team.ID < teams[right].Team.ID })
			for _, team := range teams {
				marker := " "
				if team.Team.ID == active {
					marker = "*"
				}
				fmt.Fprintf(command.OutOrStdout(), "%s %s\t%s\n", marker, team.Team.ID, team.Team.Name)
			}
			return nil
		},
	}
}

func (app *App) newTeamUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "use <team-id>",
		Short: "activate an installed team",
		Args:  exactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			workspaceValue, err := workspace.Open(app.Cwd)
			if err != nil {
				return err
			}
			team, err := app.teams.Use(command.Context(), workspaceValue, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(command.OutOrStdout(), "active team: %s\n", team.Team.ID)
			return nil
		},
	}
}

func (app *App) newTeamBootstrapCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "bootstrap <team-id>",
		Short: "create a compatible team authoring skeleton",
		Args:  exactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := app.teams.Bootstrap(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(command.OutOrStdout(), "bootstrapped team %q\n", args[0])
			return nil
		},
	}
}

func (app *App) newTeamUpdateCommand() *cobra.Command {
	var yes bool
	var dryRun bool
	command := &cobra.Command{
		Use:   "update [team-id-or-url]",
		Short: "refresh an installed team to its latest matching ref",
		Args:  maxArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			ctx := command.Context()
			out := command.OutOrStdout()
			teamIDOrURL := ""
			if len(args) == 1 {
				teamIDOrURL = args[0]
			}
			plan, err := app.teams.PlanUpdate(ctx, teamIDOrURL)
			if err != nil {
				return err
			}
			defer plan.Close()
			if plan.UpToDate {
				fmt.Fprintf(out, "team %q is already up to date (%s)\n", plan.TeamID, plan.OldCommit)
				return nil
			}
			fmt.Fprintf(out, "team %q: %s -> %s\n", plan.TeamID, plan.OldCommit, plan.NewCommit)
			for _, path := range plan.AddedFiles {
				fmt.Fprintf(out, "  + %s\n", path)
			}
			for _, path := range plan.ChangedFiles {
				fmt.Fprintf(out, "  ~ %s\n", path)
			}
			for _, path := range plan.RemovedFiles {
				fmt.Fprintf(out, "  - %s\n", path)
			}
			if dryRun {
				fmt.Fprintln(out, "(dry run; nothing changed)")
				return nil
			}
			if !yes {
				answer, err := app.prompt(ctx, "Apply this update? (y/N)", "n")
				if err != nil {
					return promptError(err)
				}
				if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
					fmt.Fprintln(out, "no changes made")
					return nil
				}
			}
			workspaceValue, err := workspace.Open(app.Cwd)
			if err != nil {
				return err
			}
			if _, err := app.teams.ApplyUpdate(ctx, workspaceValue, plan); err != nil {
				return err
			}
			fmt.Fprintf(out, "team %q updated to %s\n", plan.TeamID, plan.NewCommit)
			return nil
		},
	}
	command.Flags().BoolVarP(&yes, "yes", "y", false, "apply the update without prompting")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "show the plan without changing anything")
	return command
}

func (app *App) newTeamOutdatedCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "outdated [team-id]",
		Short: "report whether installed teams have a newer matching ref upstream",
		Args:  maxArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			teamID := ""
			if len(args) == 1 {
				teamID = args[0]
			}
			reports, err := app.teams.Outdated(command.Context(), teamID)
			if err != nil {
				return err
			}
			out := command.OutOrStdout()
			if len(reports) == 0 {
				fmt.Fprintln(out, "no installed teams")
				return nil
			}
			for _, report := range reports {
				if report.UpToDate {
					fmt.Fprintf(out, "%s: up to date (%s)\n", report.TeamID, report.CurrentCommit)
					continue
				}
				fmt.Fprintf(out, "%s: %s -> %s available\n", report.TeamID, report.CurrentCommit, report.LatestCommit)
			}
			return nil
		},
	}
}

func (app *App) newTeamAuditCommand() *cobra.Command {
	var ci bool
	command := &cobra.Command{
		Use:   "audit [team-id]",
		Short: "detect drift between installed teams and the lockfile",
		Args:  maxArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			teamID := ""
			if len(args) == 1 {
				teamID = args[0]
			}
			reports, err := app.teams.Audit(teamID)
			if err != nil {
				return err
			}
			out := command.OutOrStdout()
			dirty := false
			for _, report := range reports {
				if report.Clean() {
					fmt.Fprintf(out, "%s: clean\n", report.TeamID)
					continue
				}
				dirty = true
				for _, finding := range report.Findings {
					fmt.Fprintf(out, "%s: %s %s\n", report.TeamID, finding.Status, finding.Path)
				}
			}
			if ci && dirty {
				return &ExitError{Code: 1, Err: errors.New("team audit found drift")}
			}
			return nil
		},
	}
	command.Flags().BoolVar(&ci, "ci", false, "exit non-zero when drift is found")
	return command
}
