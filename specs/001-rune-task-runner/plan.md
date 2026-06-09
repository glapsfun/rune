# Implementation Plan: Rune — A Shared Task Runner for Humans and AI Agents

**Branch**: `001-rune-task-runner` | **Date**: 2026-06-08 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-rune-task-runner/spec.md`

## Summary

Rune is a Go-native command-line task runner configured by its own purpose-built DSL (the
`Runefile`). Humans run tasks from the CLI; AI agents and IDEs run the exact same tasks through
a standard agent-tool interface (MCP). The implementation is a hand-written compiler front end
(state-function lexer → recursive-descent parser → Pratt expression parser) feeding a static
analyzer (spanned diagnostics) and an execution engine (dependency DAG, memoization, parallel
fan-out). Task bodies run under a pluggable executor: a cross-platform pure-Go shell
(`mvdan/sh`) by default, or Python/Node/custom interpreters via temp-file exec, or an
**AI-agent** executor that drives an installed agent CLI (`claude`/`codex`/`copilot`). An MCP
server exposes non-private tasks as agent-callable tools; a local interface is always available
while any remote endpoint is opt-in, localhost-bound, and token-gated. Delivery is phased:
MVP = a `just`-class shell runner with static validation (Stories 1–2); v1 = through the
AI/agent surface and CI/CD ergonomics (Stories 3–5).

## Technical Context

**Language/Version**: Go 1.26.x (toolchain present: go1.26.2); target `go 1.24` minimum in
`go.mod` for broad compatibility. Pure Go, **CGO disabled** for portable static binaries.

**Primary Dependencies**:
- `mvdan.cc/sh/v3` (`interp`, `syntax`, `shell`; optionally `moreinterp/coreutils`) — default
  shell executor and `.env` parsing.
- `github.com/modelcontextprotocol/go-sdk` (`mcp`, `jsonschema`) — MCP server + client.
- `github.com/spf13/cobra` — global-flag parsing + shell completions, wrapped by a custom
  dynamic task dispatcher.
- `github.com/fsnotify/fsnotify` — `--watch`.
- `golang.org/x/sync/errgroup` — bounded parallel dependency execution.
- `github.com/fatih/color` — terminal styling (diagnostics, listings).
- Stdlib only for the hand-written `token`/`lexer`/`ast`/`parser`/`analyzer`/`eval` and for
  caching (`crypto/sha256`).

**Storage**: Filesystem only. Inputs: `Runefile`/`.runefile`, imported/`mod` files, `.env`.
Outputs: cache fingerprints as JSON under a project-local `.rune/cache/` directory. No database.

**Testing**: `go test` (table-driven unit tests; golden files for AST dumps and formatter
output); Go native fuzzing (`go test -fuzz`) for lexer + parser; integration tests that run the
**compiled binary** against fixture Runefiles asserting stdout/stderr/exit code; a
compatibility-corpus guard. CI runs lint + the full suite on Linux, macOS, and Windows.

**Target Platform**: Linux, macOS, Windows — single self-contained static binary
(cross-compiled via `goreleaser`).

**Project Type**: Single project — a CLI tool that is internally a small compiler + runtime.

**Performance Goals**: Parse + statically analyze a typical (≤200-task) Runefile in < 50 ms;
sub-100 ms cold start to first task dispatch; a cache hit completes in < 10% of the task's
original wall-clock time (SC-006); parallel fan-out reduces wall-clock vs. sequential on
multi-core machines (SC-009).

**Constraints**: Single static binary, no CGO (Principle V); default `(sh)` executor MUST NOT
shell out to the system shell; expression language MUST stay total/non-Turing-complete
(Principle III); every static-detectable error MUST carry `file:line:col` + caret span and be
reported before any side effect (Principle II); secrets MUST come from the environment / the
agent CLI's own session only, never the Runefile (Principle VII); a remote MCP endpoint MUST be
opt-in, localhost-bound, token-gated (FR-031).

**Scale/Scope**: v1 covers Brief Phases 0–7 (scaffold → front end → analysis → expressions →
execution → multi-language/attributes → CI/CD power → AI/MCP). Phase 8 polish (completions,
`--choose`, corpus) is in-scope for v1 hardening. Expected task files: tens to low-hundreds of
tasks; modules a few levels deep.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principle | Gate | Status |
|---|-----------|------|--------|
| I | Command Runner, Not a Build System | Tasks always-run; caching is per-task opt-in (`[cache]`); no timestamp skipping; cache hits logged | ✅ Design honors (see `contracts/cache-fingerprint.md`) |
| II | Errors Are a Feature | `diag` renders `file:line:col` + caret; `analyzer` runs before any execution; covers unknown-task, undefined-var, cycle, arity | ✅ Front end designed for spans; tokens carry offset/line/col |
| III | Minimal, Total DSL | Expression sublanguage has no loops/recursion; logic lives in bodies | ✅ `eval` is a total tree-walker; grammar bounded (see `contracts/grammar.md`) |
| IV | Hand-Written Front End, Idiomatic Go | State-function lexer + recursive-descent + Pratt; no `goyacc`/ANTLR; no shipped prototype libs | ✅ `internal/{token,lexer,ast,parser}` hand-written |
| V | Boringly Portable | Pure-Go (`CGO_ENABLED=0`); default shell via `mvdan/sh`; single binary; forward-slash path join | ✅ No CGO deps selected; sha256 from stdlib |
| VI | Test-First, Multi-Layer Verification | TDD; golden files; binary-level integration tests; fuzz targets; compat corpus; 3-OS CI | ✅ Testing strategy committed above + `quickstart.md` |
| VII | AI-Native, Secure by Default | MCP first-class; agent read-only by default; author-declared destructiveness; env/CLI-only secrets; remote opt-in/localhost/token | ✅ `contracts/mcp-tools.md` encodes FR-025…FR-031 |

**Result**: PASS — no violations. **Complexity Tracking** table below is empty (nothing to justify).

## Project Structure

### Documentation (this feature)

```text
specs/001-rune-task-runner/
├── plan.md              # This file (/speckit-plan output)
├── spec.md              # Feature specification (+ Clarifications)
├── research.md          # Phase 0 output — decisions, rationale, alternatives
├── data-model.md        # Phase 1 output — AST + runtime domain model
├── quickstart.md        # Phase 1 output — runnable end-to-end validation scenarios
├── contracts/           # Phase 1 output — external interface contracts
│   ├── cli.md           #   CLI command/flag surface + exit codes
│   ├── grammar.md       #   Runefile grammar (EBNF) — seed for docs/GRAMMAR.md
│   ├── mcp-tools.md     #   task → MCP tool mapping, schemas, annotations, transport/auth
│   └── cache-fingerprint.md  # cache fingerprint format + storage layout
├── checklists/
│   └── requirements.md  # spec quality checklist (16/16)
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root)

