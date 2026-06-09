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
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

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
	opts.Ctx = ctx

	root := &cobra.Command{
		Use:           "rune [global flags] [VAR=VALUE ...] [TASK [ARGS...]] ...",
		Short:         "A shared task runner for humans and AI agents",
		Version:       fmt.Sprintf("%s (commit %s)", version, commit),
		SilenceUsage:  true,
		SilenceErrors: true,
		// Task names are dynamic; stop global-flag parsing at the first
		// positional so trailing task flags pass through untouched.
		RunE: func(cmd *cobra.Command, posArgs []string) error {
			opts.Cwd, _ = os.Getwd()
			opts.Stdin = cmd.InOrStdin()
			opts.Stdout = cmd.OutOrStdout()
			opts.Stderr = cmd.ErrOrStderr()
			opts.Color = useColor()
			opts.Version = version
			// Reserved subcommands (mcp, serve) are handled explicitly so they
			// never silently shadow a task of the same name.
			if len(posArgs) > 0 && (posArgs[0] == "mcp" || posArgs[0] == "serve") {
				return runServe(opts, posArgs)
			}
			if len(posArgs) > 0 && posArgs[0] == "completion" {
				shell := "bash"
				if len(posArgs) > 1 {
					shell = posArgs[1]
				}
				return genCompletion(cmd.Root(), shell, os.Stdout)
			}
			return cli.Run(opts, posArgs)
		},
	}
	root.Flags().SetInterspersed(false)

	f := root.Flags()
	f.StringVarP(&opts.File, "file", "f", "", "use a specific Runefile instead of upward discovery")
	f.BoolVar(&opts.List, "list", false, "list non-private tasks with docs; run nothing")
	f.BoolVar(&opts.DryRun, "dry-run", false, "print the resolved execution plan; run nothing")
	f.BoolVar(&opts.Summary, "summary", false, "print task names that would run, one per line")
	f.BoolVar(&opts.Dump, "dump", false, "emit the parsed Runefile (canonical text, or JSON)")
	f.StringVar(&opts.DumpFormat, "format", "", "output format for --dump (json)")
	f.StringArrayVar(&opts.Set, "set", nil, "override a variable: --set NAME VALUE")
	f.BoolVar(&opts.Watch, "watch", false, "re-run on file changes")
	f.BoolVar(&opts.Choose, "choose", false, "interactive task picker")
	f.BoolVar(&opts.Yes, "yes", false, "auto-approve [confirm] tasks")
	f.BoolVar(&opts.Quiet, "quiet", false, "suppress command echo")
	f.BoolVar(&opts.Fmt, "fmt", false, "rewrite the Runefile in canonical formatting")
	f.BoolVar(&opts.ClearCache, "clear-cache", false, "remove the project-local .rune/cache directory")

	// Rune's own messages go to stderr so stdout stays clean for piping.
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	root.SetArgs(args)

	err := root.Execute()
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

// runServe dispatches the reserved `mcp` / `serve` subcommands. `mcp` is
// shorthand for stdio serving; `serve` accepts --http / --addr / --token-file.
func runServe(opts cli.Options, args []string) error {
	if args[0] == "mcp" {
		return cli.ServeMCP(opts, false, "", "")
	}
	useHTTP := false
	addr := ""
	tokenFile := ""
	for i := 1; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--http":
			useHTTP = true
		case a == "--mcp":
			// MCP is the only protocol; accepted for clarity.
		case a == "--addr":
			if i+1 < len(args) {
				i++
				addr = args[i]
			}
		case strings.HasPrefix(a, "--addr="):
			addr = strings.TrimPrefix(a, "--addr=")
		case a == "--token-file":
			if i+1 < len(args) {
				i++
				tokenFile = args[i]
			}
		case strings.HasPrefix(a, "--token-file="):
			tokenFile = strings.TrimPrefix(a, "--token-file=")
		}
	}
	return cli.ServeMCP(opts, useHTTP, addr, tokenFile)
}

// useColor reports whether ANSI color should be emitted on stderr (where Rune's
// own messages go): a TTY, NO_COLOR unset, and color globally enabled.
func useColor() bool {
	if color.NoColor {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())
}
