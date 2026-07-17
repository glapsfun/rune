package lsp

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
)

// registerWatchers asks the client (via client/registerCapability) to watch
// Runefiles so the server is told when imported files change on disk, even when
// they are not open in the editor (spec FR-022). Sent after `initialized`. This
// is best-effort: clients that watch via their own configuration (e.g. the
// VS Code extension's synchronize.fileEvents) still deliver the notifications,
// and a client that lacks dynamic registration simply errors, which we ignore.
func (s *Server) registerWatchers() {
	s.mu.Lock()
	s.nextReqID++
	id := s.nextReqID
	s.mu.Unlock()

	// Scope the watch globs to the detected workspace root (FR-021) when known,
	// so watching is confined to the project rather than the whole filesystem.
	s.mu.Lock()
	root := s.root
	s.mu.Unlock()
	prefix := ""
	if root != "" {
		prefix = strings.TrimRight(filepath.ToSlash(root), "/") + "/"
	}

	raw := json.RawMessage(strconv.Itoa(id))
	params := RegistrationParams{Registrations: []Registration{{
		ID:     "rune-watch-runefiles",
		Method: "workspace/didChangeWatchedFiles",
		RegisterOptions: DidChangeWatchedFilesRegistrationOptions{Watchers: []FileSystemWatcher{
			{GlobPattern: prefix + "**/Runefile"},
			{GlobPattern: prefix + "**/.runefile"},
			{GlobPattern: prefix + "**/*.rune"},
		}},
	}}}
	pr, err := json.Marshal(params)
	if err != nil {
		s.log.Printf("marshal registration: %v", err)
		return
	}
	if err := s.conn.Write(&Message{JSONRPC: "2.0", ID: &raw, Method: "client/registerCapability", Params: pr}); err != nil {
		s.log.Printf("register watchers: %v", err)
	}
	// The client's response arrives as a message with an id and no method; the
	// dispatch loop ignores it (nothing to do on success).
}

// didChangeWatchedFiles re-analyzes and republishes every open document affected
// by an on-disk change to a watched file: the changed file itself (if open) and
// any open document that imports it (spec FR-022). Analysis is cheap, so this
// over-approximates safely for small workspaces.
func (s *Server) didChangeWatchedFiles(params json.RawMessage) {
	var p DidChangeWatchedFilesParams
	if err := json.Unmarshal(params, &p); err != nil {
		s.log.Printf("didChangeWatchedFiles: %v", err)
		return
	}
	changed := make(map[string]bool, len(p.Changes))
	for _, c := range p.Changes {
		changed[uriToPath(c.URI)] = true
	}
	if len(changed) == 0 {
		return
	}

	for _, path := range s.affectedOpenDocs(changed) {
		go s.analyzeAndPublish(path, s.getVersion(path))
	}
}

// affectedOpenDocs returns the open documents that should be re-analyzed given a
// set of changed files: a document is affected if it is itself changed, or its
// last snapshot's import graph shows it importing a changed file.
func (s *Server) affectedOpenDocs(changed map[string]bool) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []string
	for path := range s.docs {
		if changed[path] {
			out = append(out, path)
			continue
		}
		snap := s.snaps[path]
		if snap == nil {
			// No graph yet: re-analyze to be safe (it may import the changed file).
			out = append(out, path)
			continue
		}
		for y := range changed {
			affected := false
			for _, importer := range snap.Imports.TransitiveImporters(y) {
				if importer == path {
					affected = true
					break
				}
			}
			if affected {
				out = append(out, path)
				break
			}
		}
	}
	return out
}
