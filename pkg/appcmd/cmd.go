package appcmd

import (
	"context"
	"io"
	"os"

	"github.com/demdxx/goconfig"
)

// ICommand is the interface all commands must implement.
type ICommand interface {
	String() string
	Cmd() string
	Help() string
	Run(ctx context.Context, args []string) error
}

// HelpProvider is an optional interface commands can implement to print
// detailed help that includes the configuration options table.
type HelpProvider interface {
	PrintHelp(w io.Writer)
}

// CommandFunc is the execution function for a typed command.
type CommandFunc[T any] func(ctx context.Context, args []string, config *T) error

// ContextInitFunc initializes or enriches the context before command execution.
type ContextInitFunc func(ctx context.Context) (context.Context, error)

// Command is a generic command with a typed configuration struct T.
// The configuration is loaded from defaults, environment variables, and CLI
// flags using the goconfig library, following the struct tags:
//
//   - default:"…"     — default value
//   - env:"…"         — environment variable name
//   - envPrefix:"…"   — prefix applied to all env tags in nested structs
//   - cli:"…"         — long CLI flag (without "--")
//   - json:"…"        — display name fallback
//   - field:"…"       — display name (highest priority)
type Command[T any] struct {
	Name        string
	HelpDesc    string
	Exec        CommandFunc[T]
	ContextInit ContextInitFunc
}

func (c *Command[T]) String() string { return c.Name }
func (c *Command[T]) Cmd() string    { return c.Name }
func (c *Command[T]) Help() string   { return c.HelpDesc }

// PrintHelp writes full command help including the configuration options table to w.
func (c *Command[T]) PrintHelp(w io.Writer) {
	printCommandUsage[T](w, c.Name, c.HelpDesc)
}

// Run executes the command. It intercepts --help / -h / help before loading
// config so users can request help even with missing required env vars.
// Scanning stops at "--" so downstream flags are not misinterpreted.
func (c *Command[T]) Run(ctx context.Context, args []string) error {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			c.PrintHelp(os.Stdout)
			return nil
		}
		if arg == "--" {
			break
		}
	}

	var config T
	err := goconfig.Load(
		&config,
		goconfig.WithDefaults(),
		goconfig.WithEnv(),
		goconfig.WithCustomArgs(args...),
	)
	if err != nil {
		return err
	}
	if c.ContextInit != nil {
		if ctx, err = c.ContextInit(ctx); err != nil {
			return err
		}
	}
	return c.Exec(ctx, args, &config)
}

// WithInitContext returns a shallow copy of the command with the given context initializer.
func (c *Command[T]) WithInitContext(ctxWrapper ContextInitFunc) *Command[T] {
	newCmd := *c
	newCmd.ContextInit = ctxWrapper
	return &newCmd
}
