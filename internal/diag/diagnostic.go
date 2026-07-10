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

// RelatedLocation is a secondary span attached to a diagnostic, such as each
// task/file participating in a cycle (spec FR-009). Message is optional context
// (e.g. "build depends on test").
type RelatedLocation struct {
	Span    token.Span
	Message string
}

// Diagnostic is a single located message. Code, when set, is a stable public
// identifier from the RUNE#### catalog (spec FR-010); it is empty for legacy or
// uncoded emits. Related carries any secondary locations (FR-009).
type Diagnostic struct {
	Severity Severity
	Span     token.Span
	Message  string
	Code     string
	Related  []RelatedLocation
}

// New builds an error diagnostic.
func New(span token.Span, msg string) Diagnostic {
	return Diagnostic{Severity: Error, Span: span, Message: msg}
}

// Warn builds a warning diagnostic.
func Warn(span token.Span, msg string) Diagnostic {
	return Diagnostic{Severity: Warning, Span: span, Message: msg}
}

// WithCode returns a copy of d carrying the given stable diagnostic code.
func (d Diagnostic) WithCode(code string) Diagnostic {
	d.Code = code
	return d
}

// WithRelated returns a copy of d with the given related locations attached.
func (d Diagnostic) WithRelated(rel ...RelatedLocation) Diagnostic {
	d.Related = append(d.Related, rel...)
	return d
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
