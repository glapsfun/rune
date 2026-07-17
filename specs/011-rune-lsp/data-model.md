# Phase 1 Data Model: Rune Language Server Protocol

Entities are the in-memory types the feature introduces or extends. All reference the existing `token.Span` (`{ File string; Start, End Position }` with byte `Offset`/`Line`/`Col`). Nothing here is persisted.

## Existing types extended

### `diag.Diagnostic` (extended — `internal/diag`)

| Field | Type | Notes |
|-------|------|-------|
| `Severity` | `diag.Severity` | existing (`Error`, `Warning`) |
| `Span` | `token.Span` | existing — carries the origin file |
| `Message` | `string` | existing |
| `Code` | `string` | **NEW** — stable RUNE#### code (FR-010); empty for legacy/uncoded emits |
| `Related` | `[]RelatedLocation` | **NEW** — related spans (FR-009), e.g. every node in a cycle |

### `diag.RelatedLocation` (new — `internal/diag`)

| Field | Type | Notes |
|-------|------|-------|
| `Span` | `token.Span` | location of the related node (may be in another file) |
| `Message` | `string` | e.g. `"build depends on test"` |

**Validation / rules**: `Code`, when set, MUST be one of the catalog constants (see `contracts/diagnostic-codes.md`). Cycle diagnostics (RUNE2003/RUNE3002) MUST populate `Related` with every participating task/file.

## Analysis layer (`internal/analysis`)

### `SourceStore` (interface)

```
Read(ctx, uri DocumentURI) ([]byte, error)
Exists(ctx, uri DocumentURI) bool
```

Implementations: `DiskSourceStore` (reads disk), `OverlaySourceStore{ disk SourceStore; documents map[DocumentURI]OpenDocument }` (overlay wins when open, else disk). Read-only by contract (no write method).

### `OpenDocument`

| Field | Type | Notes |
|-------|------|-------|
| `URI` | `DocumentURI` | editor document identity (file URI) |
| `Version` | `int` | LSP document version; monotonic per document |
| `Text` | `string` | current full buffer (after applying incremental edits) |

**Rules**: `Version` strictly increases per document; edits are applied to `Text` before analysis; the applied version is stamped onto the resulting `Snapshot`.

### `AnalyzeRequest`

| Field | Type | Notes |
|-------|------|-------|
| `URI` | `DocumentURI` | entry document to analyze |
| `Content` | `string` | optional inline content (overlay) |
| `Version` | `int` | version being analyzed |
| `Workspace` | `Workspace` | resolution scope + import graph |

### `Snapshot` (immutable result)

| Field | Type | Notes |
|-------|------|-------|
| `URI` | `DocumentURI` | analyzed entry document |
| `Version` | `int` | document version this snapshot reflects |
| `File` | `*ast.File` | composed AST (post-import splice) |
| `Sources` | `diag.SourceProvider` | resolves source bytes for rendering |
| `Diagnostics` | `diag.List` | all diagnostics (parser + semantic + project), across files |
| `Symbols` | `*language.Index` | symbol index over the composed file |
| `Imports` | `ImportGraph` | who-imports-what for invalidation |

**Rules**: A snapshot is never mutated after creation. Diagnostics are published only if `Snapshot.Version` still equals the document's current version (FR-016).

### `Workspace`

| Field | Type | Notes |
|-------|------|-------|
| `Root` | `DocumentURI` | project root (see detection order) |
| `EntryFile` | `DocumentURI` | primary Runefile |
| `Documents` | `map[DocumentURI]*Document` | tracked documents |
| `Imports` | `ImportGraph` | import relation |

**Root detection order** (FR-021): explicit client `workspaceFolder` → nearest dir with a `Runefile` → nearest dir with `.git` → current document's directory. Each workspace folder is an independent project in the first release.

### `ImportGraph`

| Field | Type | Notes |
|-------|------|-------|
| `ImportsByFile` | `map[DocumentURI][]DocumentURI` | forward edges |
| `ImportedByFile` | `map[DocumentURI][]DocumentURI` | reverse edges (transitive-importer lookup) |

**Rules**: On a file change, all transitive entries in `ImportedByFile` are invalidated and their roots re-analyzed (FR-022).

