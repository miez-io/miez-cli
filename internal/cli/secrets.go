package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// isInteractive reports whether in is a real terminal an MCP secret prompt
// can be shown on, and CI is not set (the escape hatch that forces
// non-interactive behavior even when stdin happens to be a TTY).
func isInteractive(in io.Reader) bool {
	if os.Getenv("CI") != "" {
		return false
	}
	file, ok := in.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}

// secretPrompt prompts for one missing MCP secret value on app.Out/app.In,
// masking the input when masked is true and app.In is a real terminal.
func (app *App) secretPrompt(variable string, masked bool) (string, error) {
	fmt.Fprintf(app.Out, "%s: ", variable)
	if masked {
		if file, ok := app.In.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
			data, err := term.ReadPassword(int(file.Fd()))
			fmt.Fprintln(app.Out)
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(string(data)), nil
		}
	}
	if app.input == nil {
		app.input = bufio.NewReader(app.In)
	}
	line, err := app.input.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
