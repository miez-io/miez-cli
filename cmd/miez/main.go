package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/manuel/miez-cli/internal/cli"
)

func main() {
	app, err := cli.New(os.Stdout, os.Stderr, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := app.Execute(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		var exitError *cli.ExitError
		if errors.As(err, &exitError) {
			os.Exit(exitError.Code)
		}
		os.Exit(1)
	}
}
