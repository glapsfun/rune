package lsp

import (
	"encoding/json"

	"github.com/rune-task-runner/rune/internal/analysis"
)

func (s *Server) didOpen(params json.RawMessage) {
	var p DidOpenTextDocumentParams
	if err := json.Unmarshal(params, &p); err != nil {
		s.log.Printf("didOpen: %v", err)
		return
	}
	path := uriToPath(p.TextDocument.URI)
	s.overlay.Set(analysis.OpenDocument{URI: path, Version: p.TextDocument.Version, Text: p.TextDocument.Text})
	s.setVersion(path, p.TextDocument.Version)
	// Analyze immediately on open so diagnostics appear at once.
	go s.analyzeAndPublish(path, p.TextDocument.Version)
}

func (s *Server) didChange(params json.RawMessage) {
	var p DidChangeTextDocumentParams
	if err := json.Unmarshal(params, &p); err != nil {
		s.log.Printf("didChange: %v", err)
		return
	}
	path := uriToPath(p.TextDocument.URI)
	text := ""
	if doc, ok := s.overlay.Get(path); ok {
		text = doc.Text
	}
	text = applyChanges(text, p.ContentChanges)
	s.overlay.Set(analysis.OpenDocument{URI: path, Version: p.TextDocument.Version, Text: text})
	s.setVersion(path, p.TextDocument.Version)
	// Debounce: rapid edits collapse to one analysis of the latest version.
	s.scheduleAnalyze(path, p.TextDocument.Version)
}

func (s *Server) didSave(params json.RawMessage) {
	var p DidSaveTextDocumentParams
	if err := json.Unmarshal(params, &p); err != nil {
		s.log.Printf("didSave: %v", err)
		return
	}
	path := uriToPath(p.TextDocument.URI)
	go s.analyzeAndPublish(path, s.getVersion(path))
}

func (s *Server) didClose(params json.RawMessage) {
	var p DidCloseTextDocumentParams
	if err := json.Unmarshal(params, &p); err != nil {
		s.log.Printf("didClose: %v", err)
		return
	}
	path := uriToPath(p.TextDocument.URI)
	s.overlay.Remove(path)
	s.clearDoc(path)
	// Clear diagnostics for the closed document under the canonical URI that
	// publishSnapshot used to publish them (pathToURI(uriToPath(...))); the raw
	// client URI may differ (e.g. a redundant ./) and would leave stale squiggles.
	s.notify("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         pathToURI(path),
		Diagnostics: []Diagnostic{},
	})
}

// applyChanges applies LSP content changes to text. A change with no Range is a
// full replacement; an incremental change splices at the byte offsets its range
// maps to (via a per-change LineIndex, since each edit shifts the text).
func applyChanges(text string, changes []TextDocumentContentChangeEvent) string {
	for _, ch := range changes {
		if ch.Range == nil {
			text = ch.Text
			continue
		}
		ix := NewLineIndex(text)
		start, _ := ix.PositionToByteOffset(ch.Range.Start)
		end, _ := ix.PositionToByteOffset(ch.Range.End)
		text = analysis.ApplyByteEdit(text, start, end, ch.Text)
	}
	return text
}

// --- version bookkeeping ---

func (s *Server) setVersion(path string, version int) {
	s.mu.Lock()
	s.docs[path] = version
	s.mu.Unlock()
}

func (s *Server) getVersion(path string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.docs[path]
}

func (s *Server) clearDoc(path string) {
	s.mu.Lock()
	delete(s.docs, path)
	delete(s.snaps, path)
	delete(s.published, path)
	if t, ok := s.timers[path]; ok {
		t.Stop()
		delete(s.timers, path)
	}
	if cancel, ok := s.cancels[path]; ok {
		cancel()
		delete(s.cancels, path)
	}
	s.mu.Unlock()
}
