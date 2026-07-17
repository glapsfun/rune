# Implementation Plan: Rune Language Server Protocol

**Branch**: `011-rune-lsp` | **Date**: 2026-07-10 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/011-rune-lsp/spec.md`

## Summary

Add first-class IDE support for Runefiles through a new `rune lsp` subcommand (a JSON-RPC/LSP 3.17 server over stdio) and a standalone `rune analyze` command, both driven by a single new **analysis service** that wraps Rune's existing parse → compose → analyze pipeline behind a source-overlay abstraction. The server delivers real-time diagnostics, context-aware completion, go-to-definition (including cross-file), hover, document symbols, and canonical formatting — reusing the existing parser, analyzer, import resolver, diagnostics, and formatter with **no second grammar**. No task, shell, agent, or network side effect ever runs during analysis.

The technical spine is four additive `internal/` packages (`analysis`, `language`, `lsp`, `formatter`) plus targeted, backward-compatible extensions to two existing packages: `diag.Diagnostic` gains a stable `Code` and `Related []RelatedLocation`; `config.Compose` is refactored to read imports through an injected source store instead of `os.ReadFile` (so editor overlays apply transitively). The formatter is extracted from `internal/cli` into `internal/formatter` so both the CLI and the LSP call it directly.

## Technical Context

**Language/Version**: Go 1.25.0 (module `github.com/rune-task-runner/rune`)

**Primary Dependencies**: Existing engine packages only for language logic (`lexer`, `parser`, `analyzer`, `config`, `eval`, `ast`, `diag`, `token`). **Target: zero new third-party dependencies** — a minimal hand-written JSON-RPC 2.0 stdio framing (`Content-Length` headers) plus a hand-written typed subset of LSP 3.17 (only the ~10 request/notification payloads we implement). This honors Principle V (Boringly Portable) and mirrors the `internal/semver` precedent (dependency-free over a third-party import). Full evaluation in `research.md`.

**Storage**: N/A. All state is in-memory: open-document overlay map, per-workspace analysis snapshots, and the import graph. Nothing is persisted; nothing is written to the project.

**Testing**: `go test ./...` inside the Docker Compose harness (never on host). Layers: unit (position conversion, scope, index, completion, definition, hover, diagnostic mapping, overlay edit application, snapshot invalidation); golden (diagnostics for each RUNE code; `rune analyze` human + `--json` output); binary-level protocol integration (drive a real `rune lsp` subprocess through the JSON-RPC lifecycle); fuzz (parser recovery, UTF-16 position conversion, incremental edit application, malformed JSON-RPC).

**Target Platform**: Single static binary, `CGO_ENABLED=0`, on Linux, macOS, and Windows. Transport is stdio only.

**Project Type**: CLI + language engine front-end. The LSP is another interface over the existing engine, alongside the CLI and MCP server — not a separate binary.

**Performance Goals** (SC-010, product targets, not release blockers): open-document analysis P50 < 50 ms / P95 < 150 ms; completion/definition/hover P95 < 50 ms; 100-file workspace initial index < 1 s; single imported-file update < 250 ms. Benchmarks (`BenchmarkParseRunefile`, `BenchmarkAnalyzeRunefile`, `BenchmarkBuildSymbolIndex`, `BenchmarkCompletion`, `BenchmarkImportedFileInvalidation`) added before any optimization (Principle VIII: no optimization without a profile).

**Constraints**: stdout carries protocol bytes only (logs → stderr or `--log-file`); the analysis pipeline must terminate and must not panic on any input, with all diagnostic ranges in-bounds (fuzz-enforced); zero execution/network/file-write side effects during analysis; diagnostics for a superseded document version must never be published after a newer version arrives (context cancellation + version guard).

**Scale/Scope**: Individual Runefiles are small (tens to low-hundreds of lines). Workspaces up to ~100 Runefiles. Six LSP capabilities in the first release; exclusions per spec (rename, references, semantic tokens, code actions, editor task execution, inlay hints, workspace symbol search, TCP, incremental AST, Tree-sitter).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|-----------|------------|
| **I. Command Runner, Not a Build System** | ✅ The LSP and `rune analyze` execute nothing. FR-028/029 forbid task/shell/agent/network/file-write side effects; SC-005 enforces it with a side-effect-detecting test. Reinforces, does not bend, the principle. |
| **II. Errors Are a Feature** | ✅ Reuses the existing `file:line:col` + caret span model and extends it: stable diagnostic codes (FR-010) and related locations for cycles (FR-009). Diagnostics now reach the editor live. |
| **III. Minimal, Total DSL** | ✅ No grammar or expression-language change. The DSL surface is untouched. |
| **IV. Hand-Written Front End, Idiomatic Go** | ⚠️ Justified. Parser recovery stays in the existing hand-written parser — **no Tree-sitter, no second grammar** (spec-mandated). BUT the locked `internal/` layout gains four packages (`analysis`, `language`, `lsp`, `formatter`). Structural addition → recorded in Complexity Tracking below. `mcpserver/` stays public; the new packages are `internal/` engine packages consistent with the existing style. |
| **V. Boringly Portable** | ✅ Pure Go, `CGO_ENABLED=0`, three OSes. Target zero new third-party deps (hand-rolled JSON-RPC + minimal LSP types). No host-shell or WSL assumptions. |
| **VI. Test-First, Multi-Layer Verification** | ✅ Golden + binary-integration + fuzz are core to the feature (SC-002/004/009). Protocol tests drive a real subprocess. Docs remain tested fixtures. |
| **VII. AI-Native, Secure by Default** | ✅ The analysis surface is read-only by construction (FR-028). Secrets are never expanded into messages. Matches the "read-only by default" agent posture. |
| **VIII. Go Engineering Discipline** | ✅ Errors wrapped with `%w`; every goroutine (debounce/analysis worker) has a clear owner and a context-driven exit; constructors over globals; no optimization before benchmarks; `golangci-lint` clean under gofumpt/goimports. |

**Engineering Constraints**: Docker-only testing honored. Locked layout deviation justified below. Backward compatibility: all changes are additive — the `diag.Diagnostic` and `config.Compose` changes keep existing call sites working (empty code default; a disk-backed source store preserves current behavior). Surface changes carry docs: `rune lsp` / `rune analyze` docs and the diagnostic-code catalog ship in the same effort; `docs/GRAMMAR.md` is unaffected (no grammar change).

**Gate result**: PASS with one justified deviation (see Complexity Tracking).

## Project Structure

### Documentation (this feature)

```text
specs/011-rune-lsp/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   ├── cli-analyze.md       # `rune analyze` command contract (flags, output, exit codes)
│   ├── cli-lsp.md           # `rune lsp` command contract (flags, transport, logging)
│   ├── lsp-capabilities.md  # advertised capabilities + supported requests/notifications
│   └── diagnostic-codes.md  # the stable RUNE#### catalog (public contract, FR-010)
├── checklists/
│   └── requirements.md  # spec quality checklist (from /speckit-specify)
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root)

