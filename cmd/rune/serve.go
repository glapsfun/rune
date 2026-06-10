package main

import (
	"github.com/spf13/cobra"

	"github.com/rune-task-runner/rune/internal/cli"
)

// newServeCmd builds the `serve` command (aliased `mcp`): it starts the MCP
// server over stdio by default, or Streamable HTTP with --http. It replaces the
// hand-rolled argument loop that previously lived in main.go.
func newServeCmd(opts *cli.Options) *cobra.Command {
	var (
		useHTTP   bool
		addr      string
		tokenFile string
		mcpFlag   bool
	)

	cmd := &cobra.Command{
		Use:     "serve",
		Aliases: []string{"mcp"},
		Short:   "Run the MCP server for agents and IDEs",
		RunE: func(_ *cobra.Command, _ []string) error {
			_ = mcpFlag // MCP is the only protocol; --mcp is accepted for clarity.
			return cli.ServeMCP(*opts, useHTTP, addr, tokenFile)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&useHTTP, "http", false, "serve over Streamable HTTP instead of stdio")
	f.StringVar(&addr, "addr", "", "HTTP listen address (with --http)")
	f.StringVar(&tokenFile, "token-file", "", "bearer-token file for HTTP auth (with --http)")
	f.BoolVar(&mcpFlag, "mcp", false, "accepted for clarity; MCP is the only protocol")

	return cmd
}

// validateServeFlags rejects HTTP-only flags supplied without --http. The
// complementary rule — HTTP requires a token file — is enforced inside
// cli.ServeMCP, so it is intentionally NOT duplicated here.
func validateServeFlags(useHTTP bool, addr, tokenFile string) error {
	return nil // implemented in US3
}