```text
cmd/rune/main.go               # entrypoint; cobra global flags + dynamic task dispatcher
internal/
  token/                       # token kinds + Position (byte offset / line / col)
  lexer/                       # Rob Pike state-function lexer
  ast/                         # File, Setting, Assignment, Task, Param, Dep, Attribute, Expr…
  parser/                      # recursive-descent parser; Pratt sub-parser for expressions
  analyzer/                    # var resolution, unknown-dep, cycle detection, arity checks
  diag/                        # diagnostic model + renderer (file:line:col + caret span)
  eval/                        # expression evaluation, {{…}} interpolation, builtin registry
  config/                      # Runefile discovery (walk up tree), settings resolution
  dotenv/                      # .env loading (via mvdan/sh shell package)
  runtime/
    scheduler/                 # dependency DAG, topo-sort, memoization, parallel (errgroup)
    shell/                     # mvdan/sh integration for (sh) bodies
    interp/                    # python/node/custom bodies (temp-file write + exec)
    agent/                     # Provider interface + agent-CLI driver + in-process MCP client
  cache/                       # content-hash fingerprints for [cache] tasks
mcpserver/                     # MCP *server*: expose tasks as tools (public, reusable package)
docs/GRAMMAR.md                # non-normative grammar (generated from contracts/grammar.md)
testdata/                      # integration fixtures (Runefiles + expected stdout/stderr/exit)
.github/workflows/ci.yml       # lint + test on Linux/macOS/Windows; fuzz smoke
.goreleaser.yaml               # multi-platform static binaries + checksums
go.mod / go.sum
```

**Structure Decision**: Single Go module rooted at the repo. `cmd/rune` is the only binary;
all compiler/runtime logic lives in `internal/` small focused packages per Constitution
Principle IV. `mcpserver/` is public (non-`internal`) so it is reusable/embeddable. This mirrors
the Brief §5.1 layout verbatim, which the constitution locks as the package discipline.

## Complexity Tracking

> No Constitution Check violations. No entries.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