## Language layer (`internal/language`)

### `Symbol` + `SymbolKind`

`SymbolKind ∈ { Task, Variable, Parameter, Setting, Attribute, Builtin, Import, Module }`.

| Field | Type | Notes |
|-------|------|-------|
| `Name` | `string` | base name (e.g. `build`) |
| `QualifiedName` | `string` | namespaced name (e.g. `backend::build`) if applicable |
| `Kind` | `SymbolKind` | classification |
| `Definition` | `token.Span` | declaration span (carries origin file) |
| `Selection` | `token.Span` | precise name span for navigation |
| `Scope` | `ScopeID` | enclosing scope (file / task) |
| `Documentation` | `string` | doc comment or registry doc |
| `Signature` | `string` | e.g. `build target="debug"` |
| `Exported` | `bool` | false for `[private]` tasks |

### `Index`

| Field | Type | MVP? |
|-------|------|------|
| `ByName` | `map[string][]Symbol` | ✅ required |
| `ByQualified` | `map[string]Symbol` | ✅ required |
| `ByDocument` | `map[DocumentURI][]Symbol` | ✅ required |
| `References` | `map[SymbolID][]Location` | ⛔ deferred (rename/find-refs, out of scope) |
| `ScopeTree` | `*ScopeTree` | scope lookup for variables/params |

### `Builtin` (language registry entry)

| Field | Type | Notes |
|-------|------|-------|
| `Name` | `string` | e.g. `env`, `parallel`, `working-directory` |
| `Kind` | `BuiltinKind` | `Function` \| `Setting` \| `Attribute` \| `Executor` |
| `Signature` | `string` | e.g. `env(name, default?) -> string`, `[parallel]` |
| `Documentation` | `string` | hover/completion text |
| `IntroducedIn` | `string` | semver the symbol first appeared (feeds version checks) |

**Rules**: The registry is the single source consumed by hover, completion, and (incrementally) analyzer validation and CLI reference generation (FR-027). Validation diagnostics for invalid setting/attribute/executor (RUNE2007/2008/2009) MUST agree with what completion offers.

## Protocol layer (`internal/lsp`)

### `LineIndex`

| Field | Type | Notes |
|-------|------|-------|
| `Text` | `string` | document text for this version |
| `LineOffsets` | `[]int` | byte offset of each line start |

Methods: `ByteOffsetToPosition(offset) Position`, `PositionToByteOffset(pos) (int, error)`, `SpanToRange(span) Range`. Converts between byte offsets/columns and LSP line + UTF-16 character positions (R3). Rebuilt per document version.

### `InvalidStmt` (new AST node — `internal/ast`, if needed by R2)

| Field | Type | Notes |
|-------|------|-------|
| `Raw` | `string` | raw text of the unrecoverable region |
| `Sp` | `token.Span` | its span |

**Rules**: The analyzer skips rules that depend on an `InvalidStmt` subtree but continues checking valid siblings (FR-004).

## Relationships

```
Workspace 1─* Document
Workspace 1─1 ImportGraph
Service.Analyze(AnalyzeRequest) ─> Snapshot
Snapshot 1─1 *ast.File (composed) 1─* Symbol   (via language.Index)
Snapshot 1─* Diagnostic 1─* RelatedLocation
Diagnostic ─uses─> Code ∈ diagnostic-codes catalog
OverlaySourceStore ─wraps─> DiskSourceStore
config.Compose ─reads via─> SourceStore   (refactor: was os.ReadFile)
internal/lsp ─converts via─> LineIndex ─> LSP Position/Range
internal/cli & internal/lsp ─call─> internal/formatter.Format(*ast.File)
```

## State transitions (open document lifecycle)

```
didOpen(uri, v0, text)      → overlay[uri] = {v0, text};  analyze(v0) → publish
didChange(uri, v_n, edits)  → apply edits → overlay[uri].v = v_n; cancel in-flight;
                              debounce 100ms; analyze(v_n); publish IFF current == v_n
didSave(uri)                → analyze(current); publish
didClose(uri)               → drop overlay[uri]; subsequent reads fall back to disk
watched file changed        → invalidate transitive importers; re-analyze affected roots; publish
```
