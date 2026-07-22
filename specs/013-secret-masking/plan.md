# Implementation Plan: Secret Masking & Sanitization

**Branch**: `013-secret-masking` | **Date**: 2026-07-21 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/013-secret-masking/spec.md`

## Summary

Rune will mask secret values in everything it emits — task stdout/stderr
passthrough, echoed command lines, its own status/log/error lines, and MCP tool
results — so credentials never reach a terminal transcript or an agent's chat
history. Secrets are identified from the *names* of variables in the task's
effective environment (built-in sensitive patterns such as `TOKEN`, `SECRET`,
`PASSWORD`, `API_KEY`) plus explicit author declarations; masking replaces
verbatim occurrences with `***` at emission time.

Technical approach (from [research.md](research.md)): a new dependency-free
`internal/mask` package provides a secret-value `Set` (derived from env pairs,
declarations `set secrets := [...]`, exemptions `set unmasked := [...]`,
minimum length 4, multi-line values split per line) and a concurrent, streaming
`io.Writer` wrapper with bounded carry for chunk-spanning values. The wrapper
is installed once, around `Options.Stdout`/`Options.Stderr` at engine
construction — the architecture's single choke point — so the CLI path, shell
echo, agent write-back, and the MCP adapter's buffers are all covered without
executor changes. When the set is empty the writers are not wrapped, making the
no-secrets byte-identical guarantee structural.

## Technical Context

**Language/Version**: Go 1.25.0 (pure Go, `CGO_ENABLED=0`)

**Primary Dependencies**: stdlib only for the new package; existing vendored
`mvdan.cc/sh/v3` (shell executor) and `mark3labs/mcp-go` (MCP) are untouched

**Storage**: N/A (no persistence; secret set is derived per run, held in memory)

**Testing**: Docker-only harness (`docker-compose run --rm test go test ./...`);
layers: unit (`internal/mask`), binary integration (`test/integration`), MCP
stdio end-to-end, golden corpus (`testdata/corpus`), docs-as-fixtures
(`test/docs`)

**Target Platform**: Linux, macOS, Windows — single static binary

**Project Type**: CLI tool + embeddable library (`mcpserver/`) — existing
locked layout

**Performance Goals**: SC-004 — a task producing 10 MB of output completes
within 10% of its unmasked wall-clock time; no perceptible latency added to
interactive streaming

**Constraints**: masking applied at emission time (no post-processing window,
FR-003); byte-identical output for secret-free Runefiles (FR-008); bounded
memory (carry ≤ longest secret − 1 bytes per writer); concurrent-safe (parallel
tasks share writers)

**Scale/Scope**: secret sets are small (typically < 20 values); output sizes up
to CI-log scale (tens of MB); two new settings, one new internal package, no
new CLI flags or subcommands

## Constitution Check

*GATE: evaluated against Constitution v1.0.0 before Phase 0; re-evaluated after
Phase 1 design — PASS (one structural addition justified in Complexity
Tracking).*

| Principle | Verdict | Notes |
|---|---|---|
| I. Command Runner, Not a Build System | PASS | No execution/caching semantics change. `[cache]` is file-based; no task output text is stored or replayed. |
| II. Errors Are a Feature | PASS | `set secrets`/`set unmasked` values are evaluated via `config.ResolveSettings`'s existing `evalList`, which emits positioned diagnostics; unknown-setting typos get RUNE2008 from the analyzer registry. Malformed declarations fail before execution with `file:line:col` + caret (FR-009). |
| III. Minimal, Total DSL | PASS | No expression-language growth (no loops/recursion/new operators). Two new *settings* follow the direct precedent of feature 010 (`minimum_version`); settings are declarations, not expressions. |
| IV. Hand-Written Front End, Idiomatic Go | PASS* | No parser/lexer changes (settings reuse existing grammar). One new small, focused package `internal/mask` — a structural addition to the locked layout, justified in Complexity Tracking. |
| V. Boringly Portable | PASS | Pure Go, stdlib only, byte-oriented masking identical on all OSes; no shell dependence. |
| VI. Test-First, Multi-Layer Verification | PASS | Red-Green-Refactor; unit + integration + MCP e2e + golden corpus + docs fixtures (research D7). Existing goldens unchanged by construction (empty-set path unwrapped). |
| VII. AI-Native, Secure by Default | PASS | This feature *implements* the principle's remaining gap: task output. Masking is on by default, no agent-facing off switch (FR-006); secrets still come only from the environment. |
| VIII. Go Engineering Discipline | PASS | No new deps; no Aho-Corasick without a profile; wrapper has a clear owner and flush lifecycle; `%w` wrapping; golangci-lint/gofumpt clean. |

**Engineering Constraints check**: Docker-only testing — all test commands in
[quickstart.md](quickstart.md) run through `docker-compose`. Backward
compatibility — files without the new settings behave identically (masking from
built-in patterns is new *output* behavior, not a DSL-semantics change; the
DSL surface itself is opt-in). Surface changes carry docs — `docs/GRAMMAR.md`,
`docs/runefile.md`, how-to, MCP doc, and runnable example ship in the same PR
(research D8).

## Project Structure

### Documentation (this feature)

```text
specs/013-secret-masking/
├── plan.md              # This file
├── research.md          # Phase 0 — decisions D1–D8
├── data-model.md        # Phase 1 — entities & derivation rules
├── quickstart.md        # Phase 1 — runnable validation guide
├── contracts/
│   └── secret-masking.md  # Phase 1 — grammar + behavior contract
└── tasks.md             # Phase 2 (/speckit-tasks — not created here)
```

### Source Code (repository root)

```text
internal/
├── mask/                        # NEW package (research D3)
│   ├── set.go                   # Set derivation: patterns, declarations,
│   │                            #   exemptions, min length, multi-line split
│   ├── set_test.go
│   ├── writer.go                # streaming io.WriteCloser wrapper w/ carry
│   └── writer_test.go
├── language/builtin.go          # register settings: secrets, unmasked
├── config/settings.go           # Settings.Secrets/.Unmasked + switch cases
└── cli/
    ├── run.go                   # build mask.Set after buildEnv (+ union of
    │                            #   [env] attr values); wrap Options writers
    │                            #   at engine construction; flush lifecycle
    └── mcp.go                   # same wrap on the MCP adapter's engine

