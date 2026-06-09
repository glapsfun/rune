// Package token defines the lexical tokens of the Runefile language and the
// source-position types that every token and AST node carries. Precise
// positions are the foundation of Principle II ("errors are a feature"): every
// diagnostic renders file:line:col with a caret-underlined span.
package token

import "fmt"

// Position is a location in a source file: a 0-based byte offset plus a 1-based
// line and column (column counted in bytes from the start of the line).
type Position struct {
	Offset int // 0-based byte offset into the file
	Line   int // 1-based line number
	Col    int // 1-based column number (bytes)
}

// String renders the position as line:col (used in diagnostics).
func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}

// Span is a half-open range [Start, End) within a single file. It is attached to
// every AST node and every Diagnostic.
type Span struct {
	File  string
	Start Position
	End   Position
}

// String renders the span as file:line:col anchored at its start.
func (s Span) String() string {
	if s.File == "" {
		return s.Start.String()
	}
	return fmt.Sprintf("%s:%s", s.File, s.Start)
}

// IsValid reports whether the span has been populated.
func (s Span) IsValid() bool {
	return s.Start.Line > 0
}

// To returns a span covering from s.Start to other.End (same file as s).
func (s Span) To(other Span) Span {
	return Span{File: s.File, Start: s.Start, End: other.End}
}
