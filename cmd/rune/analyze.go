package main

import (
	"github.com/spf13/cobra"

	"github.com/rune-task-runner/rune/internal/cli"
)

// newAnalyzeCmd builds the `analyze` command: static analysis of a Runefile and
// its transitive imports, reporting diagnostics without executing anything
// (spec FR-023). Exit 0 (no errors), 3 (error diagnostics), 1 (internal failure).
func newAnalyzeCmd(opts *cli.Options) *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "analyze [path]",
		Short: "Statically analyze a Runefile; run nothing",
		Long: `Analyze a Runefile (default: the discovered Runefile) together with its
transitive imports and report all diagnostics. Nothing is executed. Exit code is
0 when there are no error diagnostics, 3 when there are, and 1 on internal
failure. Use --json for machine-readable output.`,
		Example: `  rune analyze
  rune analyze Runefile
  rune analyze --json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var path string
			if len(args) == 1 {
				path = args[0]
			}
			return cli.Analyze(*opts, path, jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable diagnostics")
	return cmd
}
