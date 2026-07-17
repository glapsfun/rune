package lsp

import (
	"context"
	"encoding/json"

	"github.com/rune-task-runner/rune/internal/language"
)

// documentSymbol handles textDocument/documentSymbol, returning a hierarchical
// outline (categories as parents, declarations as children).
func (s *Server) documentSymbol(id *json.RawMessage, params json.RawMessage) {
	var p DocumentSymbolParams
	if err := json.Unmarshal(params, &p); err != nil {
		s.reply(id, nil)
		return
	}
	ctx := context.Background()
	path := uriToPath(p.TextDocument.URI)
	snap, text, err := s.snapshotFor(ctx, path)
	if err != nil {
		s.reply(id, []DocumentSymbol{})
		return
	}
	ix := NewLineIndex(text)
	outline := language.Outline(snap.File, path)
	language.SortOutlineByPosition(outline)

	out := make([]DocumentSymbol, 0, len(outline.Groups))
	for _, g := range outline.Groups {
		children := make([]DocumentSymbol, 0, len(g.Entries))
		for _, e := range g.Entries {
			r := ix.SpanToRange(e.Selection)
			children = append(children, DocumentSymbol{
				Name:           e.Name,
				Detail:         e.Detail,
				Kind:           symbolKind(g.Kind),
				Range:          r,
				SelectionRange: r,
			})
		}
		gr := groupRange(children)
		out = append(out, DocumentSymbol{
			Name:           g.Name,
			Kind:           SKNamespace,
			Range:          gr,
			SelectionRange: firstSelection(children, gr),
			Children:       children,
		})
	}
	s.reply(id, out)
}

func symbolKind(k language.SymbolKind) int {
	switch k {
	case language.SymbolTask:
		return SKFunction
	case language.SymbolVariable:
		return SKVariable
	case language.SymbolSetting:
		return SKProperty
	case language.SymbolImport, language.SymbolModule:
		return SKModule
	default:
		return SKNamespace
	}
}

// groupRange spans from the first child's start to the last child's end so the
// group node encloses its children (LSP requires children ⊆ parent range).
func groupRange(children []DocumentSymbol) Range {
	if len(children) == 0 {
		return Range{}
	}
	r := children[0].Range
	for _, c := range children[1:] {
		if positionLess(c.Range.End, r.End) {
			continue
		}
		r.End = c.Range.End
	}
	return r
}

func firstSelection(children []DocumentSymbol, fallback Range) Range {
	if len(children) > 0 {
		return children[0].SelectionRange
	}
	return fallback
}

func positionLess(a, b Position) bool {
	if a.Line != b.Line {
		return a.Line < b.Line
	}
	return a.Character < b.Character
}
