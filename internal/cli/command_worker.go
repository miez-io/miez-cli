package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/manuel/miez-cli/internal/model"
)

func (app *App) newWorkerCommand() *cobra.Command {
	command := &cobra.Command{Use: "worker", Short: "configure active worker capabilities"}
	command.AddCommand(app.newWorkerSkillCommand(), app.newWorkerModelCommand())
	return command
}

func (app *App) newWorkerModelCommand() *cobra.Command {
	command := &cobra.Command{Use: "model", Short: "inspect or change agent models"}
	command.AddCommand(app.newWorkerModelListCommand(), app.newWorkerModelSetCommand())
	return command
}

func (app *App) newWorkerModelListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list model ids usable in an agent worker's model (or a team's default_model) field",
		Args:  noArgs,
		RunE: func(command *cobra.Command, args []string) error {
			// The active team can extend or override the built-in models
			// (see the generated team index's models field), so list its merged catalog
			// when one is active, and just the built-ins otherwise.
			var activeTeam model.Team
			if _, team, _, err := app.activeTeam(); err == nil {
				activeTeam = team.Team
			}
			for _, id := range model.ModelIDs(activeTeam) {
				fmt.Fprintln(command.OutOrStdout(), id)
			}
			return nil
		},
	}
}

func (app *App) newWorkerModelSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <worker-id> <model-id>",
		Short: "switch a kind: agent worker to a different model",
		Args:  exactArgs(2),
		ValidArgsFunction: func(command *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			_, team, _, err := app.activeTeam()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) == 0 {
				ids := make([]string, 0, len(team.Team.Workers))
				for _, worker := range team.Team.Workers {
					if worker.Kind == model.WorkerAgent {
						ids = append(ids, worker.ID)
					}
				}
				return ids, cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) == 1 {
				return model.ModelIDs(team.Team), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(command *cobra.Command, args []string) error {
			workspaceValue, team, state, err := app.activeTeam()
			if err != nil {
				return err
			}
			worker, ok := team.Team.Worker(args[0])
			if !ok {
				return fmt.Errorf("team %q has no worker %q", team.Team.ID, args[0])
			}
			if worker.Kind != model.WorkerAgent {
				return fmt.Errorf("worker %q is kind %q; only kind: agent workers have a model", worker.ID, worker.Kind)
			}
			modelID := args[1]
			if _, ok := model.FindModel(team.Team, modelID); !ok {
				return fmt.Errorf("model %q is not supported; choose one of: %s", modelID, strings.Join(model.ModelIDs(team.Team), ", "))
			}
			state.SetWorkerModel(worker.ID, modelID)
			if err := app.teams.PersistMutation(workspaceValue, team, state, ""); err != nil {
				return err
			}
			fmt.Fprintf(command.OutOrStdout(), "worker model set: %s -> %s\n", worker.ID, modelID)
			return nil
		},
	}
}

func (app *App) newWorkerSkillCommand() *cobra.Command {
	command := &cobra.Command{Use: "skill", Short: "assign or remove worker skills"}
	command.AddCommand(app.newWorkerSkillMutationCommand("add", true), app.newWorkerSkillMutationCommand("remove", false))
	return command
}

func (app *App) newWorkerSkillMutationCommand(name string, add bool) *cobra.Command {
	return &cobra.Command{
		Use:   name + " <worker-id> <skill-id>",
		Short: map[bool]string{true: "assign a skill", false: "remove a skill override"}[add],
		Args:  exactArgs(2),
		ValidArgsFunction: func(command *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			_, team, _, err := app.activeTeam()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) == 0 {
				ids := make([]string, 0, len(team.Team.Workers))
				for _, worker := range team.Team.Workers {
					ids = append(ids, worker.ID)
				}
				return ids, cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) == 1 {
				ids := make([]string, 0, len(team.Team.Skills))
				for _, skill := range team.Team.Skills {
					ids = append(ids, skill.ID)
				}
				return ids, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(command *cobra.Command, args []string) error {
			workspaceValue, team, state, err := app.activeTeam()
			if err != nil {
				return err
			}
			worker, ok := team.Team.Worker(args[0])
			if !ok {
				return fmt.Errorf("team %q has no worker %q", team.Team.ID, args[0])
			}
			if _, ok := team.Team.Skill(args[1]); !ok {
				return fmt.Errorf("team %q has no skill %q", team.Team.ID, args[1])
			}
			if _, err := team.ReadSkill(args[1]); err != nil {
				return err
			}
			if !add && !workerHasSkill(worker, state, args[1]) {
				return fmt.Errorf("worker %q does not have skill %q", worker.ID, args[1])
			}
			state.SetWorkerSkill(worker.ID, args[1], add)
			if err := app.teams.PersistMutation(workspaceValue, team, state, ""); err != nil {
				return err
			}
			fmt.Fprintf(command.OutOrStdout(), "worker %s skill %s: %s\n", name, args[0], args[1])
			return nil
		},
	}
}

func workerHasSkill(worker model.Worker, state model.TeamState, skillID string) bool {
	for _, skill := range worker.Skills {
		if skill == skillID {
			return true
		}
	}
	for _, skill := range state.WorkerSkills[worker.ID] {
		if skill == skillID {
			return true
		}
	}
	return false
}
