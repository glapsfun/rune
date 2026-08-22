package parser

import (
	"testing"

	"github.com/rune-task-runner/rune/internal/ast"
)

func TestParseContextAttribute(t *testing.T) {
	f, diags := Parse("Runefile", "[context]\nhealth:\n    @echo hi\n")
	if diags.HasErrors() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(f.Tasks) != 1 {
		t.Fatalf("want 1 task, got %d", len(f.Tasks))
	}
	if f.Tasks[0].Attr(ast.AttrContext) == nil {
		t.Fatalf("task should carry the context attribute: %+v", f.Tasks[0].Attributes)
	}
}

func TestParseContextAttributeCombines(t *testing.T) {
	f, diags := Parse("Runefile", "[context, private]\nhealth:\n    @echo hi\n")
	if diags.HasErrors() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	tk := f.Tasks[0]
	if tk.Attr(ast.AttrContext) == nil || tk.Attr(ast.AttrPrivate) == nil {
		t.Fatalf("both attributes should parse: %+v", tk.Attributes)
	}
}
