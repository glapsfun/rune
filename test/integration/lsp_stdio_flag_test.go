package integration

import (
	"fmt"
	"strings"
	"testing"
)

// lspMessage frames a JSON-RPC body with the LSP Content-Length header.
func lspMessage(body string) string {
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)
}

// lspHandshake is a minimal initialize → initialized → shutdown → exit
// session; a conforming server replies to initialize and exits 0.
func lspHandshake() string {
	var b strings.Builder
	b.WriteString(lspMessage(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"processId":null,"rootUri":null,"capabilities":{}}}`))
	b.WriteString(lspMessage(`{"jsonrpc":"2.0","method":"initialized","params":{}}`))
	b.WriteString(lspMessage(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":null}`))
	b.WriteString(lspMessage(`{"jsonrpc":"2.0","method":"exit"}`))
	return b.String()
}

// The `--stdio` flag is a de-facto convention passed by LSP clients —
// vscode-languageclient appends it for TransportKind.stdio executables, and
// other editors do the same. `rune lsp` MUST accept (and ignore) it instead
// of dying with "unknown flag", which crash-looped every VS Code install.
func TestLSPAcceptsStdioFlag(t *testing.T) {
	dir := writeRunefile(t, "build:\n    echo hi\n")

	for _, args := range [][]string{
		{"lsp"},            // the documented invocation
		{"lsp", "--stdio"}, // what LSP clients actually spawn
	} {
		r := runWithStdin(t, dir, lspHandshake(), args...)
		if strings.Contains(r.stderr, "unknown flag") {
			t.Fatalf("%v rejected a flag:\nstderr: %s", args, r.stderr)
		}
		if !strings.Contains(r.stdout, `"serverInfo"`) {
			t.Fatalf("%v did not answer initialize (exit %d)\nstdout: %q\nstderr: %s",
				args, r.code, r.stdout, r.stderr)
		}
		if r.code != 0 {
			t.Fatalf("%v exited %d after clean shutdown\nstderr: %s", args, r.code, r.stderr)
		}
	}
}
