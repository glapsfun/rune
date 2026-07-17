package analyzer

import (
	"testing"

	"github.com/rune-task-runner/rune/internal/diag"
	"github.com/rune-task-runner/rune/internal/parser"
)

// analyzeDiags parses src and returns the analyzer's raw diagnostics.
func analyzeDiags(t *testing.T, src string) diag.List {
	t.Helper()
	f, pdiags := parser.Parse("Runefile", src)
	if pdiags.HasErrors() {
		t.Fatalf("unexpected parse errors: %v", pdiags)
	}
	return Analyze(f)
}

// hasCode reports whether any diagnostic carries the given stable code.
func hasCode(diags diag.List, code string) *diag.Diagnostic {
	for i := range diags {
		if diags[i].Code == code {
			return &diags[i]
		}
	}
	return nil
}

func TestSemanticCodes(t *testing.T) {
	cases := []struct {
		name string
		src  string
		code string
		sev  diag.Severity
	}{
		{"unknown dependency", "a: b\n    @echo a\n", diag.CodeUnknownDependency, diag.Error},
		{"duplicate task", "a:\n    @echo a\na:\n    @echo a2\n", diag.CodeDuplicateTask, diag.Error},
		{"undefined variable", "greet:\n    @echo {{nope}}\n", diag.CodeUndefinedVariable, diag.Error},
		{"wrong arg count", "greet name:\n    @echo {{name}}\nrun: (greet \"a\" \"b\")\n    @echo run\n", diag.CodeWrongArgCount, diag.Error},
		{"duplicate parameter", "greet a a:\n    @echo hi\n", diag.CodeDuplicateParam, diag.Error},
		{"dependency cycle", "c: c\n    @echo c\n", diag.CodeDependencyCycle, diag.Error},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			diags := analyzeDiags(t, tc.src)
			d := hasCode(diags, tc.code)
			if d == nil {
				t.Fatalf("expected code %s, got %v", tc.code, diags)
			}
			if d.Severity != tc.sev {
				t.Errorf("code %s: severity = %v, want %v", tc.code, d.Severity, tc.sev)
			}
			if !d.Span.IsValid() {
				t.Errorf("code %s: span is not populated (range must be valid)", tc.code)
			}
		})
	}
}

// TestCycleRelatedLocations verifies RUNE2003 lists every task in the cycle as a
// related location (spec FR-009).
func TestCycleRelatedLocations(t *testing.T) {
	src := "a: b\n    @echo a\nb: c\n    @echo b\nc: a\n    @echo c\n"
	diags := analyzeDiags(t, src)
	d := hasCode(diags, diag.CodeDependencyCycle)
	if d == nil {
		t.Fatalf("expected a dependency-cycle diagnostic, got %v", diags)
	}
	if len(d.Related) < 3 {
		t.Fatalf("expected >=3 related locations (a, b, c), got %d: %+v", len(d.Related), d.Related)
	}
	names := map[string]bool{}
	for _, r := range d.Related {
		names[r.Message] = true
		if !r.Span.IsValid() {
			t.Errorf("related location for %q has invalid span", r.Message)
		}
	}
	for _, want := range []string{"a", "b", "c"} {
		if !names[want] {
			t.Errorf("cycle related locations missing task %q (got %v)", want, names)
		}
	}
}
