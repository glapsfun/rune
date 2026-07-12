package lsp

import (
	"context"
	"encoding/json"

	"github.com/rune-task-runner/rune/internal/analysis"
	"github.com/rune-task-runner/rune/internal/formatter"
	"github.com/rune-task-runner/rune/internal/language"
	"github.com/rune-task-runner/rune/internal/parser"
)

// snapshotFor analyzes the current content of the document at path and returns
// the snapshot plus the raw text (for position conversion). Analysis is cheap
// (~microseconds), so it is done per request rather than cached.
func (s *Server) snapshotFor(ctx context.Context, path string) (*analysis.Snapshot, string, error) {
	text := ""
	if b, err := s.overlay.Read(ctx, path); err == nil {
		text = string(b)
	}
	snap, err := s.svc.Analyze(ctx, analysis.AnalyzeRequest{URI: path, Version: s.getVersion(path)})
	return snap, text, err
}

// offsetAt converts an LSP position in the document text to a byte offset.
func offsetAt(text string, pos Position) int {
	off, _ := NewLineIndex(text).PositionToByteOffset(pos)
	return off
}

// definition handles textDocument/definition.
func (s *Server) definition(id *json.RawMessage, params json.RawMessage) {
	var p TextDocumentPositionParams
	if err := json.Unmarshal(params, &p); err != nil {
		s.reply(id, nil)
		return
	}
	ctx := context.Background()
	path := uriToPath(p.TextDocument.URI)
	snap, text, err := s.snapshotFor(ctx, path)
	if err != nil {
		s.reply(id, nil)
		return
	}
	offset := offsetAt(text, p.Position)
	spans, ok := language.Definition(snap.Symbols, snap.File, path, offset)
	if !ok {
		s.reply(id, nil)
		return
	}
	locs := make([]Location, 0, len(spans))
	for _, sp := range spans {
		fileText := s.fileText(ctx, sp.File)
		fix := NewLineIndex(fileText)
		locs = append(locs, Location{URI: pathToURI(sp.File), Range: fix.SpanToRange(sp)})
	}
	s.reply(id, locs)
}

// hover handles textDocument/hover.
func (s *Server) hover(id *json.RawMessage, params json.RawMessage) {
	var p TextDocumentPositionParams
	if err := json.Unmarshal(params, &p); err != nil {
		s.reply(id, nil)
		return
	}
	ctx := context.Background()
	path := uriToPath(p.TextDocument.URI)
	snap, text, err := s.snapshotFor(ctx, path)
	if err != nil {
		s.reply(id, nil)
		return
	}
	offset := offsetAt(text, p.Position)
	md, span, ok := language.Hover(snap.File, path, offset)
	if !ok {
		s.reply(id, nil)
		return
	}
	r := NewLineIndex(text).SpanToRange(span)
	s.reply(id, Hover{Contents: MarkupContent{Kind: "markdown", Value: md}, Range: &r})
}

// formatting handles textDocument/formatting: it returns a single full-document
// edit with the canonical formatting of the current (possibly unsaved) buffer.
// If the buffer has parse errors, no edit is returned rather than corrupting it
// (spec FR-020). The formatter is called directly — no child process, no write.
func (s *Server) formatting(id *json.RawMessage, params json.RawMessage) {
	var p DocumentFormattingParams
	if err := json.Unmarshal(params, &p); err != nil {
		s.reply(id, nil)
		return
	}
	ctx := context.Background()
	path := uriToPath(p.TextDocument.URI)
	text := s.fileText(ctx, path)

	file, diags := parser.Parse(path, text)
	if diags.HasErrors() {
		s.reply(id, []TextEdit{}) // don't format un-parseable content
		return
	}
	formatted := formatter.Format(file)
	if formatted == text {
		s.reply(id, []TextEdit{}) // already canonical: no-op
		return
	}
	edit := TextEdit{Range: fullRange(text), NewText: formatted}
	s.reply(id, []TextEdit{edit})
}

// fileText reads a document's current content (overlay then disk).
func (s *Server) fileText(ctx context.Context, path string) string {
	if b, err := s.overlay.Read(ctx, path); err == nil {
		return string(b)
	}
	return ""
}

// fullRange returns a range covering the entire document.
func fullRange(text string) Range {
	ix := NewLineIndex(text)
	return Range{
		Start: Position{Line: 0, Character: 0},
		End:   ix.ByteOffsetToPosition(len(text)),
	}
}
