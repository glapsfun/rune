package lsp

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func jsonUnmarshal(data json.RawMessage, v any) error { return json.Unmarshal(data, v) }
func containsStr(s, sub string) bool                  { return strings.Contains(s, sub) }

// testClient drives a Server over real JSON-RPC framing via in-memory pipes.
type testClient struct {
	t      *testing.T
	conn   *Conn
	nextID int
}

func startServer(t *testing.T) (*testClient, chan error) {
	t.Helper()
	cr, cw := io.Pipe() // client -> server
	sr, sw := io.Pipe() // server -> client
	srv := NewServer(cr, sw, Options{Version: "9.9.9", Debounce: 5 * time.Millisecond})
	done := make(chan error, 1)
	go func() { done <- srv.Run() }()
	return &testClient{t: t, conn: NewConn(sr, cw), nextID: 1}, done
}

func (c *testClient) request(method string, params any) *Message {
	c.t.Helper()
	id := c.nextID
	c.nextID++
	raw, _ := json.Marshal(id)
	rm := json.RawMessage(raw)
	pr, _ := json.Marshal(params)
	if err := c.conn.Write(&Message{JSONRPC: "2.0", ID: &rm, Method: method, Params: pr}); err != nil {
		c.t.Fatalf("write %s: %v", method, err)
	}
	// Skip any server-initiated request (e.g. client/registerCapability) and
	// return the response to our request (a message with an id and no method).
	for {
		m := c.read()
		if m.Method == "" {
			return m
		}
	}
}

func (c *testClient) notify(method string, params any) {
	c.t.Helper()
	pr, _ := json.Marshal(params)
	if err := c.conn.Write(&Message{JSONRPC: "2.0", Method: method, Params: pr}); err != nil {
		c.t.Fatalf("notify %s: %v", method, err)
	}
}

// read reads one message with a timeout so a server bug fails fast rather than
// hanging the suite.
func (c *testClient) read() *Message {
	c.t.Helper()
	type res struct {
		m   *Message
		err error
	}
	ch := make(chan res, 1)
	go func() {
		m, err := c.conn.Read()
		ch <- res{m, err}
	}()
	select {
	case r := <-ch:
		if r.err != nil {
			c.t.Fatalf("read: %v", r.err)
		}
		return r.m
	case <-time.After(3 * time.Second):
		c.t.Fatal("timed out waiting for a server message")
		return nil
	}
}

func (c *testClient) readPublish() PublishDiagnosticsParams {
	c.t.Helper()
	// Skip unrelated server messages (e.g. client/registerCapability) until the
	// next publishDiagnostics arrives.
	for {
		m := c.read()
		if m.Method != "textDocument/publishDiagnostics" {
			continue
		}
		var p PublishDiagnosticsParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			c.t.Fatalf("decode publishDiagnostics: %v", err)
		}
		return p
	}
}