```text
internal/
├── analysis/            # NEW — reusable analysis service (shared by CLI, analyze, lsp, MCP)
│   ├── service.go           # Service.Analyze(ctx, req) → *Snapshot
│   ├── snapshot.go          # Snapshot (File, Sources, Diagnostics, Symbols, Imports)
│   ├── document.go          # OpenDocument + edit application
│   ├── source.go            # SourceStore interface, DiskSourceStore, OverlaySourceStore
│   ├── workspace.go         # Workspace / root detection / import graph
│   └── options.go           # AnalyzeRequest, service options
├── language/            # NEW — language intelligence over a snapshot
│   ├── index.go             # Symbol Index (ByName, ByQualified, ByDocument)
│   ├── symbol.go            # Symbol + SymbolKind
│   ├── scope.go             # scope lookup for variables/params
│   ├── completion.go        # context-aware completion engine
│   ├── definition.go        # definition resolution (incl. cross-file)
│   ├── hover.go             # hover content assembly
│   └── builtin.go           # single language registry (built-ins, settings, attributes)
├── lsp/                 # NEW — protocol glue only (no language logic)
│   ├── server.go            # lifecycle: initialize/initialized/shutdown/exit
│   ├── jsonrpc.go           # Content-Length framing + JSON-RPC 2.0 dispatch
│   ├── protocol.go          # minimal typed LSP 3.17 payload subset
│   ├── handler.go           # request routing
│   ├── documents.go         # didOpen/didChange/didSave/didClose → overlay
│   ├── diagnostics.go       # debounce, cancellation, publishDiagnostics
│   ├── completion.go        # LSP completion ↔ language.completion
│   ├── definition.go        # LSP definition ↔ language.definition
│   ├── hover.go             # LSP hover ↔ language.hover
│   ├── symbols.go           # documentSymbol ↔ language.index
│   ├── formatting.go        # formatting ↔ internal/formatter
│   └── convert.go           # LineIndex: byte-span ↔ UTF-16 position conversion
├── formatter/           # NEW — extracted from internal/cli/fmt.go
│   └── formatter.go         # Format(*ast.File) string (canonical formatter)
├── diag/                # CHANGED — Diagnostic gains Code + Related []RelatedLocation
├── config/              # CHANGED — Compose reads imports via injected SourceStore
├── parser/              # CHANGED — hardened recovery mode; InvalidStmt as needed
├── analyzer/            # CHANGED — stable codes on emitted diagnostics; tolerate partial ASTs
├── ast/                 # (unchanged, or minimal InvalidStmt node)
├── lexer/ token/ eval/ runtime/ cache/ …  # unchanged

cmd/rune/
├── lsp.go               # NEW — cobra command: rune lsp [--log-file --log-level]
├── analyze.go           # NEW — cobra command: rune analyze [path] [--json]
└── root.go serve.go …   # register new commands; thread build version into lsp serverInfo

editors/                 # NEW — editor integration (Milestone 5)
├── vscode/              # VS Code client extension
└── README.md            # Neovim, Zed, Helix setup docs

test/ or *_test.go       # protocol integration harness + fuzz targets colocated per Go convention
testdata/lsp/            # NEW fixtures: basic/ imports/ cycles/ incomplete/ unicode/ multimodule/
```

