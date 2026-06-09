package config

import (
	"github.com/rune-task-runner/rune/internal/ast"
)

// CurrentVersion is the Runefile semantics version this build implements.
// Breaking changes are opt-in per file via `set rune_version := "..."`; the
// default interpretation never changes under a user (FR-033).
const CurrentVersion = "1"

// RuneVersion returns the version pragma declared by a file, or "" if none.
func RuneVersion(f *ast.File) string {
	for _, s := range f.Settings {
		if s.Name == "rune_version" {
			if lit, ok := s.Value.(*ast.StringLit); ok {
				return lit.Value
			}
		}
	}
	return ""
}

// Compatible reports whether this build can interpret the file's declared
// version. An empty pragma (the common case) is always compatible and uses the
// default interpretation.
func Compatible(f *ast.File) bool {
	v := RuneVersion(f)
	return v == "" || v == CurrentVersion
}