func hasDiagCode(diags []Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

// TestProtocolLifecycle exercises initialize -> initialized -> didOpen ->
// publishDiagnostics -> didChange(clear) -> shutdown -> exit (spec SC-009).
func TestProtocolLifecycle(t *testing.T) {
	client, done := startServer(t)
	const uri = "file:///tmp/proj/Runefile"

	// initialize
	resp := client.request("initialize", InitializeParams{})
	if resp.ID == nil || string(*resp.ID) != "1" {
		t.Fatalf("initialize response id = %v, want 1", resp.ID)
	}
	var initResult InitializeResult
	if err := json.Unmarshal(resp.Result, &initResult); err != nil {
		t.Fatalf("decode initialize result: %v", err)
	}
	if initResult.ServerInfo.Name != "rune" || initResult.ServerInfo.Version != "9.9.9" {
		t.Errorf("serverInfo = %+v, want rune/9.9.9", initResult.ServerInfo)
	}
	if initResult.Capabilities.TextDocumentSync == nil || initResult.Capabilities.TextDocumentSync.Change != SyncIncremental {
		t.Errorf("capabilities = %+v, want incremental textDocumentSync", initResult.Capabilities)
	}
	client.notify("initialized", struct{}{})

	// didOpen with an unknown dependency -> expect RUNE2001.
	broken := "# Build.\nbuild:\n    @echo build\n# Deploy.\ndeploy: missing\n    @echo deploy\n"
	client.notify("textDocument/didOpen", DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: uri, Version: 1, Text: broken},
	})
	pub := client.readPublish()
	if !hasDiagCode(pub.Diagnostics, "RUNE2001") {
		t.Errorf("expected RUNE2001 on open, got %+v", pub.Diagnostics)
	}

	// didChange fixing the dependency (full replacement) -> error clears.
	fixed := "# Build.\nbuild:\n    @echo build\n# Deploy.\ndeploy: build\n    @echo deploy\n"
	client.notify("textDocument/didChange", DidChangeTextDocumentParams{
		TextDocument:   VersionedTextDocumentIdentifier{URI: uri, Version: 2},
		ContentChanges: []TextDocumentContentChangeEvent{{Text: fixed}},
	})
	pub2 := client.readPublish()
	if hasDiagCode(pub2.Diagnostics, "RUNE2001") {
		t.Errorf("RUNE2001 should have cleared, got %+v", pub2.Diagnostics)
	}
	if pub2.Version == nil || *pub2.Version != 2 {
		t.Errorf("publish version = %v, want 2", pub2.Version)
	}

	// shutdown + exit.
	sr := client.request("shutdown", nil)
	if sr.Error != nil {
		t.Errorf("shutdown error: %+v", sr.Error)
	}
	client.notify("exit", nil)

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("server exited with error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server did not exit after exit notification")
	}
}

// TestIncrementalEditProducesError drives an incremental (ranged) change that
// introduces an error, exercising the LineIndex-based edit application.
func TestIncrementalEditProducesError(t *testing.T) {
	client, done := startServer(t)
	defer func() {
		client.notify("exit", nil)
		<-done
	}()
	const uri = "file:///tmp/proj/Runefile"

	client.request("initialize", InitializeParams{})
	client.notify("initialized", struct{}{})

	// Open a clean file (build documented, deploy depends on build).
	clean := "# Build.\nbuild:\n    @echo build\n# Deploy.\ndeploy: build\n    @echo deploy\n"
	client.notify("textDocument/didOpen", DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: uri, Version: 1, Text: clean},
	})
	pub := client.readPublish()
	if hasDiagCode(pub.Diagnostics, "RUNE2001") {
		t.Fatalf("clean file should have no unknown-dependency error: %+v", pub.Diagnostics)
	}

	// Incrementally replace "build" on the deploy line (line index 4) with "gone".
	// Line 4 is "deploy: build"; "build" starts at character 8.
	client.notify("textDocument/didChange", DidChangeTextDocumentParams{
		TextDocument: VersionedTextDocumentIdentifier{URI: uri, Version: 2},
		ContentChanges: []TextDocumentContentChangeEvent{{
			Range: &Range{Start: Position{Line: 4, Character: 8}, End: Position{Line: 4, Character: 13}},
			Text:  "gone",
		}},
	})
	pub2 := client.readPublish()
	if !hasDiagCode(pub2.Diagnostics, "RUNE2001") {
		t.Errorf("expected RUNE2001 after breaking edit, got %+v", pub2.Diagnostics)
	}
}

