package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newCompletionCmd builds the `completion` command. Rune provides its own rather
// than Cobra's default because the default prints help and exits 0 on an unknown
// shell; here an unsupported shell is a clear, non-zero error that names the
// supported shells (FR-013). The hidden `__complete` driver that powers dynamic
// task-name completion is registered by Cobra independently and is unaffected.
func newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate a shell completion script",
		Long: `Generate a shell completion script for rune. Completions include the task
names defined in your Runefile, resolved dynamically as you type.

Bash:
  rune completion bash > /etc/bash_completion.d/rune            # system-wide
  rune completion bash > ~/.local/share/bash-completion/completions/rune

Zsh:
  rune completion zsh > "${fpath[1]}/_rune"                     # then restart your shell

Fish:
  rune completion fish > ~/.config/fish/completions/rune.fish

PowerShell:
  rune completion powershell | Out-String | Invoke-Expression`,
		Args:                  cobra.ExactArgs(1),
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(out, true)
			case "zsh":
				return cmd.Root().GenZshCompletion(out)
			case "fish":
				return cmd.Root().GenFishCompletion(out, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(out)
			default:
				return fmt.Errorf("unsupported shell %q (want bash, zsh, fish, or powershell)", args[0])
			}
		},
	}
}
