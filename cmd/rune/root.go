package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/rune-task-runner/rune/internal/cli"
)

// newRootCmd builds the root command: Rune's task dispatcher. Built-in
// capabilities (serve/mcp, version, plus Cobra's auto-added completion/help) are
// registered as subcommands by the caller. Any first positional that does not
// match a subcommand is a dynamic task invocation handled by RunE.
func newRootCmd(opts *cli.Options, version, commit string) *cobra.Command {
	root := &cobra.Command{
		Use:           "rune [global flags] [VAR=VALUE ...] [TASK [ARGS...]] ...",
		Short:         "A shared task runner for humans and AI agents",
		Version:       fmt.Sprintf("%s (commit %s)", version, commit),
		SilenceUsage:  true,
		SilenceErrors: true,
		// Tasks are dynamic positionals, not subcommands. Without this, adding
		// subcommands makes Cobra's default validator reject an unknown first
		// arg ("unknown command") instead of routing it to RunE as a task.
		Args: cobra.ArbitraryArgs,
		// Resolve streams/context once, before the task path OR any subcommand,
		// so every command observes the same I/O and cancellation.
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			opts.Cwd, _ = os.Getwd()
			opts.Stdin = cmd.InOrStdin()
			opts.Stdout = cmd.OutOrStdout()
			opts.Stderr = cmd.ErrOrStderr()
			opts.Color = useColor()
			opts.Version = version
			opts.Ctx = cmd.Context()
			return nil
		},
		// This fires only when no subcommand matched the first positional:
		// built-in commands take precedence; `rune -- <task>` escapes to here.
		// Trailing task flags pass through untouched (see SetInterspersed below).
		RunE: func(_ *cobra.Command, args []string) error {
			return cli.Run(*opts, args)
		},
	}

	// Stop global-flag parsing at the first positional so trailing task flags
	// pass through to the task untouched.
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

	return root
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
