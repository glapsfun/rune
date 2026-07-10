package main

import (
	"github.com/spf13/cobra"

	"github.com/rune-task-runner/rune/internal/cli"
)

// newLSPCmd builds the `lsp` command: a JSON-RPC/LSP 3.17 language server over
// stdio (spec FR-011). stdout carries protocol messages only; logs go to stderr
// or --log-file (FR-012). It executes nothing (FR-028).
func newLSPCmd(opts *cli.Options) *cobra.Command {
	var (
		logFile  string
		logLevel string
	)
	cmd := &cobra.Command{
		Use:   "lsp",
		Short: "Start the Rune language server (stdio, LSP 3.17)",
		Long: `Start the Rune language server over stdin/stdout using JSON-RPC and LSP 3.17.
stdout is reserved for protocol messages; logs are written to stderr or the
--log-file path. The server executes no tasks or commands.`,
		Example: `  rune lsp
  rune lsp --log-file ~/.cache/rune/lsp.log
  rune lsp --log-level debug`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return cli.LSP(*opts, cli.LSPOptions{LogFile: logFile, LogLevel: logLevel})
		},
	}
	f := cmd.Flags()
	f.StringVar(&logFile, "log-file", "", "write logs to this file instead of stderr")
	f.StringVar(&logLevel, "log-level", "info", "log verbosity: error|warn|info|debug")
	return cmd
}
