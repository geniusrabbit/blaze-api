package appcmd

import (
	"context"
	"fmt"
	"io"
	"os"
)

// BeforeCommandRunFunc is called once before a command's Run, allowing the
// caller to enrich or validate the context (logging, tracing, etc.).
type BeforeCommandRunFunc func(ctx context.Context, cmd ICommand) (context.Context, error)

// App is the root application that dispatches CLI arguments to commands.
type App struct {
	Name        string
	Description string
	Version     string
	BuildDate   string
	BuildCommit string
	CmdList     ICommands

	BeforeCommandRun BeforeCommandRunFunc
}

// Run dispatches the command identified by args[1], handling built-in flags
// --version/-v, --help/-h, and the special "help [command]" sub-command
// before reaching user-defined commands.
func (app *App) Run(ctx context.Context, args []string) error {
	if len(args) < 2 {
		app.PrintUsage(os.Stdout)
		return nil
	}

	cmdName := args[1]

	switch cmdName {
	case "--version", "-v":
		_, _ = fmt.Fprintf(os.Stdout, "%s version %s (commit: %s, built: %s)\n",
			app.Name, app.Version, app.BuildCommit, app.BuildDate)
		return nil

	case "--help", "-h":
		app.PrintUsage(os.Stdout)
		return nil

	case "help":
		// "help <command>" → delegate to command's HelpProvider
		if len(args) > 2 {
			if target := app.CmdList.Get(args[2]); target != nil {
				if hp, ok := target.(HelpProvider); ok {
					hp.PrintHelp(os.Stdout)
					return nil
				}
				// Command found but doesn't implement HelpProvider — show generic line.
				_, _ = fmt.Fprintf(os.Stdout, "Command: %s\n  %s\n\n", target.Cmd(), target.Help())
				return nil
			}
			_, _ = fmt.Fprintf(os.Stderr, "unknown command: %q\n\n", args[2])
		}
		app.PrintUsage(os.Stdout)
		return nil
	}

	icmd := app.CmdList.Get(cmdName)
	if icmd == nil {
		_, _ = fmt.Fprintf(os.Stderr, "unknown command: %q\n\n", cmdName)
		app.PrintUsage(os.Stdout)
		return nil
	}

	if app.BeforeCommandRun != nil {
		var err error
		ctx, err = app.BeforeCommandRun(ctx, icmd)
		if err != nil {
			return fmt.Errorf("before command run: %w", err)
		}
	}

	return icmd.Run(ctx, args[2:])
}

// PrintUsage writes the full application usage to w.
func (app *App) PrintUsage(w io.Writer) {
	_, _ = fmt.Fprintf(w, "Usage: %s <command> [options]\n", app.Name)
	_, _ = fmt.Fprintf(w, "Version:      %s\n", app.Version)
	_, _ = fmt.Fprintf(w, "Build date:   %s\n", app.BuildDate)
	_, _ = fmt.Fprintf(w, "Build commit: %s\n", app.BuildCommit)
	_, _ = fmt.Fprintln(w)
	if app.Description != "" {
		_, _ = fmt.Fprintf(w, "%s\n\n", app.Description)
	}

	_, _ = fmt.Fprintf(w, "Commands:\n")
	for _, cmd := range app.CmdList {
		_, _ = fmt.Fprintf(w, "  %-14s  %s\n", cmd.Cmd(), cmd.Help())
	}
	_, _ = fmt.Fprintf(w, "  %-14s  %s\n", "help", "print this help, or 'help <command>' for full option listing")
	_, _ = fmt.Fprintln(w)

	_, _ = fmt.Fprintf(w, "Flags:\n")
	_, _ = fmt.Fprintf(w, "  --version, -v   print version information\n")
	_, _ = fmt.Fprintf(w, "  --help, -h      print help\n")
	_, _ = fmt.Fprintln(w)
}
