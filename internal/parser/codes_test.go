package parser

import (
	"testing"

	"github.com/rune-task-runner/rune/internal/diag"
)

func firstCode(diags diag.List, code string) *diag.Diagnostic {
	for i := range diags {
		if diags[i].Code == code {
			return &diags[i]
		}
	}
	return nil
}

// TestParserCodes asserts that parser/lexer diagnostics carry their stable codes
// (spec FR-010) with valid, in-bounds ranges.
func TestParserCodes(t *testing.T) {
	cases := []struct {
		name string
		src  string
		code string
	}{
		{"unterminated string", "x := \"abc\n", diag.CodeUnterminatedStr},
		{"invalid indentation (mixed)", "build:\n\t @echo hi\n", diag.CodeInvalidIndent},
		{"invalid attribute", "[bogus]\nbuild:\n    @echo hi\n", diag.CodeInvalidAttribute},
		{"incomplete expression", "x := \n", diag.CodeIncompleteExpr},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, diags := Parse("Runefile", tc.src)
			d := firstCode(diags, tc.code)
			if d == nil {
				t.Fatalf("expected code %s, got %v", tc.code, diags)
			}
			if d.Severity != diag.Error {
				t.Errorf("code %s: severity = %v, want Error", tc.code, d.Severity)
			}
			if !d.Span.IsValid() {
				t.Errorf("code %s: span not populated", tc.code)
			}
		})
	}
}

// TestUnexpectedTokenCode covers the generic RUNE1001 path (expect / top-level).
func TestUnexpectedTokenCode(t *testing.T) {
	// A bare '@' at top level is unexpected.
	_, diags := Parse("Runefile", "@\n")
	if firstCode(diags, diag.CodeUnexpectedToken) == nil {
		t.Fatalf("expected RUNE1001, got %v", diags)
	}
}