docs/
├── GRAMMAR.md                   # settings grammar additions
├── runefile.md                  # settings reference rows
├── mcp.md                       # agent-facing guarantee note
├── how-to/
│   ├── secret-masking.md        # NEW: patterns, placeholder, min length,
│   │                            #   guarantee boundary (FR-010)
│   └── settings-and-dotenv.md   # pitfalls cross-link
└── examples/secret-masking/     # NEW runnable example (docs-verify fixture)
    ├── Runefile
    └── README.md

test/integration/
└── secret_masking_test.go       # NEW binary-level tests incl. MCP stdio e2e

testdata/corpus/
├── full.rune                    # + set secrets / set unmasked lines
└── full.ast                     # regenerated deliberately
```

**Structure Decision**: Reuses the locked `internal/cli` / `internal/runtime` /
`mcpserver` split; the only structural addition is `internal/mask`, a leaf
package with no dependencies on other engine packages (it imports only stdlib),
consumed by `internal/cli`. Executors and `mcpserver/` are untouched — masking
lives entirely at the writer-plumbing layer.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| New `internal/mask` package (Principle IV locks the package layout; structural changes require justification) | Masking is a self-contained text-transformation concern with its own dense unit-test surface (chunk boundaries, overlapping values, concurrency, flush semantics) and is consumed from two call sites (`cli`, MCP adapter path) | Folding it into `internal/cli` bloats an already-large package with logic that has zero CLI coupling and makes the correctness-critical writer untestable in isolation; folding into `internal/runtime` is wrong because masking also covers Rune's own status lines, which are not runtime output |
