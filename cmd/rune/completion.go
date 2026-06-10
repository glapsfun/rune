package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// genCompletion writes a shell completion script for the given shell. Task names
// are dynamic, so completion covers the global flags and reserved subcommands.
func genCompletion(root *cobra.Command, shell string, w io.Writer) error {
	switch shell {
	case "bash":
		return root.GenBashCompletionV2(w, true)
	case "zsh":
		return root.GenZshCompletion(w)
	case "fish":
		return root.GenFishCompletion(w, true)
	case "powershell", "pwsh":
		return root.GenPowerShellCompletionWithDesc(w)
	default:
		return fmt.Errorf("unsupported shell %q (want bash, zsh, fish, or powershell)", shell)
	}
}
