# Phase 0 Research: Rune Language Server Protocol

All decisions below resolve the NEEDS-CLARIFICATION-equivalent unknowns surfaced while filling Technical Context. Format per decision: **Decision · Rationale · Alternatives considered**.

## R1. JSON-RPC / LSP protocol layer: hand-rolled vs third-party

**Decision**: Hand-write a minimal JSON-RPC 2.0 stdio transport (`Content-Length`-framed, per the LSP base protocol) plus a small, typed subset of LSP 3.17 payloads covering only the methods we implement (`initialize`, `initialized`, `shutdown`, `exit`, `textDocument/didOpen|didChange|didSave|didClose`, `textDocument/completion|definition|hover|documentSymbol|formatting`, `textDocument/publishDiagnostics`, `$/cancelRequest`, `workspace/didChangeWatchedFiles`). No new third-party dependency.

**Rationale**: Principle V (Boringly Portable) and the existing precedent — the project chose a dependency-free `internal/semver` over a third-party import specifically to protect portability. The implemented method set is small (~12 payloads), so the typed surface is modest and fully under our control. Encoding uses the standard library (`encoding/json`, `bufio`). Zero new supply-chain surface (Principle VII) and nothing to vendor for `CGO_ENABLED=0` static builds.

**Alternatives considered**:
- `go.lsp.dev/protocol` + `go.lsp.dev/jsonrpc2` — complete, well-typed, but pulls a large transitive type surface and a new dependency for a feature that uses a fraction of it. Rejected to preserve the dependency-free posture; revisit only if the method set grows substantially.
- `gopls`' internal `jsonrpc2` — not importable as a stable public package.
- Reusing the MCP SDK's JSON-RPC plumbing (`github.com/modelcontextprotocol/go-sdk`, already a dep) — its framing/types are MCP-specific, not LSP base protocol; coupling the two protocols was rejected as leaky.

## R2. Parser error recovery (FR-004, FR-005)

**Decision**: Extend the *existing* hand-written parser with a recovery posture rather than adding a `ParseMode` switch up front. The parser already recovers at top level (`parseItem` → `recoverToNewline` → continue) and returns `nil` for failed items. Harden this so that: (a) partial/incomplete constructs (unterminated string, open bracket, dangling `:`) synchronize at the documented boundaries (newline, next declaration keyword, attribute closing bracket, dedent to top level, EOF) without dropping subsequent valid declarations; (b) where a declaration is unrecoverable, emit an `ast.InvalidStmt{Raw, Span}` node so the analyzer can skip it while still processing siblings. Introduce an explicit `ParseRecover` flag only if strict execution parsing must reject something recovery would accept (evaluate during implementation).

**Rationale**: The constitution mandates one hand-written parser (Principle IV) and the spec forbids a second grammar. The current parser is already substantially recovering; the gap is robustness on incomplete *expressions* and *task headers*, not a rewrite. Keeping one code path (recovery is a superset that execution can still gate on `HasErrors()`) avoids grammar drift. The FR-005 invariant (terminate, never panic, in-bounds ranges) is a fuzz target regardless of mode.

**Alternatives considered**:
- Tree-sitter grammar — explicitly excluded by the spec; would create a second grammar (violates Principle IV). Rejected.
- A separate recovery parser — duplicates the grammar; same objection.
- Two hard-separated modes (`ParseStrict`/`ParseRecover`) from day one — deferred; risks divergent behavior. A single hardened path with an optional gate is simpler and testable via the existing `HasErrors()` seam that execution already uses.

