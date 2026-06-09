# Implementation Plan: Idiomatic Go Refactor — Skill-Governed Review & Refactoring

**Branch**: `003-idiomatic-go-refactor` | **Date**: 2026-06-08 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/003-idiomatic-go-refactor/spec.md`

## Summary

A code-level review and **externally behavior-preserving** refactor of the ~8.7k-line Rune Go
codebase against the bundled `golang-*` skills (Constitution Principle VIII). The work is
review-first: US1 produces a traceable findings report (skill → rule → `file:line` →
severity → fix); US2–US5 remediate by severity. A scout of the code already surfaces the
shape of the work, so this plan is concrete rather than speculative:

- **Concurrency is small and localized** — goroutines exist only in `mcpserver/transport.go`
  (2), with concurrency primitives in `internal/runtime/scheduler/{scheduler,parallel}.go`
  and `internal/cli/watch.go`. So the concurrency review is *deep but narrow*.
- **One `init()` + mutable package global**: `internal/eval/builtins.go` builds the
  `builtins` map in `init()` — the canonical Principle VIII anti-pattern and the headline
  design fix (US3).
- **Constructors are sparse** (only `eval.NewScope`, 2 params) — so functional-options work
  is *minimal*; US3 focuses on the `init()`/global, a `naming` issue (a global literally
  named `unsafe` in `cache.go`, shadowing the stdlib package), and file-responsibility
  tidy in the large `internal/cli` package (1857 LOC / 11 files).
- **Error-handling audit targets**: a swallowed server error
  (`mcpserver/transport.go:75 _ = srv.Serve(ln)`), a silent `cache.Store` discard
  (`cli/run.go:252`), dead `_ = lb/_ = rb` in `parser/attribute.go`, and 7 `fmt.Errorf`
  calls without `%w` (US2 + the `errorlint` gate in US4).
- **Net-new**: there are **no benchmarks** and **goleak is not a dependency** — US5 adds
  hot-path benchmarks and US2 adds leak-guards to the concurrency-bearing tests.

The shipped artifact, Runefile semantics, CLI/MCP behavior, and diagnostics do not change;
the existing suite (with golden files) is the behavior oracle.

## Technical Context

**Language/Version**: Go (go.mod `go 1.25.0`; toolchain `go1.26.2`). Pure Go, `CGO_ENABLED=0`
for the shipped binary; `-race` test runs use CGO.

**Primary Dependencies**: No new *runtime* dependencies. New *test/tooling* deps only:
`go.uber.org/goleak` (leak detection in concurrency tests), additional `golangci-lint` v2
linters (config only, no module change), `golang.org/x/perf/cmd/benchstat` (run via
`go run`, not vendored). Existing concurrency uses `golang.org/x/sync/errgroup` (already
present).

**Storage**: N/A. Artifacts produced: a review report (Markdown), refactor commits,
benchmark files (`Benchmark*` in `*_test.go`), and an expanded `.golangci.yml`.

**Testing**: The **existing** suite is the behavior oracle — table-driven unit tests, golden
files (compare-by-default), binary-level integration tests, the corpus, and lexer/parser
fuzz. New test code: hot-path **benchmarks** and **goleak** `TestMain` guards in
`scheduler`, `mcpserver`, and `cli` (watch). Per global policy the suite runs **inside
Docker** (`docker-compose run --rm test …`); `-race` runs with `-e CGO_ENABLED=1`.

**Target Platform**: unchanged — single static binary, Linux/macOS/Windows × amd64/arm64.

**Project Type**: Single Go module — a CLI tool that is internally a small compiler +
runtime. This feature touches the Go source and `.golangci.yml`; it adds no product surface.

**Performance Goals**: No regression. Benchmarks establish a baseline for the hot paths
(lex → parse → analyze → eval → schedule); any optimization must show a `benchstat` win
(Principle VIII). Performance work is P3 and strictly gated.

**Constraints**: **Externally behavior-preserving** — zero change to Runefile semantics,
CLI contracts, exit codes, diagnostics, or MCP tool behavior; **no golden regenerated**
(SC-004). The **constitution-locked package layout (Principle IV) is preserved** — no locked
package renamed. The public `mcpserver` API and CLI stay backward-compatible (SC-009).

**Scale/Scope**: ~8.7k non-test LOC across 19 packages. Review covers all; remediation is
concentrated in `eval` (builtins), `mcpserver` (transport), `runtime/scheduler`,
`internal/cli` (largest), and the error/naming nits surfaced above.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principle | Gate for this feature | Status |
|---|-----------|------------------------|--------|
| I | Command Runner, Not a Build System | Cache/always-run semantics unchanged; the `cli/run.go` cache-store audit only changes whether a failure is **surfaced**, not caching behavior | ✅ Behavior-preserving |
| II | Errors Are a Feature | Diagnostics/spans unchanged; the `eval` builtins refactor preserves the exact function set + error messages (golden/diag tests guard) | ✅ Preserved |
| III | Minimal, Total DSL | No DSL/grammar change; `parser/attribute.go` dead-var fix must not change parse results (corpus guards) | ✅ Preserved |
| IV | Hand-Written Front End, Idiomatic Go | No parser generator; refactor stays **within** the locked package layout; file splits in `cli` keep package names | ✅ Honored — `contracts/preservation-invariants.md` |
| V | Boringly Portable | Shipped binary stays static/CGO-free; `-race`/goleak are test-only | ✅ Honored |
| VI | Test-First, Multi-Layer Verification | Existing suite is the oracle (no golden regenerated); new benchmark + goleak test code follows the testing discipline; `-race` green | ✅ Advances VI |
| VII | AI-Native, Secure by Default | `mcpserver` refactor preserves secure-by-default (read-only default, env-only secrets, destructive-gated); transport-error fix changes only error *surfacing*, not auth | ✅ Preserved — `contracts/preservation-invariants.md` |
| VIII | Idiomatic Go Engineering Discipline (skill-governed) | This feature **is** the implementation of Principle VIII: review against all `golang-*` skills, remediate, and encode the rules as an expanded lint gate | ✅ Directly implements — `contracts/review-rubric.md`, `lint-gate.md` |

**Result**: **PASS** — no violations; the feature operationalizes Principle VIII while
preserving I–VII. Complexity Tracking is empty.

## Project Structure

### Documentation (this feature)

```text
specs/003-idiomatic-go-refactor/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 — decisions: rubric, linter set, goleak, bench strategy
├── data-model.md        # Phase 1 — Finding entity + lifecycle (the review's data shape)
├── quickstart.md        # Phase 1 — validation scenarios per user story
├── contracts/           # Phase 1 — the "interfaces" this feature defines/must preserve
│   ├── review-rubric.md          # skill → rule → severity taxonomy (US1 output contract)
│   ├── preservation-invariants.md# externally-observable contracts that MUST NOT change
│   └── lint-gate.md              # expanded golangci-lint set + what each linter encodes
├── checklists/
│   └── requirements.md  # spec quality checklist (12/12)
└── tasks.md             # Phase 2 (/speckit-tasks — NOT created here)
```

### Source Code (repository root) — areas this feature touches

```text
# REVIEWED (all Go packages) — US1 covers every package below:
cmd/rune/                      internal/{token,lexer,ast,parser,analyzer,diag,eval,
                               config,dotenv,cache,cli}  internal/runtime/{,scheduler,
                               shell,interp,agent}  mcpserver/

