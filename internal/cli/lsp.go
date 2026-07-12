package cli

import (
	"io"
	"os"

	"github.com/rune-task-runner/rune/internal/lsp"
)

// LSPOptions configures the language server started by LSP.
type LSPOptions struct {
	LogFile  string // path to write logs to; empty means stderr
	LogLevel string // error|warn|info|debug (coarse for the MVP)
}

// LSP starts the Rune language server over stdio (spec FR-011). stdout carries
// only JSON-RPC protocol messages; logs go to stderr or --log-file (FR-012).
// It executes nothing (FR-028).
func LSP(opts Options, lspOpts LSPOptions) error {
	var logw io.Writer = opts.Stderr
	if lspOpts.LogFile != "" {
		f, err := os.OpenFile(lspOpts.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return &UsageError{Err: err}
		}
		defer f.Close()
		logw = f
	}

	srv := lsp.NewServer(opts.Stdin, opts.Stdout, lsp.Options{
		Version:   opts.Version,
		LogWriter: logw,
	})
	return srv.Run()
}
