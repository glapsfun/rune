package config

import (
	"strings"
	"testing"

	"github.com/rune-task-runner/rune/internal/parser"
)

func TestMinimumVersionAbsent(t *testing.T) {
	f, _ := parser.Parse("Runefile", "greet:\n    @echo hi\n")
	req, diags := MinimumVersion(f)
	if diags.HasErrors() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.Present {
		t.Errorf("expected no requirement, got Present=true")
	}
	if ok, _ := req.Satisfied("0.1.0"); !ok {
		t.Errorf("absent requirement should be satisfied by any version")
	}
}

func TestMinimumVersionValid(t *testing.T) {
	f, _ := parser.Parse("Runefile", "set minimum_version := \"0.8.0\"\ngreet:\n    @echo hi\n")
	req, diags := MinimumVersion(f)
	if diags.HasErrors() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !req.Present || req.Raw != "0.8.0" {
		t.Fatalf("req = %+v, want Present, Raw=0.8.0", req)
	}

	cases := []struct {
		installed string
		want      bool
	}{
		{"0.7.2", false},
		{"0.8.0", true},
		{"0.9.1", true},
	}
	for _, c := range cases {
		if ok, dev := req.Satisfied(c.installed); ok != c.want || dev {
			t.Errorf("Satisfied(%q) = (%v, dev=%v), want (%v, false)", c.installed, ok, dev, c.want)
		}
	}
}

func TestMinimumVersionDevBuildNotBlocked(t *testing.T) {
	f, _ := parser.Parse("Runefile", "set minimum_version := \"0.8.0\"\ngreet:\n    @echo hi\n")
	req, _ := MinimumVersion(f)
	ok, dev := req.Satisfied("dev")
	if !ok || !dev {
		t.Errorf("Satisfied(\"dev\") = (%v, %v), want (true, true)", ok, dev)
	}
}

func TestMinimumVersionSpanPointsAtValue(t *testing.T) {
	src := "set minimum_version := \"0.8.0\"\n"
	f, _ := parser.Parse("Runefile", src)
	req, _ := MinimumVersion(f)
	// The value literal "0.8.0" starts at column 24 (1-based) on line 1.
	if req.Span.Start.Line != 1 || req.Span.Start.Col != 24 {
		t.Errorf("span = %s, want 1:24 (the value literal)", req.Span)
	}
}

func TestMinimumVersionNonStatic(t *testing.T) {
	// A value derived from a non-literal expression is rejected.
	src := "required := \"0.8.0\"\nset minimum_version := required\ngreet:\n    @echo hi\n"
	f, _ := parser.Parse("Runefile", src)
	_, diags := MinimumVersion(f)
	if !diags.HasErrors() {
		t.Fatalf("expected a diagnostic for a non-static value")
	}
	if !strings.Contains(diags[0].Message, "static semantic version") {
		t.Errorf("message = %q, want it to mention 'static semantic version'", diags[0].Message)
	}
}

func TestMinimumVersionInvalidSemver(t *testing.T) {
	for _, bad := range []string{"0.8", "latest", "v0.8.0", ">=0.8,<1.0"} {
		src := "set minimum_version := \"" + bad + "\"\ngreet:\n    @echo hi\n"
		f, _ := parser.Parse("Runefile", src)
		_, diags := MinimumVersion(f)
		if !diags.HasErrors() {
			t.Errorf("value %q: expected an invalid-semver diagnostic", bad)
			continue
		}
		if !strings.Contains(diags[0].Message, "valid semantic version") {
			t.Errorf("value %q: message = %q, want it to mention 'valid semantic version'", bad, diags[0].Message)
		}
	}
}

// TestMinimumVersionCoexistsWithRuneVersion proves minimum_version and
// rune_version are independent settings (FR-013): both can be declared and each
// is read without disturbing the other.
func TestMinimumVersionCoexistsWithRuneVersion(t *testing.T) {
	src := "set rune_version := \"1\"\nset minimum_version := \"0.8.0\"\ngreet:\n    @echo hi\n"
	f, _ := parser.Parse("Runefile", src)

	if v := RuneVersion(f); v != "1" {
		t.Errorf("RuneVersion = %q, want 1", v)
	}
	if !Compatible(f) {
		t.Errorf("rune_version 1 should be compatible")
	}
	req, diags := MinimumVersion(f)
	if diags.HasErrors() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !req.Present || req.Raw != "0.8.0" {
		t.Errorf("minimum_version = %+v, want Raw=0.8.0", req)
	}
}