**Structure Decision**: Rune is a single Go module with a locked `internal/` engine layout and a public `mcpserver/`. This feature adds four `internal/` packages that layer strictly *above* the existing engine (they import `parser`/`analyzer`/`config`/`diag`, never the reverse) and extracts the existing formatter into its own package. The LSP is wired as `cmd/rune/lsp.go` (and `analyze.go`) — new cobra subcommands on the existing binary, not a separate program. This matches the "one binary, one parser, one analyzer, one formatter" product decision and keeps the language engine the single source of truth.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Four new `internal/` packages (`analysis`, `language`, `lsp`, `formatter`) beyond the layout locked in Principle IV | The LSP is a genuinely new interface surface (protocol, symbol intelligence, source overlay) that cannot live in the execution-oriented `cli` package without coupling analysis to execution. A shared `analysis.Service` is explicitly required so CLI/analyze/lsp/MCP report identical diagnostics (FR-002). | Putting it all in `internal/cli` was rejected: it would entangle the read-only analysis surface with the execution engine, violate the "analysis service consumed by multiple interfaces" requirement, and make the safety guarantee (no execution) harder to enforce and test. Reusing the engine packages unchanged is not possible because `Compose` reads from disk and `diag` lacks codes/related locations. |
| Extract formatter to `internal/formatter` (moves code out of `internal/cli`) | FR-020 requires the LSP to call the formatter directly (not shell out); a private `cli.formatFile` is unreachable from `internal/lsp` without an import cycle (cli imports everything). | Leaving it in `cli` and having `lsp` import `cli` was rejected — `cli` depends on the runtime/execution stack, which the read-only LSP must not pull in. Extraction is a pure move with the CLI as its first caller. |
