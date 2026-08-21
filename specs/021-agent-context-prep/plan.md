# Implementation Plan: Agent Context Prep (`[context]`)

**Branch**: `021-agent-context-prep` | **Date**: 2026-08-21 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/021-agent-context-prep/spec.md`

**Companion**: [implementation-detail.md](implementation-detail.md) — the approved
code-level task breakdown (verified against the codebase); tasks.md is derived
from it.

## Summary

Add one task attribute, `[context]`, marking a single task as the project's
context hook. Rune runs it when an agent surface starts and injects its
masked, capped output into the agent's context before any task is chosen:
`ServeMCP` passes it as MCP server `instructions` (delivered in every
session's initialize result), and `executeAgent` prepends it to the provider
prompt under fixed delimiters. The gathering primitive is one new method on
the existing `mcpAdapter` — `gatherContext` — which reuses the masked
`Call()` path, a 10-second timeout, an 8 KiB cap, and degrades to a one-line
notice on any failure (never blocking). The parser gains one bare-attribute
case; the analyzer enforces the hook contract (singleton, no `[confirm]`, no
defaultless parameters, not an `agent` task); the formatter needs no change.

## Technical Context

**Language/Version**: Go 1.25 (`go.mod`); no new dependencies — the MCP SDK
(`github.com/modelcontextprotocol/go-sdk v1.6.1`) already exposes
`ServerOptions.Instructions`, verified in the module cache

**Primary Dependencies**: existing internal packages only: `internal/ast`
(one constant), `internal/parser` (one case label), `internal/analyzer` (one
check method), `internal/cli` (new `contextprep.go`, wiring in `serve.go` /
`agentexec.go` / `mcp.go`), public `mcpserver/` (one `Options` field passed
through to the SDK)

**Storage**: N/A

**Testing**: `go test` inside the Docker harness (`docker-compose run --rm
test go test ./...`); parser/analyzer unit tests, formatter golden fixture,
`internal/cli` unit tests for gathering (timeout/truncation/masking/degrade),
`mcpserver` in-memory-transport test reading `InitializeResult().Instructions`,
binary-level integration tests with an `sh` stub agent, docs harness re-run

**Target Platform**: Linux, macOS, Windows; the hook respects the existing
`AvailableOn(goos)` rule, so an OS-mismatched hook is absent, not an error

**Project Type**: CLI + embeddable MCP server

**Performance Goals**: one extra task execution per agent-session start /
agent-task invocation, hard-capped at 10 s; zero cost when no `[context]`
task exists (nil lookup short-circuits)

**Constraints**: timeout (10 s) and cap (8 KiB) are compile-time constants —
no flags, env vars, or settings (NFR-002); Runefiles without `[context]`
parse/analyze/format byte-identically (NFR-001); hook output passes the
existing secret mask **before** truncation; context prep may never block a
session or task (FR-005); fixed prompt delimiters and notice format are
verbatim in the spec

**Scale/Scope**: 1 AST constant, 1 parser case, 1 analyzer check, 1 new
~60-line cli file, 3 wiring sites, 1 mcpserver field; ~9 source files plus
tests and 3 docs files. No LSP-specific work (diagnostics flow from the
shared analyzer); no VS Code grammar change needed for a bare attribute

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principle | Verdict | Notes |
|---|-----------|---------|-------|
| I | Command Runner, Not a Build System | PASS | The hook is an explicit, author-declared task run — no caching or freshness-skipping. Degrade-on-failure is visible (stderr warning + in-context notice), never a silent skip of requested work. |
| II | Errors Are a Feature | PASS | All four misuse forms (duplicate hook, `[confirm]`, defaultless param, `agent` executor) are analyzer errors with spans, code RUNE2007, reported before anything runs; the duplicate names both offenders via a related location. |
| III | Minimal, Total DSL | PASS | One bare attribute; zero expression-language change. Attributes are the established extension point (`[network]`, `[no-exit-message]` precedents). |
| IV | Hand-Written Front End, Idiomatic Go | PASS | One case label in the hand-written parser; gathering lives in `internal/cli` beside its consumers; `mcpserver/` stays embeddable — new `Options.Instructions` is plain data, no callback into internals. |
| V | Boringly Portable | PASS | No OS-specific code; the hook body runs under the same `mvdan.cc/sh` executor as any task; OS attributes compose via the existing predicate. |
| VI | Test-First, Multi-Layer Verification | PASS | Red-green ordering preserved from implementation-detail.md: parser test → analyzer tests → gathering fault-injection tests → transport-level instructions test → binary-level integration tests → docs harness. Formatter golden fixture guards round-trip. |
| VII | AI-Native, Secure by Default | PASS (strengthened) | Context flows through the same masking choke point as tool output, then truncates (mask-before-truncate is asserted by test); the hook is never a callable tool; nothing new is remotely reachable — instructions ride the existing initialize result. |
| VIII | Go Engineering Discipline | PASS | `gatherContext` takes a context and enforces a timeout on the hook run; no goroutines, no globals (the timeout var exists only for test override, matching house test patterns); golangci-lint clean required. |

**Engineering Constraints check**: Docker-only testing — respected. Locked
package layout — respected (no new packages; one new file in `internal/cli`).
Backward compatibility — respected (attribute is additive; NFR-001 guards
byte-identical behavior for existing files). Surface changes carry docs —
`docs/GRAMMAR.md`, `docs/runefile.md`, `docs/mcp.md` updated in the same PR
and enforced by the docs harness.

**Post-Phase-1 re-check**: design artifacts add no packages, no DSL surface
beyond the one attribute, no deviations. Gate passes.

## Project Structure

### Documentation (this feature)

```text
specs/021-agent-context-prep/
├── spec.md                    # Feature specification
├── plan.md                    # This file
├── implementation-detail.md   # Approved code-level task breakdown (input to tasks.md)
├── research.md                # Phase 0 output
├── data-model.md              # Phase 1 output
├── quickstart.md              # Phase 1 output
├── contracts/                 # Phase 1 output
│   └── context-attribute.md   # Grammar, analyzer, injection-surface contracts
├── checklists/
│   └── requirements.md        # Spec quality checklist (done)
└── tasks.md                   # Phase 2 output (/speckit-tasks — not created here)
```

### Source Code (repository root)

```text
internal/
├── ast/ast.go                 # + AttrContext constant
├── parser/attribute.go        # + bare-attribute case; attribute_context_test.go
├── analyzer/analyzer.go       # + checkContext(); context_test.go
└── cli/
    ├── contextprep.go         # NEW: contextTask lookup + (*mcpAdapter).gatherContext
    ├── contextprep_test.go    # NEW: fault-injection, masking, truncation tests
    ├── mcp.go                 # Tasks(): exclude the [context] task from tools
    ├── serve.go               # ServeMCP: gather once, pass Options.Instructions
    ├── agentexec.go           # executeAgent: buildAgentPrompt prefix, per invocation
    └── agentprompt_test.go    # NEW: prompt assembly tests

mcpserver/
├── server.go                  # Options.Instructions → mcp.ServerOptions
└── server_test.go             # initialize-result instructions tests

test/integration/context_test.go  # NEW: sh-stub agent end-to-end tests
testdata/fmt/context-attr.{rune,fmt}  # formatter golden fixture

docs/GRAMMAR.md, docs/runefile.md, docs/mcp.md  # attribute + guide updates
```

**Structure Decision**: no new packages. Gathering lives in `internal/cli`
because both agent surfaces already live there and both hold an `mcpAdapter`,
whose masked `Call()` path is the security choke point the feature must reuse
(Principle VII). The public `mcpserver/` package receives only a plain string
option, keeping it embeddable and policy-free.

## Complexity Tracking

No constitution violations — table not needed.
