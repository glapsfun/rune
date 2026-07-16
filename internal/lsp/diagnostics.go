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

	// Version guard: drop results for a document that has since been closed
	// (racing didClose already cleared it) or superseded by a newer version.
	s.mu.Lock()
	current, tracked := s.docs[path]
	if tracked && version == current {
		s.snaps[path] = snap // cache for import-graph lookups (watched files)
	}
	s.mu.Unlock()
	if !tracked || version != current {
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
		if v, ok := s.publishVersion(file, entryPath, version); ok {
			params.Version = &v
		}
		s.notify("textDocument/publishDiagnostics", params)
	}

	// Clear diagnostics on files this entry published to previously but no longer
	// does (an import was removed, or an imported file's error was fixed) — those
	// URIs would otherwise keep showing stale squiggles. Files that are open in
	// their own right manage their own diagnostics, so leave them untouched.
	s.mu.Lock()
	cur := make(map[string]bool, len(byFile))
	for file := range byFile {
		cur[file] = true
	}
	var stale []string
	for file := range s.published[entryPath] {
		if cur[file] {
			continue
		}
		if _, open := s.docs[file]; open && file != entryPath {
			continue
		}
		// Another open document may still legitimately flag this imported file;
		// clearing here would wipe diagnostics that entry still owns (and would
		// not be re-triggered), so only clear when no other entry publishes to it.
		ownedElsewhere := false
		for entry, files := range s.published {
			if entry != entryPath && files[file] {
				ownedElsewhere = true
				break
			}
		}
		if ownedElsewhere {
			continue
		}
		stale = append(stale, file)
	}
	s.published[entryPath] = cur
	s.mu.Unlock()

	for _, file := range stale {
		s.notify("textDocument/publishDiagnostics", PublishDiagnosticsParams{
			URI:         pathToURI(file),
			Diagnostics: []Diagnostic{},
		})
	}
}

// publishVersion returns the version to attach to a publishDiagnostics for file.
// The entry document uses the analyzed version; another open document uses its
// current tracked version; a file that is not open carries no version (false).
func (s *Server) publishVersion(file, entryPath string, entryVersion int) (int, bool) {
	if file == entryPath {
		return entryVersion, true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.docs[file]
	return v, ok
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
