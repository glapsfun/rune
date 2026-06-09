package diag

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/fatih/color"

	"github.com/rune-task-runner/rune/internal/token"
)

// sprintf is an indirection so diagnostic.go need not import fmt directly.
func sprintf(format string, args ...any) string { return fmt.Sprintf(format, args...) }

// SourceProvider returns the raw bytes of a file referenced by a diagnostic
// span, plus whether the file was found.
type SourceProvider func(file string) ([]byte, bool)

// Render formats a single diagnostic. When source is non-nil the offending line
// is shown beneath the header with a caret underline covering the span. When
// useColor is true, the severity word and caret are colorized.
func Render(d Diagnostic, source []byte, useColor bool) string {
	var b bytes.Buffer

	sev := d.Severity.String()
	loc := d.Span.String()
	if useColor {
		c := color.New(color.FgRed, color.Bold)
		if d.Severity == Warning {
			c = color.New(color.FgYellow, color.Bold)
		}
		fmt.Fprintf(&b, "%s: %s: %s", loc, c.Sprint(sev), d.Message)
	} else {
		fmt.Fprintf(&b, "%s: %s: %s", loc, sev, d.Message)
	}

	if source != nil && d.Span.Start.Line > 0 {
		if snippet := renderSnippet(d.Span, source, useColor); snippet != "" {
			b.WriteByte('\n')
			b.WriteString(snippet)
		}
	}
	return b.String()
}

// RenderAll formats every diagnostic in the list (one per stanza) using src to
// fetch the source for each diagnostic's file.
func RenderAll(list List, src SourceProvider, useColor bool) string {
	var b strings.Builder
	for i, d := range list {
		var source []byte
		if src != nil {
			if data, ok := src(d.Span.File); ok {
				source = data
			}
		}
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(Render(d, source, useColor))
	}
	return b.String()
}

// renderSnippet builds a "  N | <line>" gutter row plus a caret underline that
// spans Start.Col..End.Col on that line.
func renderSnippet(span token.Span, source []byte, useColor bool) string {
	line := sourceLine(source, span.Start.Line)
	gutter := fmt.Sprintf("%d", span.Start.Line)
	pad := strings.Repeat(" ", len(gutter))

	var b bytes.Buffer
	fmt.Fprintf(&b, "%s | %s\n", gutter, line)
	fmt.Fprintf(&b, "%s | %s", pad, caretUnderline(line, span, useColor))
	return b.String()
}

// sourceLine returns the 1-based nth line of source without its terminator.
func sourceLine(source []byte, n int) string {
	if n <= 0 {
		return ""
	}
	lines := bytes.Split(source, []byte("\n"))
	if n > len(lines) {
		return ""
	}
	return strings.TrimRight(string(lines[n-1]), "\r")
}

// caretUnderline produces leading whitespace (preserving tabs from the source
// line for alignment) up to Start.Col, then a run of carets covering the span.
func caretUnderline(line string, span token.Span, useColor bool) string {
	startCol := span.Start.Col
	if startCol < 1 {
		startCol = 1
	}
	width := 1
	if span.End.Line == span.Start.Line && span.End.Col > span.Start.Col {
		width = span.End.Col - span.Start.Col
	}

	var lead bytes.Buffer
	for i := 0; i < startCol-1; i++ {
		if i < len(line) && line[i] == '\t' {
			lead.WriteByte('\t')
		} else {
			lead.WriteByte(' ')
		}
	}
	carets := strings.Repeat("^", width)
	if useColor {
		carets = color.New(color.FgRed, color.Bold).Sprint(carets)
	}
	return lead.String() + carets
}