// TestDefinitionHoverFormatting drives the three request handlers over the
// protocol against an open document.
func TestDefinitionHoverFormatting(t *testing.T) {
	client, done := startServer(t)
	defer func() { client.notify("exit", nil); <-done }()
	const uri = "file:///tmp/proj/Runefile"

	initRes := client.request("initialize", InitializeParams{})
	var ir InitializeResult
	_ = jsonUnmarshal(initRes.Result, &ir)
	if !ir.Capabilities.DefinitionProvider || !ir.Capabilities.HoverProvider || !ir.Capabilities.DocumentFormatting {
		t.Fatalf("capabilities missing definition/hover/formatting: %+v", ir.Capabilities)
	}
	client.notify("initialized", struct{}{})

	doc := "# Build the app.\nbuild target=\"debug\":\n    @echo {{target}}\n# Deploy.\ndeploy: build\n    @echo deploy\n"
	client.notify("textDocument/didOpen", DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: uri, Version: 1, Text: doc},
	})
	client.readPublish() // consume diagnostics

	// definition on "build" in "deploy: build" (line 4, char 8).
	defRes := client.request("textDocument/definition", TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 4, Character: 8},
	})
	var locs []Location
	if err := jsonUnmarshal(defRes.Result, &locs); err != nil || len(locs) != 1 {
		t.Fatalf("definition result = %s (err %v)", string(defRes.Result), err)
	}
	if locs[0].Range.Start.Line != 1 { // build task header is line index 1
		t.Errorf("definition points at line %d, want 1", locs[0].Range.Start.Line)
	}

	// hover on the "build" task name (line 1, char 0).
	hovRes := client.request("textDocument/hover", TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 1, Character: 0},
	})
	var hov Hover
	if err := jsonUnmarshal(hovRes.Result, &hov); err != nil {
		t.Fatalf("hover decode: %v", err)
	}
	if hov.Contents.Kind != "markdown" || !containsStr(hov.Contents.Value, `build target="debug"`) {
		t.Errorf("hover = %+v", hov.Contents)
	}
}

func TestFormattingReturnsCanonicalEdit(t *testing.T) {
	client, done := startServer(t)
	defer func() { client.notify("exit", nil); <-done }()
	const uri = "file:///tmp/proj/Runefile"
	client.request("initialize", InitializeParams{})
	client.notify("initialized", struct{}{})

	// 2-space body indent is valid but not canonical (canonical is 4 spaces).
	client.notify("textDocument/didOpen", DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: uri, Version: 1, Text: "# B.\nbuild:\n  @echo hi\n"},
	})
	client.readPublish()

	res := client.request("textDocument/formatting", DocumentFormattingParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
	})
	var edits []TextEdit
	if err := jsonUnmarshal(res.Result, &edits); err != nil {
		t.Fatalf("formatting decode: %v", err)
	}
	if len(edits) != 1 || !containsStr(edits[0].NewText, "    @echo hi") {
		t.Errorf("formatting edits = %+v, want one canonical edit with 4-space indent", edits)
	}
}

// TestCompletionOverProtocol drives textDocument/completion for a dependency.
func TestCompletionOverProtocol(t *testing.T) {
	client, done := startServer(t)
	defer func() { client.notify("exit", nil); <-done }()
	const uri = "file:///tmp/proj/Runefile"

	initRes := client.request("initialize", InitializeParams{})
	var ir InitializeResult
	_ = jsonUnmarshal(initRes.Result, &ir)
	if ir.Capabilities.CompletionProvider == nil {
		t.Fatal("completion capability not advertised")
	}
	client.notify("initialized", struct{}{})

	// "deploy: bu" — cursor after "bu" on line 4.
	doc := "# B.\nbuild:\n    @echo b\n# D.\ndeploy: bu\n    @echo d\n"
	client.notify("textDocument/didOpen", DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: uri, Version: 1, Text: doc},
	})
	client.readPublish()

	res := client.request("textDocument/completion", CompletionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 4, Character: 10}, // end of "deploy: bu"
	})
	var items []CompletionItem
	if err := jsonUnmarshal(res.Result, &items); err != nil {
		t.Fatalf("completion decode: %v (%s)", err, string(res.Result))
	}
	found := false
	for _, it := range items {
		if it.Label == "build" {
			found = true
			if it.Kind != CIKMethod {
				t.Errorf("build kind = %d, want %d", it.Kind, CIKMethod)
			}
		}
	}
	if !found {
		t.Errorf("completion should suggest 'build', got %+v", items)
	}
}

