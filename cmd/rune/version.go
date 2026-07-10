package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rune-task-runner/rune/internal/cli"
	"github.com/rune-task-runner/rune/internal/config"
)

// newVersionCmd builds the `version` command. With no flags it prints the
// installed Rune version (its first line byte-identical to `rune --version`)
// followed by the Runefile language version. With --check it reports whether the
// installed binary satisfies the applicable Runefile's `minimum_version`, and
// --json makes that report machine-readable.
func newVersionCmd(opts *cli.Options) *cobra.Command {
	var (
		check   bool
		jsonOut bool
	)
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version info, or check Runefile compatibility",
		Long: `Print Rune's version and the Runefile language version. The first line is
identical to 'rune --version'. Use --check to report whether the installed
binary satisfies the current Runefile's minimum_version, and --json for
machine-readable output.`,
		Example: `  rune version
  rune version --check
  rune version --check --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if check {
				return cli.VersionCheck(*opts, jsonOut)
			}
			if jsonOut {
				return &cli.UsageError{Err: errors.New("--json requires --check")}
			}
			root := cmd.Root()
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%s version %s\n", root.Name(), root.Version)
			fmt.Fprintf(out, "runefile language %s\n", config.CurrentVersion)
			return nil
		},
	}
	f := cmd.Flags()
	f.BoolVar(&check, "check", false, "report whether the installed Rune satisfies the Runefile's minimum_version")
	f.BoolVar(&jsonOut, "json", false, "with --check, print machine-readable JSON")
	return cmd
}
