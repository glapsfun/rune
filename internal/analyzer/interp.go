package analyzer

import (
	"strings"

	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/diag"
	"github.com/rune-task-runner/rune/internal/parser"
	"github.com/rune-task-runner/rune/internal/token"
)

// interpSeg is one {{ ... }} interpolation found in a body line, with its
// expression text and a source span pointing at the {{ ... }} region.
type interpSeg struct {
	text string
	span token.Span
}

// extractInterps scans a body line's raw text for {{ ... }} interpolations,
// honoring {{{{ / }}}} brace escapes. Spans are computed from base (the source
// position of raw[0]); body lines are single physical lines, so offset and
// column advance together.
func extractInterps(raw string, base token.Span) []interpSeg {
	var segs []interpSeg
	i := 0
	for i < len(raw) {
		if strings.HasPrefix(raw[i:], "{{{{") {
			i += 4
			continue
		}
		if strings.HasPrefix(raw[i:], "}}}}") {
			i += 4
			continue
		}
		if strings.HasPrefix(raw[i:], "{{") {
			end := strings.Index(raw[i+2:], "}}")
			if end < 0 {
				return segs
			}
			text := raw[i+2 : i+2+end]
			start := token.Position{
				Offset: base.Start.Offset + i,
				Line:   base.Start.Line,
				Col:    base.Start.Col + i,
			}
			closeAt := i + 2 + end + 2
			fin := token.Position{
				Offset: base.Start.Offset + closeAt,
				Line:   base.Start.Line,
				Col:    base.Start.Col + closeAt,
			}
			segs = append(segs, interpSeg{
				text: text,
				span: token.Span{File: base.File, Start: start, End: fin},
			})
			i = closeAt
			continue
		}
		i++
	}
	return segs
}

// parseFragment parses an interpolation expression fragment.
func parseFragment(text string, span token.Span) (ast.Expr, diag.List) {
	return parser.ParseExprFragment(span.File, text)
}
