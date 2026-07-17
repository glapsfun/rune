package analysis

import (
	"context"
	"os"
	"sync"

	"github.com/rune-task-runner/rune/internal/diag"
)

// DocumentURI identifies a document. At the analysis layer it is a cleaned
// filesystem path (what the parser and import composer already use); the LSP
// layer converts file:// URIs to and from this form at its boundary. It is a
// type alias so it is interchangeable with the path strings the engine uses.
type DocumentURI = string

// SourceStore resolves document content. It is READ-ONLY by contract: the
// analysis surface never writes project files (spec FR-028).
type SourceStore interface {
	Read(ctx context.Context, uri DocumentURI) ([]byte, error)
	Exists(ctx context.Context, uri DocumentURI) bool
}

// DiskSourceStore reads documents straight from the filesystem.
type DiskSourceStore struct{}

func (DiskSourceStore) Read(_ context.Context, uri DocumentURI) ([]byte, error) {
	return os.ReadFile(uri)
}

func (DiskSourceStore) Exists(_ context.Context, uri DocumentURI) bool {
	_, err := os.Stat(uri)
	return err == nil
}

// OverlaySourceStore serves open (possibly unsaved) editor documents in
// preference to disk, falling back to its disk backing for any document that is
// not open. This is the resolution rule from the spec (FR-003): use the editor
// overlay when a document is open, otherwise read from disk — and it applies to
// imported files too. It is safe for concurrent use.
type OverlaySourceStore struct {
	disk SourceStore

	mu        sync.RWMutex
	documents map[DocumentURI]OpenDocument
}

// NewOverlaySourceStore returns an overlay backed by disk (or the given store).
func NewOverlaySourceStore(disk SourceStore) *OverlaySourceStore {
	if disk == nil {
		disk = DiskSourceStore{}
	}
	return &OverlaySourceStore{disk: disk, documents: map[DocumentURI]OpenDocument{}}
}

// Set records (or replaces) an open document in the overlay.
func (o *OverlaySourceStore) Set(doc OpenDocument) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.documents[doc.URI] = doc
}

// Remove drops an open document, so subsequent reads fall back to disk
// (textDocument/didClose).
func (o *OverlaySourceStore) Remove(uri DocumentURI) {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.documents, uri)
}

// Get returns the open document for uri, if one is open.
func (o *OverlaySourceStore) Get(uri DocumentURI) (OpenDocument, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	d, ok := o.documents[uri]
	return d, ok
}

func (o *OverlaySourceStore) Read(ctx context.Context, uri DocumentURI) ([]byte, error) {
	if d, ok := o.Get(uri); ok {
		return []byte(d.Text), nil
	}
	return o.disk.Read(ctx, uri)
}

func (o *OverlaySourceStore) Exists(ctx context.Context, uri DocumentURI) bool {
	if _, ok := o.Get(uri); ok {
		return true
	}
	return o.disk.Exists(ctx, uri)
}

// Provider adapts a SourceStore to the diag.SourceProvider signature used by the
// parser diagnostics renderer and config.Compose. A read error maps to
// (nil, false).
func Provider(ctx context.Context, s SourceStore) diag.SourceProvider {
	return func(path string) ([]byte, bool) {
		b, err := s.Read(ctx, path)
		if err != nil {
			return nil, false
		}
		return b, true
	}
}
