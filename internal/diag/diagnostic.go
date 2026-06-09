// Package diag is the diagnostic model and renderer. Per Principle II, every
// statically detectable error carries a file:line:col span and renders the
// offending source line with a caret-underlined range.
package diag

import "github.com/rune-task-runner/rune/internal/token"

// Severity classifies a diagnostic.
type Severity int

const (
	Error Severity = iota
	Warning
)

func (s Severity) String() string {
	if s == Warning {
		return "warning"
	}
	return "error"
}

// Diagnostic is a single located message.
type Diagnostic struct {
	Severity Severity
	Span     token.Span
	Message  string
}

// New builds an error diagnostic.
func New(span token.Span, msg string) Diagnostic {
	return Diagnostic{Severity: Error, Span: span, Message: msg}
}

// Warn builds a warning diagnostic.
func Warn(span token.Span, msg string) Diagnostic {
	return Diagnostic{Severity: Warning, Span: span, Message: msg}
}

// List is an ordered collection of diagnostics.
type List []Diagnostic

// Add appends a diagnostic.
func (l *List) Add(d Diagnostic) { *l = append(*l, d) }

// Errorf appends an error diagnostic with a formatted message.
func (l *List) Errorf(span token.Span, format string, args ...any) {
	l.Add(Diagnostic{Severity: Error, Span: span, Message: sprintf(format, args...)})
}

// HasErrors reports whether any diagnostic is an error.
func (l List) HasErrors() bool {
	for _, d := range l {
		if d.Severity == Error {
			return true
		}
	}
	return false
}
