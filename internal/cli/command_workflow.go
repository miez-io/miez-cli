package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func (app *App) newWorkflowCommand() *cobra.Command {
	command := &cobra.Command{Use: "workflow", Short: "configure the active workflow"}
	command.AddCommand(
		app.newWorkflowUseCommand(),
		app.newWorkflowWorkerCommand(),
	)
	return command
}

func (app *App) newWorkflowUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "use <workflow-id>",
		Short: "select the active team's workflow",
		Args:  exactArgs(1),
		ValidArgsFunction: func(command *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			_, team, _, err := app.activeTeam()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			ids := make([]string, 0, len(team.Team.Workflows))
			for _, workflow := range team.Team.Workflows {
				ids = append(ids, workflow.ID)
			}
			return ids, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(command *cobra.Command, args []string) error {
			workspaceValue, team, state, err := app.activeTeam()
			if err != nil {
				return err
			}
			if !team.Team.HasWorkflow(args[0]) {
				return fmt.Errorf("team %q has no workflow %q", team.Team.ID, args[0])
			}
			state.ActiveWorkflow = args[0]
			if err := app.teams.PersistMutation(workspaceValue, team, state, ""); err != nil {
				return err
			}
			fmt.Fprintf(command.OutOrStdout(), "active workflow: %s\n", args[0])
			return nil
		},
	}
}

func (app *App) newWorkflowWorkerCommand() *cobra.Command {
	command := &cobra.Command{Use: "worker", Short: "toggle a worker in the active workflow"}
	command.AddCommand(app.newWorkflowWorkerToggleCommand("enable", false), app.newWorkflowWorkerToggleCommand("disable", true))
	return command
}

func (app *App) newWorkflowWorkerToggleCommand(name string, disabled bool) *cobra.Command {
	return &cobra.Command{
		Use:   name + " <worker-id>",
		Short: map[bool]string{true: "disable a workflow worker", false: "enable a workflow worker"}[disabled],
		Args:  exactArgs(1),
		ValidArgsFunction: func(command *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			_, team, state, err := app.activeTeam()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			workflow, err := state.SelectedWorkflow(team.Team)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			seen := map[string]struct{}{}
			ids := []string{}
			for _, phase := range workflow.Phases {
				for _, workerID := range phase.Workers {
					if _, exists := seen[workerID]; exists {
						continue
					}
					seen[workerID] = struct{}{}
					ids = append(ids, workerID)
				}
			}
			return ids, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(command *cobra.Command, args []string) error {
			workspaceValue, team, state, err := app.activeTeam()
			if err != nil {
				return err
			}
			if !state.ActiveWorkflowUsesWorker(team.Team, args[0]) {
				return fmt.Errorf("active workflow does not use worker %q", args[0])
			}
			state.SetWorkerDisabled(args[0], disabled)
			if len(state.WorkflowWorkerIDs(team.Team)) == 0 {
				return errors.New("cannot disable every worker in the active workflow")
			}
			if err := app.teams.PersistMutation(workspaceValue, team, state, ""); err != nil {
				return err
			}
			fmt.Fprintf(command.OutOrStdout(), "workflow worker %s: %s\n", name, args[0])
			return nil
		},
	}
}
