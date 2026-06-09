package config

import (
	"testing"

	"github.com/rune-task-runner/rune/internal/parser"
)

func TestRuneVersionPragma(t *testing.T) {
	f, _ := parser.Parse("Runefile", "set rune_version := \"1\"\ngreet:\n    @echo hi\n")
	if v := RuneVersion(f); v != "1" {
		t.Errorf("RuneVersion = %q, want 1", v)
	}
	if !Compatible(f) {
		t.Error("version 1 should be compatible")
	}
}

func TestNoPragmaUsesDefaultInterpretation(t *testing.T) {
	// A file with no pragma must always be interpreted with the default (current)
	// semantics — interpretation never changes under a user (FR-033).
	f, _ := parser.Parse("Runefile", "greet:\n    @echo hi\n")
	if v := RuneVersion(f); v != "" {
		t.Errorf("RuneVersion = %q, want empty", v)
	}
	if !Compatible(f) {
		t.Error("a file with no pragma must be compatible")
	}
}

func TestUnknownVersionIncompatible(t *testing.T) {
	f, _ := parser.Parse("Runefile", "set rune_version := \"999\"\ngreet:\n    @echo hi\n")
	if Compatible(f) {
		t.Error("an unknown future version should be flagged incompatible")
	}
}