# REMEDIATION HOTSPOTS (from the scout) — US2/US3:
internal/eval/builtins.go      # (~) US3: kill init() + mutable `builtins` global → constructor/OnceValue
internal/cache/cache.go        # (~) US3: rename global `unsafe` (shadows stdlib package)
mcpserver/transport.go         # (~) US2: surface swallowed `srv.Serve` error; goroutine ownership/exit
internal/cli/run.go            # (~) US2: handle/log silent `cache.Store` discard
internal/parser/attribute.go   # (~) US2: resolve dead `_ = lb/_ = rb` (latent issue or remove)
internal/runtime/scheduler/*   # (~) US2: confirm ctx.Done() in every blocking select; bounded
internal/cli/watch.go          # (~) US2: confirm watcher goroutine lifecycle + ctx.Done()
internal/cli/ (11 files)       # (~) US3: file/responsibility tidy WITHIN the package (no rename)
# plus the 7 fmt.Errorf-without-%w call sites (US2/US4)

# NET-NEW — US2/US5:
*_test.go (scheduler, mcpserver, cli)   # (+) US2: goleak TestMain guards
internal/{lexer,parser,eval}/*_test.go, runtime/scheduler  # (+) US5: hot-path Benchmark* funcs

# GATES — US4:
.golangci.yml                  # (~) add errorlint, contextcheck, gocritic/revive, bodyclose, etc.
go.mod                         # (~) add go.uber.org/goleak (test dependency)
```

**Structure Decision**: No package boundaries change (Principle IV locked layout). "Refactor
structure" here means *within-package* file organization (notably splitting the 1857-LOC
`internal/cli` by responsibility), removing the `init()`/global from `eval`, and naming/error
cleanups — all verified by the unchanged test suite. The review (US1) is the spine; every
remediation commit traces back to a finding (SC-010).

## Complexity Tracking

> No Constitution Check violations. No entries.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
