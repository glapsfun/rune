package lsp

// Minimal typed subset of LSP 3.17 — only the payloads for the methods this
// server implements (spec FR-014: advertise/handle only what is implemented).
// Position and Range are defined in convert.go.

// --- lifecycle ---

// InitializeParams is the (partial) client initialize request. Only the fields
// the server uses are modeled; unknown fields are ignored by encoding/json.
type InitializeParams struct {
	ProcessID        int               `json:"processId,omitempty"`
	RootURI          string            `json:"rootUri,omitempty"`
	WorkspaceFolders []WorkspaceFolder `json:"workspaceFolders,omitempty"`
}

type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerCapabilities advertises only implemented features. Provider fields are
// pointers/omitempty so unimplemented capabilities are simply absent.
type ServerCapabilities struct {
	TextDocumentSync   *TextDocumentSyncOptions `json:"textDocumentSync,omitempty"`
	CompletionProvider *CompletionOptions       `json:"completionProvider,omitempty"`
	DefinitionProvider bool                     `json:"definitionProvider,omitempty"`
	HoverProvider      bool                     `json:"hoverProvider,omitempty"`
	DocumentSymbol     bool                     `json:"documentSymbolProvider,omitempty"`
	DocumentFormatting bool                     `json:"documentFormattingProvider,omitempty"`
}

type TextDocumentSyncOptions struct {
	OpenClose bool         `json:"openClose"`
	Change    int          `json:"change"` // 0 none, 1 full, 2 incremental
	Save      *SaveOptions `json:"save,omitempty"`
}

type SaveOptions struct {
	IncludeText bool `json:"includeText"`
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

// TextDocumentSyncKind values.
const (
	SyncNone        = 0
	SyncFull        = 1
	SyncIncremental = 2
)

// --- document synchronization ---

type TextDocumentItem struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
	Text    string `json:"text"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// TextDocumentContentChangeEvent is either an incremental change (Range set) or
// a full replacement (Range nil).
type TextDocumentContentChangeEvent struct {
	Range *Range `json:"range,omitempty"`
	Text  string `json:"text"`
}

type DidSaveTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// --- diagnostics ---

type PublishDiagnosticsParams struct {
	URI string `json:"uri"`
	// Version is the document version the diagnostics were computed for. It is
	// optional per LSP; omitted (nil) rather than sent as a misleading 0 for
	// files that are not open, since version-checking clients discard a payload
	// whose version does not match the buffer they hold.
	Version     *int         `json:"version,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Diagnostic struct {
	Range              Range                          `json:"range"`
	Severity           int                            `json:"severity"` // 1 error, 2 warning
	Code               string                         `json:"code,omitempty"`
	Source             string                         `json:"source,omitempty"`
	Message            string                         `json:"message"`
	RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
}

type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// LSP DiagnosticSeverity values.
const (
	SeverityError   = 1
	SeverityWarning = 2
)

// --- definition / hover / formatting ---

// TextDocumentPositionParams is the request shape for definition and hover.
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// Hover is the result of textDocument/hover.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type MarkupContent struct {
	Kind  string `json:"kind"` // "markdown" | "plaintext"
	Value string `json:"value"`
}

type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// TextEdit replaces Range with NewText.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// --- completion ---

// CompletionParams is the textDocument/completion request (position-based; the
// optional completion context is ignored).
type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// CompletionItem is one suggestion returned to the client.
type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind,omitempty"`
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
}

// LSP CompletionItemKind values (subset).
const (
	CIKMethod   = 2
	CIKFunction = 3
	CIKVariable = 6
	CIKKeyword  = 14
	CIKProperty = 10
	CIKEnum     = 13
)

// --- document symbols ---

type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DocumentSymbol is a hierarchical outline node. Range covers the whole symbol;
// SelectionRange (⊆ Range) is what navigation selects.
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           int              `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// LSP SymbolKind values (subset).
const (
	SKModule    = 2
	SKNamespace = 3
	SKProperty  = 7
	SKFunction  = 12
	SKVariable  = 13
)

// --- watched files / dynamic registration ---

type DidChangeWatchedFilesParams struct {
	Changes []FileEvent `json:"changes"`
}

type FileEvent struct {
	URI  string `json:"uri"`
	Type int    `json:"type"` // 1 created, 2 changed, 3 deleted
}

// Registration payloads for client/registerCapability (server → client).
type RegistrationParams struct {
	Registrations []Registration `json:"registrations"`
}

type Registration struct {
	ID              string `json:"id"`
	Method          string `json:"method"`
	RegisterOptions any    `json:"registerOptions,omitempty"`
}

type DidChangeWatchedFilesRegistrationOptions struct {
	Watchers []FileSystemWatcher `json:"watchers"`
}

type FileSystemWatcher struct {
	GlobPattern string `json:"globPattern"`
}
