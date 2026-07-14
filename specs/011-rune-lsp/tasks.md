---
description: "Task list for Rune Language Server Protocol"
---

# Tasks: Rune Language Server Protocol

**Input**: Design documents from `/specs/011-rune-lsp/`

**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Tests**: INCLUDED. The constitution mandates test-first (Principle VI) and the spec defines golden, protocol-integration, and fuzz suites (SC-002/004/009). Test tasks precede the implementation they cover.

**Testing policy**: All Go tests run **inside Docker Compose** (`docker-compose run --rm test go test ./...`), never on host.

**Organization**: Tasks are grouped by user story. Foundational (Phase 2) produces a complete `analysis.Snapshot` (diagnostics + symbols + import graph) shared by every interface.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: US1–US8 (setup/foundational/polish carry no story label)

## Path Conventions

Single Go module `github.com/rune-task-runner/rune`. New engine packages under `internal/`; commands under `cmd/rune/`; editor clients under `editors/`; fixtures under `testdata/lsp/`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Scaffold the new packages, command stubs, and fixtures so later tasks have a home.

- [X] T001 Scaffold new packages with doc-comment package files: `internal/analysis/`, `internal/language/`, `internal/lsp/`, `internal/formatter/` (one `doc.go` each)
- [X] T002 [P] Create LSP test fixtures skeleton in `testdata/lsp/{basic,imports,cycles,incomplete,unicode,multimodule}/` with the MVP Runefile from `quickstart.md` in `basic/`
- [X] T003 [P] Add `rune lsp` and `rune analyze` cobra command stubs (return "not implemented") in `cmd/rune/lsp.go` and `cmd/rune/analyze.go`, registered in `cmd/rune/root.go`
- [X] T004 [P] Add benchmark files with signatures (`BenchmarkParseRunefile`, `BenchmarkAnalyzeRunefile`, `BenchmarkBuildSymbolIndex`, `BenchmarkCompletion`, `BenchmarkImportedFileInvalidation`) in `internal/analysis/bench_test.go` and `internal/language/bench_test.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The shared analysis engine, model extensions, and refactors that ALL user stories build on. Completing this phase yields an `analysis.Service` that produces complete snapshots with diagnostics identical to the execution path.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Diagnostic model & code catalog (FR-007/009/010)

- [X] T005 [P] Extend `diag.Diagnostic` with `Code string` and `Related []RelatedLocation`, and add the `RelatedLocation` type, in `internal/diag/diagnostic.go` (keep existing constructors working; add code-carrying helpers)
- [X] T006 [P] Add the RUNE#### code constants catalog per `contracts/diagnostic-codes.md` in `internal/diag/codes.go`
- [X] T007 Attach stable codes at every emit site (and `Related` for cycles) in `internal/parser/`, `internal/analyzer/`, `internal/config/`; regenerate affected golden files deliberately
- [X] T008 [P] Golden diagnostic tests: one fixture per RUNE#### code asserting `{Code, Severity, Range}` (and `Related` for RUNE2003/RUNE3002) in `internal/analyzer/` / `internal/parser/` testdata

### Source overlay & import composition (FR-003)

- [X] T009 [P] Define `SourceStore` interface + `DiskSourceStore` + `OverlaySourceStore` in `internal/analysis/source.go`
- [X] T010 [P] `OpenDocument` type + incremental edit application in `internal/analysis/document.go`, with unit test + fuzz target `FuzzApplyEdits` for edit application
- [X] T011 Refactor `config.Compose`/`spliceImports`/`loadMods` to read imports/mods through an injected reader (adapting `SourceStore`) instead of `os.ReadFile`, in `internal/config/compose.go`; update CLI (`internal/cli/run.go`, `internal/cli/serve.go`) to pass a disk-backed store (behavior preserved)

### Formatter extraction (FR-020)

- [X] T012 Extract `formatFile` + helpers into `internal/formatter/formatter.go` as `Format(*ast.File) string`; update `internal/cli/fmt.go` (`fmtRewrite`, `--fmt`) to call it; move `fmt_test.go` to `internal/formatter/` (byte-for-byte output preserved)

### Parser recovery (FR-004/005)

- [X] T013 Parser recovery: verified the existing drop-and-continue recovery keeps valid declarations around broken regions (`parser.TestRecoveryKeepsValidDeclarations`), documented the unterminated-group-to-EOF limitation (`TestUnterminatedGroupIsBounded`). `InvalidStmt` intentionally NOT added — drop-and-continue leaves no broken subtree, so it has no consumer (see research.md R2 outcome)
- [X] T014 Analyzer runs on partially-recovered ASTs (surviving declarations) without panic and still flags their errors — `analyzer.TestAnalyzeRecoveredFile`. Moot `InvalidStmt`-skip per T013 decision
- [X] T015 [P] Fuzz target `FuzzParseRecover` asserting terminate / no-panic / all diagnostic ranges in-bounds (FR-005/SC-004) in `internal/parser/`

### Language registry & symbol index (FR-026/027, R6/R9)

- [X] T016 [P] Build the single language registry (`Builtin{Name,Kind,Signature,Documentation,IntroducedIn}` for builtins, settings, attributes, executors) in `internal/language/builtin.go`
- [X] T017 Refactor analyzer validation for invalid setting/attribute/executor (RUNE2007/2008/2009) to consult the registry so completion and validation agree, in `internal/analyzer/`
- [X] T017a Implement the undocumented-public-task check emitting RUNE2010 (warning severity, non-gating per FR-008a) for public tasks with no doc comment, in `internal/analyzer/`, with a golden test
- [X] T018 [P] `Symbol` + `SymbolKind` + `ScopeTree`/scope lookup in `internal/language/symbol.go` and `internal/language/scope.go`
- [X] T019 `Index` (`ByName`/`ByQualified`/`ByDocument`) built from the composed `*ast.File` — cross-file attribution via `Sp.File`, namespaced (`ns::task`) tasks, private tagging — in `internal/language/index.go`, with unit test

### Analysis service (FR-002)

- [X] T020 [P] `Snapshot` type in `internal/analysis/snapshot.go`; `AnalyzeRequest` + service options in `internal/analysis/options.go`
- [X] T021 `Workspace` + root detection order (FR-021) + `ImportGraph` (forward/reverse edges) in `internal/analysis/workspace.go`, with unit test for detection order
- [X] T022 `Service.Analyze(ctx, req) → *Snapshot` orchestrating parse → compose(via store) → analyze → build index → build import graph, in `internal/analysis/service.go`, with unit test
- [X] T023 [P] Safety guarantee (FR-028/SC-005): structural test asserting `internal/analysis`, `internal/language`, `internal/lsp` import no `internal/runtime/*`, `os/exec`, or net packages, plus a behavioral no-side-effects test, in `internal/analysis/safety_test.go`
- [X] T024 [P] Implement real benchmark bodies for `BenchmarkParseRunefile`, `BenchmarkAnalyzeRunefile`, `BenchmarkBuildSymbolIndex` (T004)

**Checkpoint**: `analysis.Service` produces complete snapshots; diagnostics match the execution path. All user stories can now begin.

---

## Phase 3: User Story 1 - Live diagnostics while editing (Priority: P1) 🎯 MVP

**Goal**: Editors show Rune syntax/semantic/project errors live as the developer types, running nothing.

**Independent Test**: Point an LSP client at `rune lsp`, open the `basic` fixture with `deploy: missing`; an error appears on `missing`; fixing it clears the error; introducing an unterminated string mid-edit keeps surrounding tasks analyzed and does not crash the server.

**Note**: This phase delivers the *diagnostics* MVP (US1). The spec's full MVP scenario (SC-006) additionally requires US3/US4/US5 and is validated at T062, not here.

### Tests for User Story 1 ⚠️ (write first)

- [X] T025 [P] [US1] `LineIndex` unit tests across the unicode/line-ending matrix (ASCII, Ukrainian, emoji, combining marks, CRLF, LF, empty lines, EOF) — SC-008 — in `internal/lsp/convert_test.go`
- [X] T026 [P] [US1] Fuzz targets `FuzzUTF16Position` (convert) and `FuzzJSONRPC` (malformed messages) in `internal/lsp/`
- [X] T027 [P] [US1] Protocol integration test driving a real `rune lsp` subprocess: initialize → initialized → didOpen → publishDiagnostics → didChange(clear) → shutdown → exit (SC-009), in `internal/lsp/protocol_test.go`

### Implementation for User Story 1

- [X] T028 [P] [US1] `LineIndex` byte-offset↔UTF-16 position/range conversion (`ByteOffsetToPosition`, `PositionToByteOffset`, `SpanToRange`) in `internal/lsp/convert.go`
- [X] T029 [US1] JSON-RPC 2.0 `Content-Length` framing + read/write/dispatch loop in `internal/lsp/jsonrpc.go`
- [X] T030 [US1] Minimal typed LSP 3.17 payload subset (initialize params/result, text-sync, diagnostics, position/range) in `internal/lsp/protocol.go`
- [X] T031 [US1] Server lifecycle (initialize/initialized/shutdown/exit) + capabilities advertisement matching `contracts/cli-lsp.md` + `serverInfo.version` from the real build version, in `internal/lsp/server.go` and `internal/lsp/handler.go`
- [X] T032 [US1] Document sync didOpen/didChange(incremental)/didSave/didClose applying edits to `OverlaySourceStore`, in `internal/lsp/documents.go`
- [X] T033 [US1] Diagnostics pipeline: map `diag.Diagnostic`→LSP (code/severity/range-via-LineIndex/relatedInformation), ~100 ms debounce + context cancellation + version guard, `publishDiagnostics`, in `internal/lsp/diagnostics.go`
- [X] T034 [US1] Imported-file change propagation: on `initialized`, register file watchers via dynamic `client/registerCapability` for `**/Runefile` and `**/*.rune` (fall back to client-side watching if the client lacks the capability); on `workspace/didChangeWatchedFiles`, invalidate transitive importers → re-analyze roots → republish (FR-022), covering on-disk imports not open in the editor, in `internal/lsp/diagnostics.go` + `internal/analysis/workspace.go`
- [X] T035 [US1] Wire the real `rune lsp` command: stdio transport, `--log-file`, `--log-level`, logs never on stdout (FR-012), in `cmd/rune/lsp.go`

**Checkpoint**: Live diagnostics work end-to-end in an editor — the MVP.

---

## Phase 4: User Story 2 - Standalone analysis command (Priority: P1)

**Goal**: `rune analyze` reports diagnostics (with codes, across transitive imports) and correct exit codes without running anything.

**Independent Test**: `rune analyze testdata/lsp/basic/Runefile` prints `…: error[RUNE2001]: unknown dependency "missing"` + summary and exits 3; fixed file exits 0; `--json` emits structured diagnostics; no task runs.

**Note**: US2 has **no dependency on US1** (no LSP scaffolding needed) — it is the smallest increment over the foundation and may be delivered first.

### Tests for User Story 2 ⚠️ (write first)

- [X] T036 [P] [US2] Golden test for human output format `FILE:LINE:COL: SEVERITY[CODE]: MESSAGE` + summary counts, in `internal/cli/testdata/`
- [X] T037 [P] [US2] Golden test for `--json` output schema, and exit-code tests (0 clean / 3 errors / 1 internal), in `internal/cli/`
- [X] T038 [P] [US2] Parity test: `rune analyze` diagnostics equal the LSP snapshot diagnostics for the same fixture (SC-002), in `internal/cli/`

### Implementation for User Story 2

- [X] T039 [US2] Implement `rune analyze [path] [--json]` over `analysis.Service` — default `Runefile`, transitive imports (FR-009a), human formatter + summary, exit 0/1/3 (FR-025) — in `cmd/rune/analyze.go` and `internal/cli/analyze.go`
- [X] T040 [US2] `--json` structured output (code/severity/message/range/file/related) in `internal/cli/analyze.go`

**Checkpoint**: `rune analyze` is a usable CI gate with the same diagnostics as the editor.

---

## Phase 5: User Story 3 - Context-aware completion (Priority: P2)

**Goal**: Cursor-context-aware completion for dependencies, variables/params, settings, attributes, executors, and builtins.

**Independent Test**: Type `deploy env: bui` → `build` suggested with params + docs; accepting yields `deploy env: build`. Also `set wor`→`working-directory`, `[conf`→`confirm`, `(py`→`python`, `os_`→`os_family()`.

**Depends on**: Foundational (index + registry) and US1 (LSP server scaffolding).

### Tests for User Story 3 ⚠️ (write first)

- [X] T041 [P] [US3] Golden completion tests per context (dependency incl. private same-file-only, interpolation, setting, attribute, executor, builtin) using `testdata/lsp/` fixtures, in `internal/language/completion_test.go`

### Implementation for User Story 3

- [X] T042 [US3] Cursor-context detection (dependency / interpolation / `set` / attribute-bracket / executor / builtin-prefix) in `internal/language/completion.go`
- [X] T043 [US3] Completion result assembly: tasks (namespaced; private same-file-only per FR-019a), scope-visible variables+params (params ranked first), settings, attributes, executors, builtins — each with signature + documentation from the registry, in `internal/language/completion.go`
- [X] T044 [US3] LSP completion handler + trigger characters `[ ( . { :` mapping, enabling `completionProvider`, in `internal/lsp/completion.go`

**Checkpoint**: Completion works in-editor for all six contexts.

---

## Phase 6: User Story 4 - Go-to-definition & cross-file navigation (Priority: P2)

**Goal**: Jump to declarations for dependencies, variables, parameters, imported namespaces, and imported tasks — across files.

**Independent Test**: Definition on `build` in `deploy: build` jumps to the task; definition on `backend.build` (imports fixture) opens the imported file at `build`; an open+unsaved import resolves via overlay.

**Depends on**: Foundational (index) and US1 (LSP server scaffolding).

### Tests for User Story 4 ⚠️ (write first)

- [X] T045 [P] [US4] Definition unit + integration tests: same-file (task/var/param) and cross-file (`backend.build`) using `testdata/lsp/imports/`, plus overlay-for-open-import case, in `internal/language/definition_test.go`

### Implementation for User Story 4

- [X] T046 [US4] Definition resolution for dependency, variable reference, parameter interpolation, imported namespace, and imported task (cross-file via `Sp.File`) in `internal/language/definition.go`
- [X] T047 [US4] LSP definition handler enabling `definitionProvider`, in `internal/lsp/definition.go`

**Checkpoint**: Navigation works within and across imported Runefiles.

---

## Phase 7: User Story 5 - Hover documentation (Priority: P3)

**Goal**: Hover shows task/parameter/attribute/builtin documentation from the shared registry.

**Independent Test**: Hover `build` shows signature + doc + executor + group + `Defined in:`; hover a parameter shows type/default/declaring task; hover an attribute/builtin shows registry docs.

**Depends on**: Foundational (index + registry) and US1.

### Tests for User Story 5 ⚠️ (write first)

- [X] T048 [P] [US5] Golden hover tests for task, parameter, attribute, and builtin, in `internal/language/hover_test.go`

### Implementation for User Story 5

- [X] T049 [US5] Hover content assembly (task signature+doc+executor+group+location; parameter; attribute; builtin) in `internal/language/hover.go`
- [X] T050 [US5] LSP hover handler enabling `hoverProvider`, in `internal/lsp/hover.go`

**Checkpoint**: Hover works for all symbol kinds.

---

## Phase 8: User Story 6 - Document symbols / outline (Priority: P3)

**Goal**: Editor outline lists settings, variables, imports, tasks, and modules.

**Independent Test**: Open a Runefile with each category; the outline lists them under their categories and each entry navigates to its declaration.

**Depends on**: Foundational (index) and US1.

### Tests for User Story 6 ⚠️ (write first)

- [X] T051 [P] [US6] Golden document-symbol test for a multi-category fixture, in `internal/lsp/symbols_test.go`

### Implementation for User Story 6

- [X] T052 [US6] `documentSymbol` projection from the `Index` (settings/variables/imports/tasks/modules with range + selectionRange) enabling `documentSymbolProvider`, in `internal/lsp/symbols.go`

**Checkpoint**: Outline reflects file structure.

---

## Phase 9: User Story 7 - Canonical formatting (Priority: P3)

**Goal**: Formatting returns a full-document edit from Rune's canonical formatter over the (possibly unsaved) buffer, spawning no process and writing no file.

**Independent Test**: Format a poorly-formatted unsaved buffer → result equals `formatter.Format`; formatting formatted content is a no-op; no child process, no file write.

**Depends on**: Foundational (formatter extraction T012) and US1.

### Tests for User Story 7 ⚠️ (write first)

- [X] T053 [P] [US7] Formatting handler tests: output equals `formatter.Format`, idempotent, no edit on severe parse errors, in `internal/lsp/formatting_test.go`

### Implementation for User Story 7

- [X] T054 [US7] LSP formatting handler calling `internal/formatter.Format` over the overlay buffer, returning one full-document `TextEdit`, enabling `documentFormattingProvider`, in `internal/lsp/formatting.go`

**Checkpoint**: Formatting works on unsaved content via the canonical formatter.

---

## Phase 10: User Story 8 - Editor setup out of the box (Priority: P3)

**Goal**: VS Code extension + Neovim/Zed/Helix configuration make `rune lsp` reachable with minimal setup.

**Independent Test**: Follow the documented setup for one editor; opening a Runefile starts the server and shows diagnostics; each advertised capability responds.

**Depends on**: US1 (server) and the capability stories the editor will exercise.

### Implementation for User Story 8

- [X] T055 [P] [US8] VS Code client extension (language config for `runefile`, launches `rune lsp`) in `editors/vscode/`
- [X] T056 [P] [US8] Neovim, Zed, and Helix setup documentation in `editors/README.md`
- [ ] T057 [US8] Manual editor validation per `quickstart.md` §9 across at least VS Code + one other editor

**Checkpoint**: `rune lsp` works in common editors.

---

## Phase 11: Polish & Cross-Cutting Concerns

**Purpose**: Performance validation, docs, extra fuzz, and the full CI gate.

- [ ] T058 [P] Implement `BenchmarkCompletion` + `BenchmarkImportedFileInvalidation` and validate SC-010 targets (no optimization before profiling), in `internal/language/` and `internal/analysis/`
- [ ] T059 [P] Additional fuzz: completion and definition at arbitrary positions, in `internal/lsp/`
- [ ] T060 [P] Documentation: `rune lsp` + `rune analyze` guides and the diagnostic-code catalog in `docs/`; update `README.md`/`CONTRIBUTING.md`; ensure `docs-verify` passes
- [ ] T061 Run the full gate inside Docker: `go test ./...`, `-race`, `golangci-lint`, golden, fuzz-smoke, docs-verify, release-dryrun
- [ ] T062 Run `quickstart.md` end-to-end validation (all nine scenarios); SC-006 requires US1+US3+US4+US5 complete

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: no dependencies.
- **Foundational (Phase 2)**: depends on Setup — **BLOCKS all user stories**.
- **US1 (Phase 3)** and **US2 (Phase 4)**: both P1; depend only on Foundational. US2 is independent of US1.
- **US3–US7 (Phases 5–9)**: depend on Foundational **and** US1's LSP server scaffolding (they add handlers to the same server). Independent of each other.
- **US8 (Phase 10)**: depends on US1 (server) and whichever capabilities it demonstrates.
- **Polish (Phase 11)**: depends on all desired stories.

### Critical shared prerequisites (inside Foundational)

- Diagnostic model + codes (T005–T008) → used by US1, US2.
- Source overlay + Compose refactor (T009–T011) → used by US1 (overlays) and all analysis.
- Registry + symbol index (T016–T019) → used by US3, US4, US5, US6 (and registry by US1/US2 validation).
- Formatter extraction (T012) → used by US7.
- `analysis.Service` (T020–T022) → used by everything.

### Parallel Opportunities

- Setup: T002, T003, T004 in parallel.
- Foundational: T005‖T006, T009‖T010, T016‖T018, T020, plus tests T008/T015/T023/T024 in parallel with their siblings once deps land.
- After Foundational: US1 and US2 can proceed in parallel; once US1's server exists, US3/US4/US5/US6/US7 can proceed in parallel (different files).
- Within a story: `[P]` test tasks run together; handlers in distinct files run together.

---

## Parallel Example: Foundational diagnostic model

```bash
Task: "T005 Extend diag.Diagnostic with Code + Related in internal/diag/diagnostic.go"
Task: "T006 Add RUNE#### catalog in internal/diag/codes.go"
# then, after T005/T006:
Task: "T008 Golden diagnostic tests per code"
```

## Parallel Example: LSP capability stories (after US1)

```bash
Task: "US3 completion engine in internal/language/completion.go"
Task: "US4 definition resolution in internal/language/definition.go"
Task: "US5 hover assembly in internal/language/hover.go"
```

---

## Implementation Strategy

### MVP First

1. Phase 1 Setup → Phase 2 Foundational (the bulk of the engineering).
2. Phase 3 US1 (live diagnostics) → **STOP and validate the MVP** in a real editor.
3. Optionally deliver Phase 4 US2 (`rune analyze`) alongside/first — it is independent and small, and unlocks CI gating immediately.

### Incremental Delivery

Foundation → US1 (MVP, live diagnostics) → US2 (analyze/CI) → US3 (completion) → US4 (definition) → US5 (hover) → US6 (symbols) → US7 (formatting) → US8 (editors). Each story adds value without breaking prior ones.

### Notes

- `[P]` = different files, no dependency on incomplete tasks.
- Write-first test tasks (⚠️) must fail before their implementation lands (Principle VI).
- Commit after each task or logical group; run the Docker test suite — never on host.
- The safety guarantee (T023) is load-bearing: keep `analysis`/`language`/`lsp` free of `runtime`/`os-exec`/net imports.
