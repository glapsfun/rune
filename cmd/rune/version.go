package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newVersionCmd builds the `version` command. It prints the same string as the
// `--version` flag (Cobra's default version template: "<name> version <Version>"),
// reading the root's resolved Version so the two stay byte-identical.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version and build commit",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root := cmd.Root()
			fmt.Fprintf(cmd.OutOrStdout(), "%s version %s\n", root.Name(), root.Version)
			return nil
		},
	}
}
