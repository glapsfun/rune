package analysis

import (
	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/diag"
	"github.com/rune-task-runner/rune/internal/language"
)

// Snapshot is the immutable result of analyzing one entry document at one
// version: the composed AST, a source provider for rendering, all diagnostics
// (parser + semantic + project, across files), the symbol index, and the import
// graph. Every interface (CLI, rune analyze, LSP, MCP) consumes this same shape
// so results are identical (spec FR-002).
type Snapshot struct {
	URI         DocumentURI
	Version     int
	File        *ast.File
	Sources     diag.SourceProvider
	Diagnostics diag.List
	Symbols     *language.Index
	Imports     ImportGraph
}

// HasErrors reports whether the snapshot contains any error-severity diagnostic.
func (s *Snapshot) HasErrors() bool { return s.Diagnostics.HasErrors() }
