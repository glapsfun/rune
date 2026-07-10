# Contract: LSP Capability Semantics

Detailed request/response semantics for each advertised capability. Positions are 0-based line + 0-based UTF-16 character (default `positionEncoding`). All ranges are produced via the single `LineIndex` converter (R3).

## textDocument/completion

**Trigger**: manual, or trigger characters `[ ( . { :`. **Context detection** determines the completion set:

| Cursor context | Returns | Notes |
|----------------|---------|-------|
| dependency position (after task header `name:` / `&&`) | task names incl. namespaced module tasks (`backend::build`) | items show parameters, documentation, source module; private tasks only within their own file (FR-019a) |
| inside `{{ … }}` interpolation | in-scope variables + parameters | task parameters rank above global variables |
| after `set ` | setting names | from registry |
| inside `[ … ]` | attribute names | `confirm`, `private`, `parallel`, `group`, `cache`, `env`, `working-directory`, `network`, `no-cd`, `no-exit-message`, platform selectors |
| executor position `name (…)` | executors | `sh`, `python`, `node`, `agent` |
| builtin prefix in expression | builtin functions | e.g. `os_family()` |

Each `CompletionItem` carries `label`, `detail` (signature), `documentation`, and an appropriate `kind`.

## textDocument/definition

Returns the `Location` (or `Location[]`) of the declaration for the symbol under the cursor:

| Symbol under cursor | Resolves to |
|---------------------|-------------|
| dependency | task declaration (`Selection` span) |
| variable reference | assignment |
| parameter interpolation | parameter declaration in the task header |
| flat-imported task (bare name, `import "…"`) | its declaration in the imported file (resolved via `Sp.File`) |
| module namespace (`mod name`) | the module file (its start) / the `mod` declaration |
| module task (`name::task`) | declaration inside the module file |

Cross-file results use overlay content when the target file is open (FR-019). Settings/attributes resolve to their documentation (may be surfaced via hover rather than a file location).

## textDocument/hover

Returns markdown content assembled from the symbol + shared registry:

- **Task**: signature (`build target="debug"`), doc comment, executor, group, `Defined in: FILE:LINE`.
- **Parameter**: `name: type`, `Default: "…"`, declaring task.
- **Attribute**: description from registry.
- **Builtin**: signature (`env(name, default?) → string`) + documentation.

## textDocument/documentSymbol

Returns a hierarchy (or flat list) grouping the file's symbols into settings, variables, imports, tasks, and modules, each with a `range` and `selectionRange` that navigate to the declaration (FR-017, User Story 6).

## textDocument/formatting

Returns a single full-document `TextEdit` whose `newText` is `internal/formatter.Format(file)` over the current (possibly unsaved) buffer. Idempotent (formatting formatted content is a no-op edit). No child process; no file write (FR-020). If the buffer has parse errors severe enough that a faithful format is impossible, the server returns no edits rather than corrupting content.

## textDocument/publishDiagnostics

Server→client notification carrying `{ uri, version, diagnostics[] }`. Each diagnostic maps `diag.Diagnostic` → LSP:

| `diag` field | LSP field |
|--------------|-----------|
| `Severity` | `severity` (Error→1, Warning→2) |
| `Span` | `range` (via LineIndex) |
| `Message` | `message` |
| `Code` | `code` |
| `Related` | `relatedInformation[]` (`{location, message}`) |

`version` matches the analyzed document version; the client uses it to drop stale publishes (mirrors the server's own guard, FR-016).

## $/cancelRequest

Cancels the in-flight request with the given id via context cancellation; the server responds to the cancelled request with a JSON-RPC cancellation error or a best-effort empty result.
