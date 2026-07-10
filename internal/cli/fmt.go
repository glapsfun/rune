package cli

import (
	"fmt"
	"os"

	"github.com/rune-task-runner/rune/internal/formatter"
	"github.com/rune-task-runner/rune/internal/parser"
)

// fmtRewrite parses the Runefile and rewrites it in canonical form in place,
// delegating the formatting itself to the shared internal/formatter package
// (the same formatter the language server calls — spec FR-020).
func fmtRewrite(opts Options, runefile string) error {
	source, err := os.ReadFile(runefile)
	if err != nil {
		return &UsageError{Err: err}
	}
	file, diags := parser.Parse(runefile, string(source))
	if diags.HasErrors() {
		renderDiags(opts, diags, newSourceProvider(runefile, source))
		return &ValidationError{Err: errorf("%d static error(s)", countErrors(diags))}
	}
	formatted := formatter.Format(file)
	if err := os.WriteFile(runefile, []byte(formatted), 0o644); err != nil {
		return &UsageError{Err: err}
	}
	fmt.Fprintf(opts.Stderr, "formatted: %s\n", runefile)
	return nil
}
