package cli

// LSPOptions configures the language server started by LSP.
type LSPOptions struct {
	LogFile  string // path to write logs to; empty means stderr
	LogLevel string // error|warn|info|debug
}

// LSP starts the Rune language server over stdio. Implemented in the US1 phase;
// stubbed for now so the command wiring compiles.
func LSP(opts Options, lspOpts LSPOptions) error {
	_ = lspOpts
	return &UsageError{Err: errorf("rune lsp: not implemented yet")}
}