**Implementation outcome (2026-07-13)**: The existing parser already recovers via **drop-and-continue** — a failed declaration emits a diagnostic, is dropped, and the loop resynchronizes at the next line, so valid siblings still parse and analyze. `FuzzParseRecover` (400k+ execs) confirms the FR-005 invariant (terminates, never panics, all diagnostic ranges in-bounds), and `TestRecoveryKeepsValidDeclarations` confirms valid declarations survive around unterminated strings, garbage lines, and incomplete assignments. Given this, **`InvalidStmt` was intentionally NOT added**: drop-and-continue leaves no broken subtree in the AST, so there is nothing for the analyzer to "skip" (T014 is satisfied by the analyzer's defensive iteration over surviving nodes, covered by `analyzer.TestAnalyzeRecoveredFile`). Adding an unused node would violate the minimalism principle. **Known limitation**: because the lexer transparently continues logical lines inside unclosed `()`/`[]`/`{}` (group continuation), an unterminated group consumes to EOF, so declarations after it are not recoverable (documented in `parser.TestUnterminatedGroupIsBounded`); this is transient while editing and out of scope to change without reworking the lexer's continuation model.

## R3. Position conversion: byte columns ↔ UTF-16 (FR-006)

**Decision**: A dedicated `lsp/convert.go` `LineIndex` built per document version. `token.Position` carries a 0-based byte `Offset` and 1-based byte `Line`/`Col`; LSP positions are 0-based line + 0-based UTF-16 code-unit character. Conversion walks the line's bytes, decoding runes and counting UTF-16 code units (2 for runes ≥ U+10000, else 1). `LineIndex` precomputes line-start byte offsets for O(log n) offset↔line lookup; per-line UTF-16 counting is done on demand.

**Rationale**: `token.Span` already carries `Offset` + byte `Line`/`Col`, so byte-accurate mapping to UTF-16 is deterministic. Centralizing it (never per-handler) is a spec requirement and the only sane way to keep the unicode matrix (ASCII, Ukrainian, emoji, combining marks, CRLF/LF, empty lines, EOF) correct and fuzz-tested (SC-008). CRLF: the byte offset already points past `\r`; the index treats `\r\n` as the line terminator so a position at end-of-line maps consistently.

**Alternatives considered**:
- UTF-8 or codepoint columns — non-conformant with LSP's default `positionEncoding` of UTF-16; rejected. (We advertise no `positionEncoding` capability, so UTF-16 is required.)
- Recomputing offsets ad hoc in each handler — the exact anti-pattern the spec calls out. Rejected.

## R4. Editor overlay threading through import composition (FR-003)

**Decision**: Introduce `analysis.SourceStore { Read(ctx, uri) ([]byte, error); Exists(ctx, uri) bool }` with `DiskSourceStore` and `OverlaySourceStore` (overlay wins when a document is open, else disk). Refactor `config.Compose` (and `spliceImports`/`loadMods`) to read imported/mod files **through an injected reader** instead of calling `os.ReadFile` directly. Today `Compose` accepts a `diag.SourceProvider` but ignores it for reads — the refactor routes all import/mod file reads through the store so overlays apply transitively.

**Rationale**: FR-003 requires unsaved editor content to be used for imports too. The existing `os.ReadFile` calls inside `spliceImports`/`loadMods` are the exact points that bypass the overlay. Making the reader injectable is a small, backward-compatible change: the CLI passes a disk-backed store (current behavior preserved), the LSP passes an overlay store. `diag.SourceProvider` (a `func(path)([]byte,bool)`) can be adapted to/from `SourceStore` to keep rendering unchanged.

**Alternatives considered**:
- Writing overlays to temp files so disk reads "just work" — violates the no-file-write safety rule (FR-028) and is racy. Rejected.
- A global mutable source hook — hidden global state, hostile to concurrency and testing (Principle VIII). Rejected in favor of explicit injection.

## R5. Diagnostic model extension: codes + related locations (FR-007/009/010)

**Decision**: Extend `diag.Diagnostic` with `Code string` and `Related []RelatedLocation` (where `RelatedLocation{ Span token.Span; Message string }`), both optional/zero-valued by default. Introduce a `diag` constant catalog for the RUNE#### codes. Existing constructors (`New`, `Warn`, `Errorf`) keep working (empty code); add code-carrying constructors/helpers. Each emit site in `parser`/`analyzer`/`config` is updated to attach its stable code; cycle diagnostics attach every involved task/file as `Related`.

**Rationale**: FR-010 (clarified) makes the RUNE#### codes a **public contract** asserted by golden tests, printed by `rune analyze`, and sent to editors. A `Code` field is the minimal model change; `Related` is required for cycle diagnostics (FR-009) and maps directly to LSP `DiagnosticRelatedInformation`. Additive fields keep all current call sites and golden output stable except where a code is now emitted (golden files regenerated deliberately, per Principle VI).

**Alternatives considered**:
- Encoding the code inside the message string — unstructured, unfilterable, fails the "published to editors" contract. Rejected.
- A parallel side-table mapping message→code — brittle and duplicative. Rejected.

## R6. Symbol index & cross-file attribution (FR-019, FR-026)

**Decision**: Build the `language.Index` from the **composed** `*ast.File` (post-`Compose`), keyed by `Symbol` with `Definition token.Span`. Because every AST node retains its originating `Sp.File` through composition, cross-file go-to-definition and per-file attribution (FR-009a) fall out for free — a task spliced from an import still points at its own file/line. Namespaced module tasks (`name::task`) are indexed by both qualified and base name. Private (`[private]`) tasks are tagged so completion can apply the same-file-only rule (FR-019a) while definition/hover ignore the tag.

**Rationale**: `Compose` already merges all tasks/vars into one `*ast.File` while preserving each node's `Sp` (which includes `File`). This is the single most load-bearing existing behavior for the LSP: the analyzer and index operate on one composed tree, yet spans still resolve to the right source file. The index needs only `ByName`/`ByQualified`/`ByDocument` for the MVP (references index deferred).

**Alternatives considered**:
- Re-parsing each file independently for the index — loses the composed view (namespacing, collisions) that the analyzer already computes. Rejected.
- Building the index inside the analyzer — couples analysis to LSP concerns; the index is a separate `language` concern consuming the analyzer's output. Rejected.

## R7. Debounce, cancellation, and version guarding (FR-016)

**Decision**: Per open document, a single-flight analysis worker. On `didChange`: apply the edit to the overlay buffer (bumping version), cancel the in-flight analysis via its `context.Context`, start a ~100 ms debounce timer, and on fire launch analysis for the latest version. Publish diagnostics only if the analyzed version still equals the current document version; otherwise discard. Requests (`completion`/`definition`/`hover`) carry the client request id and honor `$/cancelRequest`.

**Rationale**: FR-016/SC edge case "stale results" require that older-version diagnostics never overwrite newer ones. A version stamp compared at publish time is the definitive guard; context cancellation stops wasted work early (Principle VIII: every goroutine has a clear owner and context-driven exit). ~100 ms matches the spec's recommendation and is tunable.

**Alternatives considered**:
- Analyze synchronously on every keystroke — wastes CPU and risks publishing intermediate states; rejected.
- Incremental AST reparse — explicitly excluded until profiling justifies it (full reparse is fine for small Runefiles). Rejected for the MVP.

## R8. Formatter extraction (FR-020)

**Decision**: Move `formatFile` and its helpers from `internal/cli/fmt.go` into a new `internal/formatter` package exposing `Format(*ast.File) string`. `internal/cli` (the `--fmt` path and `fmtRewrite`) becomes its first caller; `internal/lsp/formatting.go` is the second. Golden formatter tests move with it.

**Rationale**: FR-020 forbids shelling out and requires a direct call. `internal/cli` transitively imports the runtime/execution stack, which the read-only LSP must not pull in, so the formatter must live below `cli`. The move is behavior-preserving (byte-for-byte identical output; existing `fmt_test.go` guards it).

**Alternatives considered**:
- `lsp` importing `cli` — drags execution code into the analysis surface; rejected.
- Duplicating the formatter — two formatters violate "one formatter" and drift. Rejected.

## R9. Single built-in / language registry (FR-027)

**Decision**: A `language/builtin.go` registry of `Builtin{ Name, Kind, Signature, Documentation, IntroducedIn }` covering built-in functions, settings, and attributes, as the single source consumed by hover, completion, and (incrementally) CLI help/reference and validation. Where the analyzer currently hard-codes valid setting/attribute/executor names, it is refactored to consult the registry so the "invalid setting/attribute/executor" diagnostics (RUNE2007/2008/2009) and completion agree by construction.

**Rationale**: FR-027 forbids duplicate metadata lists. A single registry keeps validation, completion, and hover consistent and is the natural home for `IntroducedIn` (feeding the incompatible-version story and future minimum-version checks). Building it now avoids a completion/validation skew bug class.

**Alternatives considered**:
- Separate lists per consumer — the exact duplication the spec prohibits; guarantees drift. Rejected.
- Deriving metadata from code comments/reflection — fragile and untyped. Rejected.

## R10. `rune analyze` transitive scope & exit codes (FR-009a, FR-023, FR-025)

**Decision**: `rune analyze [path]` runs the analysis service on the target (default `Runefile`) **with its transitive imports/mods**, collecting diagnostics from all files attributed to their own `file:line:col`. Human output: one line per diagnostic `file:line:col: severity[CODE]: message` + a summary count; `--json` emits structured diagnostics (code, message, severity, range, related). Exit 0 (no error-severity diagnostics), 3 (≥1 error), 1 (internal failure). The undocumented-public-task warning (FR-008a) is warning-severity and never triggers exit 3 on its own.

**Rationale**: Clarified in the spec (2026-07-10). Reuses the same service and snapshot as the LSP so results are identical (FR-002/SC-002). Exit-code semantics mirror the existing execution path's "static errors → exit 3, run nothing."

**Alternatives considered**:
- Single-file-only analysis — contradicts the clarification and the import-graph design. Rejected.
- Warnings affecting exit code — would make the doc-warning a CI gate by default, surprising users. Rejected.

## Cross-cutting: safety verification (FR-028, SC-005)

**Decision**: Enforce "no side effects during analysis" structurally — the `analysis`, `language`, and `lsp` packages do not import `internal/runtime/*`, `os/exec`, or network packages, and the `SourceStore` is read-only. Add a test that fails if any analysis/LSP/analyze operation runs a task, spawns a process, opens a socket, or writes a project file (e.g. via a guarded fake environment and an import-graph assertion / `go list` deps check in CI).

**Rationale**: The safety model is the feature's central promise (Principle I & VII). A structural (dependency-direction) guarantee plus a behavioral test is stronger than code review alone.
