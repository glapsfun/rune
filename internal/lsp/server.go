package lsp

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"

	"github.com/rune-task-runner/rune/internal/analysis"
)

// Server is the Rune language server: a JSON-RPC/LSP 3.17 server over a framed
// connection. It holds open documents in an overlay and drives the shared
// analysis service. It executes nothing (spec FR-028).
type Server struct {
	conn    *Conn
	svc     *analysis.Service
	overlay *analysis.OverlaySourceStore
	log     *log.Logger
	version string // reported in serverInfo

	debounce time.Duration

	mu         sync.Mutex
	docs       map[string]int // path -> current version
	timers     map[string]*time.Timer
	cancels    map[string]context.CancelFunc
	shutdownOK bool
}

// Options configures a Server.
type Options struct {
	// Version is reported in the initialize response's serverInfo.
	Version string
	// LogWriter receives server logs (stderr or a file); never stdout.
	LogWriter io.Writer
	// Debounce is the delay before analyzing after a change (default 100ms).
	Debounce time.Duration
}

// NewServer builds a server over r/w (stdin/stdout for a real client).
func NewServer(r io.Reader, w io.Writer, opts Options) *Server {
	overlay := analysis.NewOverlaySourceStore(analysis.DiskSourceStore{})
	logw := opts.LogWriter
	if logw == nil {
		logw = io.Discard
	}
	debounce := opts.Debounce
	if debounce == 0 {
		debounce = 100 * time.Millisecond
	}
	return &Server{
		conn:     NewConn(r, w),
		svc:      analysis.NewService(overlay),
		overlay:  overlay,
		log:      log.New(logw, "rune-lsp ", log.LstdFlags),
		version:  opts.Version,
		debounce: debounce,
		docs:     map[string]int{},
		timers:   map[string]*time.Timer{},
		cancels:  map[string]context.CancelFunc{},
	}
}

// Run serves the connection until exit or EOF. It returns nil on a clean
// shutdown+exit and an error if the stream fails unexpectedly.
func (s *Server) Run() error {
	for {
		msg, err := s.conn.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			// A malformed message must not crash the server: log and continue.
			s.log.Printf("read error: %v", err)
			if isFatalRead(err) {
				return nil
			}
			continue
		}
		if stop := s.dispatch(msg); stop {
			return nil
		}
	}
}

// isFatalRead reports whether a read error means the stream is unusable (so the
// loop should stop rather than spin). Framing/parse errors are recoverable; an
// underlying I/O failure is not.
func isFatalRead(err error) bool {
	var re *ResponseError
	return !asResponseError(err, &re) // non-protocol errors are treated as fatal
}

func asResponseError(err error, target **ResponseError) bool {
	if re, ok := err.(*ResponseError); ok {
		*target = re
		return true
	}
	return false
}

// initialize handles the initialize request, advertising only implemented
// capabilities (spec FR-014).
func (s *Server) initialize(id *json.RawMessage, params json.RawMessage) {
	var p InitializeParams
	_ = json.Unmarshal(params, &p) // fields are optional; ignore decode issues

	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: &TextDocumentSyncOptions{
				OpenClose: true,
				Change:    SyncIncremental,
				Save:      &SaveOptions{IncludeText: false},
			},
			CompletionProvider: &CompletionOptions{TriggerCharacters: []string{"[", "(", ".", "{", ":"}},
			DefinitionProvider: true,
			HoverProvider:      true,
			DocumentFormatting: true,
			DocumentSymbol:     true,
		},
		ServerInfo: ServerInfo{Name: "rune", Version: s.version},
	}
	s.reply(id, result)
}

// shutdown marks the server ready to exit; the client then sends `exit`.
func (s *Server) shutdown(id *json.RawMessage) {
	s.mu.Lock()
	s.shutdownOK = true
	s.mu.Unlock()
	s.reply(id, nil)
}

// --- reply helpers ---

func (s *Server) reply(id *json.RawMessage, result any) {
	resp, err := NewResponse(id, result)
	if err != nil {
		s.log.Printf("marshal response: %v", err)
		s.conn.Write(NewErrorResponse(id, InternalError, err.Error()))
		return
	}
	if err := s.conn.Write(resp); err != nil {
		s.log.Printf("write response: %v", err)
	}
}

func (s *Server) notify(method string, params any) {
	n, err := NewNotification(method, params)
	if err != nil {
		s.log.Printf("marshal notification %s: %v", method, err)
		return
	}
	if err := s.conn.Write(n); err != nil {
		s.log.Printf("write notification %s: %v", method, err)
	}
}