// TestDocumentSymbolOverProtocol verifies the outline is produced with grouped
// categories and navigable ranges.
func TestDocumentSymbolOverProtocol(t *testing.T) {
	client, done := startServer(t)
	defer func() { client.notify("exit", nil); <-done }()
	const uri = "file:///tmp/proj/Runefile"

	initRes := client.request("initialize", InitializeParams{})
	var ir InitializeResult
	_ = jsonUnmarshal(initRes.Result, &ir)
	if !ir.Capabilities.DocumentSymbol {
		t.Fatal("documentSymbol capability not advertised")
	}
	client.notify("initialized", struct{}{})

	doc := "output := \"dist\"\n# Build.\nbuild:\n    @echo b\n# Test.\ntest: build\n    @echo t\n"
	client.notify("textDocument/didOpen", DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: uri, Version: 1, Text: doc},
	})
	client.readPublish()

	res := client.request("textDocument/documentSymbol", DocumentSymbolParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
	})
	var syms []DocumentSymbol
	if err := jsonUnmarshal(res.Result, &syms); err != nil {
		t.Fatalf("documentSymbol decode: %v (%s)", err, string(res.Result))
	}
	groups := map[string]DocumentSymbol{}
	for _, g := range syms {
		groups[g.Name] = g
	}
	if _, ok := groups["variables"]; !ok {
		t.Errorf("outline missing variables group: %+v", syms)
	}
	tasks, ok := groups["tasks"]
	if !ok || len(tasks.Children) != 2 {
		t.Fatalf("tasks group = %+v, want 2 children", tasks)
	}
	// SelectionRange must be inside Range (LSP requirement).
	for _, c := range tasks.Children {
		if positionLess(c.Range.End, c.SelectionRange.Start) {
			t.Errorf("child %q selectionRange not within range", c.Name)
		}
		if c.Kind != SKFunction {
			t.Errorf("task %q kind = %d, want %d", c.Name, c.Kind, SKFunction)
		}
	}
}

// TestWatchedFileImportPropagation verifies that an on-disk change to an
// imported file refreshes diagnostics for an open document that imports it
// (spec FR-022), driven via workspace/didChangeWatchedFiles.
func TestWatchedFileImportPropagation(t *testing.T) {
	dir := t.TempDir()
	rootPath := filepath.Join(dir, "Runefile")
	importPath := filepath.Join(dir, "common.rune")
	if err := os.WriteFile(importPath, []byte("# S.\nshared:\n    echo s\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rootSrc := "import \"common.rune\"\n# D.\ndeploy: shared\n    echo d\n"
	if err := os.WriteFile(rootPath, []byte(rootSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	client, done := startServer(t)
	defer func() { client.notify("exit", nil); <-done }()
	client.request("initialize", InitializeParams{})
	client.notify("initialized", struct{}{})

	rootURI := pathToURI(rootPath)
	client.notify("textDocument/didOpen", DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: rootURI, Version: 1, Text: rootSrc},
	})
	pub := client.readPublish()
	if hasDiagCode(pub.Diagnostics, "RUNE2001") {
		t.Fatalf("with a valid import, deploy: shared should resolve: %+v", pub.Diagnostics)
	}

	// Change the imported file on disk so it no longer defines `shared`.
	if err := os.WriteFile(importPath, []byte("# O.\nother:\n    echo o\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	client.notify("workspace/didChangeWatchedFiles", DidChangeWatchedFilesParams{
		Changes: []FileEvent{{URI: pathToURI(importPath), Type: 2}},
	})

	pub2 := client.readPublish()
	if !hasDiagCode(pub2.Diagnostics, "RUNE2001") {
		t.Errorf("expected RUNE2001 after the imported task disappeared, got %+v", pub2.Diagnostics)
	}
}
