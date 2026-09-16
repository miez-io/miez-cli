package cli

import "github.com/spf13/cobra"

// noArgs rejects any positional arguments with a usage exit code.
func noArgs(command *cobra.Command, args []string) error {
	if err := cobra.NoArgs(command, args); err != nil {
		return &ExitError{Code: 2, Err: err}
	}
	return nil
}

// exactArgs rejects any positional argument count other than count with a
// usage exit code.
func exactArgs(count int) cobra.PositionalArgs {
	return func(command *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(count)(command, args); err != nil {
			return &ExitError{Code: 2, Err: err}
		}
		return nil
	}
}

// maxArgs rejects more than count positional arguments with a usage exit
// code, while allowing fewer.
func maxArgs(count int) cobra.PositionalArgs {
	return func(command *cobra.Command, args []string) error {
		if err := cobra.MaximumNArgs(count)(command, args); err != nil {
			return &ExitError{Code: 2, Err: err}
		}
		return nil
	}
}
