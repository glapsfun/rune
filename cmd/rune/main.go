// Command rune is a shared task runner for humans and AI agents. It parses a
// Runefile, statically validates it, and runs tasks — from the CLI or, via MCP,
// from agents and IDEs.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rune-task-runner/rune/internal/cli"
)

// Build metadata, overridden via -ldflags at release time.
var (
	version = "dev"
	commit  = "none"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	// Cancel running tasks on SIGINT/SIGTERM; the scheduler/executors observe
	// the context and child processes are terminated (exit 130).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var opts cli.Options

	root := newRootCmd(&opts, version, commit)
	// Built-in subcommands. Registering any subcommand also makes Cobra add its
	// `completion` and `help` commands automatically.
	root.AddCommand(newServeCmd(&opts), newVersionCmd())

	// Rune's own messages go to stderr so stdout stays clean for piping.
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	root.SetArgs(args)

	err := root.ExecuteContext(ctx)
	if err != nil {
		// Diagnostics for validation errors are already rendered by the
		// pipeline; a [no-exit-message] failure suppresses its banner. Only
		// print a terse banner for other error classes.
		var ve *cli.ValidationError
		var tf *cli.TaskFailure
		silent := errors.As(err, &tf) && tf.Silent
		if !errors.As(err, &ve) && !silent {
			fmt.Fprintln(os.Stderr, "rune: "+err.Error())
		}
	}
	return cli.CodeFor(err)
}
