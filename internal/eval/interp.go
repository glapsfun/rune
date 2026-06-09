package eval

import (
	"strings"

	"github.com/rune-task-runner/rune/internal/parser"
	"github.com/rune-task-runner/rune/internal/token"
)

// Interpolate expands {{ expr }} placeholders in a body line's raw text using
// the evaluator. A literal "{{" is written as "{{{{" (and "}}" as "}}}}").
// baseSpan locates the line for diagnostics; expression spans are approximate
// (anchored to the line) since body text is not pre-tokenized.
func (e *Evaluator) Interpolate(raw string, baseSpan token.Span) (string, *Error) {
	var b strings.Builder
	i := 0
	for i < len(raw) {
		// Escaped braces: {{{{ -> {{ , }}}} -> }}
		if strings.HasPrefix(raw[i:], "{{{{") {
			b.WriteString("{{")
			i += 4
			continue
		}
		if strings.HasPrefix(raw[i:], "}}}}") {
			b.WriteString("}}")
			i += 4
			continue
		}
		if strings.HasPrefix(raw[i:], "{{") {
			end := strings.Index(raw[i+2:], "}}")
			if end < 0 {
				return "", &Error{Span: baseSpan, Msg: "unterminated interpolation: missing '}}'"}
			}
			exprSrc := raw[i+2 : i+2+end]
			val, err := e.evalInterpExpr(exprSrc, baseSpan)
			if err != nil {
				return "", err
			}
			b.WriteString(val)
			i = i + 2 + end + 2
			continue
		}
		b.WriteByte(raw[i])
		i++
	}
	return b.String(), nil
}

// evalInterpExpr parses and evaluates a single interpolation expression. It
// reuses the real expression parser by wrapping the fragment in an assignment.
func (e *Evaluator) evalInterpExpr(src string, baseSpan token.Span) (string, *Error) {
	src = strings.TrimSpace(src)
	if src == "" {
		return "", &Error{Span: baseSpan, Msg: "empty interpolation '{{}}'"}
	}
	expr, diags := parser.ParseExprFragment(baseSpan.File, src)
	if diags.HasErrors() {
		return "", &Error{Span: baseSpan, Msg: "invalid interpolation expression: " + diags[0].Message}
	}
	return e.Eval(expr)
}
