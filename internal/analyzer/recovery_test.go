package analyzer

import (
	"testing"

	"github.com/rune-task-runner/rune/internal/parser"
)

// TestAnalyzeRecoveredFile verifies the analyzer runs on a partially-recovered
// AST (the parser dropped a broken declaration but kept valid siblings) without
// panicking, and still analyzes the surviving declarations — the behavior the
// LSP relies on for live diagnostics on mid-edit files (spec FR-004).
func TestAnalyzeRecoveredFile(t *testing.T) {
	// A garbage top-level line sits between two valid tasks; the parser recovers
	// and keeps both (see parser TestRecoveryKeepsValidDeclarations). The second
	// surviving task depends on a missing task, which the analyzer must flag.
	src := "alpha:\n    echo a\n$$$ garbage @@@\nbeta: nope\n    echo b\n"
	f, pdiags := parser.Parse("Runefile", src)
	if !pdiags.HasErrors() {
		t.Fatal("expected parse diagnostics for the broken region")
	}

	// Must not panic, and should still detect the unknown dependency in "second".
	diags := Analyze(f)
	if hasCode(diags, "RUNE2001") == nil {
		t.Errorf("analyzer should still flag the unknown dependency after recovery; got %v", diags)
	}
}
