package lexer

import (
	"strings"

	"github.com/rune-task-runner/rune/internal/diag"
	"github.com/rune-task-runner/rune/internal/token"
)

// lexString scans a string literal: single ('...', raw), double ("...", with
// escapes), or their triple-quoted multi-line forms (de-dented). The decoded
// value is stored in the STRING token's literal.
func (l *lexer) lexString() {
	start := l.pin()
	quote := l.at(l.pos)
	triple := l.at(l.pos+1) == quote && l.at(l.pos+2) == quote

	if triple {
		l.lexTripleString(start, quote)
		return
	}

	l.advance() // opening quote
	var b strings.Builder
	for {
		c := l.at(l.pos)
		if c == 0 || c == '\n' {
			l.diags.Codef(diag.CodeUnterminatedStr, token.Span{File: l.file, Start: start, End: l.pin()}, "unterminated string literal")
			l.emit(token.STRING, b.String(), start, l.pin())
			return
		}
		if c == quote {
			l.advance()
			l.emit(token.STRING, b.String(), start, l.pin())
			return
		}
		if quote == '"' && c == '\\' {
			l.advance() // backslash
			// A backslash with nothing after it (EOF) is a dangling escape: the
			// string is unterminated. Advancing again here would read past the
			// end of src (advance() is not bounds-checked).
			if l.pos >= len(l.src) {
				l.diags.Codef(diag.CodeUnterminatedStr, token.Span{File: l.file, Start: start, End: l.pin()}, "unterminated string literal")
				l.emit(token.STRING, b.String(), start, l.pin())
				return
			}
			b.WriteByte(unescape(l.at(l.pos)))
			l.advance() // escaped char
			continue
		}
		b.WriteByte(c)
		l.advance()
	}
}

// lexTripleString scans a triple-quoted string and de-dents its content.
func (l *lexer) lexTripleString(start token.Position, quote byte) {
	l.advance()
	l.advance()
	l.advance() // the three opening quotes
	begin := l.pos
	for {
		c := l.at(l.pos)
		if c == 0 {
			l.diags.Codef(diag.CodeUnterminatedStr, token.Span{File: l.file, Start: start, End: l.pin()}, "unterminated triple-quoted string")
			l.emit(token.STRING, dedent(l.src[begin:l.pos]), start, l.pin())
			return
		}
		if c == quote && l.at(l.pos+1) == quote && l.at(l.pos+2) == quote {
			raw := l.src[begin:l.pos]
			l.advance()
			l.advance()
			l.advance()
			value := dedent(raw)
			if quote == '"' {
				value = applyEscapes(value)
			}
			l.emit(token.STRING, value, start, l.pin())
			return
		}
		l.advance()
	}
}

// unescape decodes a single escape character following a backslash.
func unescape(c byte) byte {
	switch c {
	case 'n':
		return '\n'
	case 't':
		return '\t'
	case 'r':
		return '\r'
	case '0':
		return 0
	default:
		return c // \\ \" \' and any other => literal
	}
}

// applyEscapes decodes backslash escapes in an already-extracted string body.
func applyEscapes(s string) string {
	if !strings.ContainsRune(s, '\\') {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			b.WriteByte(unescape(s[i+1]))
			i++
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// dedent removes a leading and trailing blank line and strips the common
// leading indentation shared by all non-blank lines (the triple-string rule).
func dedent(s string) string {
	lines := strings.Split(s, "\n")
	// Drop a single leading blank line.
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	// Drop a single trailing blank line.
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	// Find the minimum indentation among non-blank lines.
	minIndent := -1
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		n := len(ln) - len(strings.TrimLeft(ln, " \t"))
		if minIndent == -1 || n < minIndent {
			minIndent = n
		}
	}
	if minIndent > 0 {
		for i, ln := range lines {
			if len(ln) >= minIndent {
				lines[i] = ln[minIndent:]
			}
		}
	}
	return strings.Join(lines, "\n")
}
