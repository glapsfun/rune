package lsp

import (
	"context"
	"time"

	"github.com/rune-task-runner/rune/internal/analysis"
	"github.com/rune-task-runner/rune/internal/diag"
)

// scheduleAnalyze debounces analysis for a document: it cancels any pending or
// in-flight analysis and (re)arms a timer to analyze the latest version after
// the debounce interval (spec FR-016).
func (s *Server) scheduleAnalyze(path string, version int) {
	s.mu.Lock()
	if t, ok := s.timers[path]; ok {
		t.Stop()
	}
	if cancel, ok := s.cancels[path]; ok {
		cancel()
		delete(s.cancels, path)
	}
	s.timers[path] = time.AfterFunc(s.debounce, func() {
		s.analyzeAndPublish(path, version)
	})
	s.mu.Unlock()
}

// analyzeAndPublish analyzes the document and publishes diagnostics, but only if
// the analyzed version is still the current one — a superseded version's
// diagnostics are never published (spec FR-016).
func (s *Server) analyzeAndPublish(path string, version int) {
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	if old, ok := s.cancels[path]; ok {
		old()
	}
	s.cancels[path] = cancel
	s.mu.Unlock()
	defer cancel()

	snap, err := s.svc.Analyze(ctx, analysis.AnalyzeRequest{URI: path, Version: version})
	if err != nil {
		s.log.Printf("analyze %s: %v", path, err)
		return
	}

	// Version guard: drop results computed for an out-of-date document.
	s.mu.Lock()
	current, tracked := s.docs[path]
	s.mu.Unlock()
	if tracked && version != current {
		return
	}

	s.publishSnapshot(ctx, path, version, snap)
}

// publishSnapshot converts and publishes diagnostics, grouped by the file each
// diagnostic belongs to so cross-file diagnostics (FR-009a) land in the right
// document. The entry document is always published (an empty set clears it).
func (s *Server) publishSnapshot(ctx context.Context, entryPath string, version int, snap *analysis.Snapshot) {
	byFile := map[string][]diag.Diagnostic{entryPath: nil}
	for _, d := range snap.Diagnostics {
		file := d.Span.File
		if file == "" {
			file = entryPath
		}
		byFile[file] = append(byFile[file], d)
	}

	indexes := map[string]*LineIndex{}
	lineIndex := func(file string) *LineIndex {
		if ix, ok := indexes[file]; ok {
			return ix
		}
		text := ""
		if b, err := s.overlay.Read(ctx, file); err == nil {
			text = string(b)
		}
		ix := NewLineIndex(text)
		indexes[file] = ix
		return ix
	}

	for file, ds := range byFile {
		ix := lineIndex(file)
		lsps := make([]Diagnostic, 0, len(ds))
		for _, d := range ds {
			lsps = append(lsps, toLSPDiagnostic(ix, lineIndex, d))
		}
		params := PublishDiagnosticsParams{URI: pathToURI(file), Diagnostics: lsps}
		if file == entryPath {
			params.Version = version
		}
		s.notify("textDocument/publishDiagnostics", params)
	}
}

// toLSPDiagnostic converts a diag.Diagnostic to its LSP form. Related locations
// may point into other files, so each is converted with its own file's index
// (obtained via the indexFor accessor).
func toLSPDiagnostic(ix *LineIndex, indexFor func(string) *LineIndex, d diag.Diagnostic) Diagnostic {
	out := Diagnostic{
		Range:    ix.SpanToRange(d.Span),
		Severity: severityToLSP(d.Severity),
		Code:     d.Code,
		Source:   "rune",
		Message:  d.Message,
	}
	for _, r := range d.Related {
		rix := ix
		if r.Span.File != d.Span.File {
			rix = indexFor(r.Span.File)
		}
		out.RelatedInformation = append(out.RelatedInformation, DiagnosticRelatedInformation{
			Location: Location{URI: pathToURI(r.Span.File), Range: rix.SpanToRange(r.Span)},
			Message:  r.Message,
		})
	}
	return out
}

func severityToLSP(sev diag.Severity) int {
	if sev == diag.Warning {
		return SeverityWarning
	}
	return SeverityError
}
