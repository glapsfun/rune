package analyzer

import (
	"testing"

	"github.com/rune-task-runner/rune/internal/diag"
	"github.com/rune-task-runner/rune/internal/parser"
)

func checkDocs(t *testing.T, src string) diag.List {
	t.Helper()
	f, pdiags := parser.Parse("Runefile", src)
	if pdiags.HasErrors() {
		t.Fatalf("unexpected parse errors: %v", pdiags)
	}
	return CheckDocumentation(f)
}

func TestUndocumentedPublicTaskWarns(t *testing.T) {
	diags := checkDocs(t, "build:\n    @echo build\n")
	d := hasCode(diags, diag.CodeUndocumentedTask)
	if d == nil {
		t.Fatalf("expected RUNE2010 warning, got %v", diags)
	}
	if d.Severity != diag.Warning {
		t.Errorf("severity = %v, want Warning (must not gate exit 3)", d.Severity)
	}
	if diags.HasErrors() {
		t.Error("documentation check must not produce error-severity diagnostics")
	}
}

func TestDocumentedTaskNoWarning(t *testing.T) {
	diags := checkDocs(t, "# Build it.\nbuild:\n    @echo build\n")
	if d := hasCode(diags, diag.CodeUndocumentedTask); d != nil {
		t.Errorf("documented task should not warn, got %+v", d)
	}
}

func TestPrivateTaskNoDocWarning(t *testing.T) {
	// A private task (leading underscore) needs no documentation.
	diags := checkDocs(t, "_helper:\n    @echo hi\n")
	if d := hasCode(diags, diag.CodeUndocumentedTask); d != nil {
		t.Errorf("private task should not warn, got %+v", d)
	}
}

// TestAnalyzeExcludesDocWarning guards the design decision that the execution
// path (Analyze) never emits the documentation warning (only CheckDocumentation
// does), so `rune` runs stay quiet and error gating is unaffected.
func TestAnalyzeExcludesDocWarning(t *testing.T) {
	diags := analyzeDiags(t, "build:\n    @echo build\n")
	if d := hasCode(diags, diag.CodeUndocumentedTask); d != nil {
		t.Errorf("Analyze must not emit RUNE2010; got %+v", d)
	}
}
