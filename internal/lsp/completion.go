package lsp

import (
	"context"
	"encoding/json"

	"github.com/rune-task-runner/rune/internal/language"
)

// completion handles textDocument/completion, returning context-aware
// suggestions from the language layer.
func (s *Server) completion(id *json.RawMessage, params json.RawMessage) {
	var p CompletionParams
	if err := json.Unmarshal(params, &p); err != nil {
		s.reply(id, nil)
		return
	}
	ctx := context.Background()
	path := uriToPath(p.TextDocument.URI)
	snap, text, err := s.snapshotFor(ctx, path)
	if err != nil {
		s.reply(id, []CompletionItem{})
		return
	}
	offset := offsetAt(text, p.Position)
	items := language.Complete(snap.Symbols, snap.File, path, text, offset)

	out := make([]CompletionItem, 0, len(items))
	for _, it := range items {
		out = append(out, CompletionItem{
			Label:         it.Label,
			Kind:          completionKind(it.Kind),
			Detail:        it.Detail,
			Documentation: it.Documentation,
		})
	}
	s.reply(id, out)
}

func completionKind(k language.CompletionKind) int {
	switch k {
	case language.CompletionTask:
		return CIKMethod
	case language.CompletionVariable:
		return CIKVariable
	case language.CompletionParameter:
		return CIKVariable
	case language.CompletionSetting:
		return CIKProperty
	case language.CompletionAttribute:
		return CIKKeyword
	case language.CompletionExecutor:
		return CIKEnum
	case language.CompletionFunction:
		return CIKFunction
	default:
		return 0
	}
}
