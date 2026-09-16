// Package cli wires the miez-cli command surface to deterministic services.
package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/manuel/miez-cli/internal/team"
)

// Version is replaced by release builds.
var Version = "0.1.0"

// ExitError carries the process exit code for a command failure.
type ExitError struct {
	Code int
	Err  error
}

func (errorValue *ExitError) Error() string {
	return errorValue.Err.Error()
}

func (errorValue *ExitError) Unwrap() error {
	return errorValue.Err
}

// App contains CLI dependencies and I/O destinations.
type App struct {
	Out   io.Writer
	Err   io.Writer
	In    io.Reader
	Cwd   string
	input *bufio.Reader
	teams *team.Service
}

// New creates a CLI application with injectable dependencies.
func New(out, errOut io.Writer, cwd string) (*App, error) {
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("resolve current directory: %w", err)
		}
	}
	return &App{
		Out:   out,
		Err:   errOut,
		In:    os.Stdin,
		Cwd:   cwd,
		teams: team.NewService(cwd),
	}, nil
}

// Execute runs the command tree with signal-aware cancellation.
func (app *App) Execute(ctx context.Context, args []string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	app.input = bufio.NewReader(app.In)
	command := app.NewCommand()
	command.SetArgs(args)
	err := command.ExecuteContext(ctx)
	if err == nil {
		return nil
	}
	var exitError *ExitError
	if errors.As(err, &exitError) {
		return err
	}
	if strings.HasPrefix(err.Error(), "unknown command") {
		return &ExitError{Code: 2, Err: err}
	}
	return err
}

// NewCommand constructs the command tree.
func (app *App) NewCommand() *cobra.Command {
	var verbose bool
	root := &cobra.Command{
		Use:           "miez",
		Short:         "Install and configure Markdown agent teams",
		Long:          "Install and configure Markdown agent teams. Commands: team, workflow, worker, check.",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(command *cobra.Command, args []string) error {
			app.teams.Interactive = isInteractive(app.In)
			app.teams.PromptSecret = app.secretPrompt
			if verbose {
				app.teams.Log = func(message string) {
					fmt.Fprintf(app.Err, "[verbose] %s\n", message)
				}
			}
			return nil
		},
		RunE: func(command *cobra.Command, args []string) error {
			return command.Help()
		},
	}
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "print which credential source resolved")
	root.SetOut(app.Out)
	root.SetErr(app.Err)
	root.SetFlagErrorFunc(func(command *cobra.Command, err error) error {
		return &ExitError{Code: 2, Err: err}
	})
	root.AddCommand(
		app.newTeamCommand(),
		app.newWorkflowCommand(),
		app.newWorkerCommand(),
		app.newCheckCommand(),
	)
	return root
}
