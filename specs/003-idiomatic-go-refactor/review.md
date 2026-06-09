# Code Review — Idiomatic Go Refactor (US1)

Skill-governed audit of the Rune Go codebase against the bundled `golang-*` skills
(Constitution Principle VIII). Schema and severity legend per
[`contracts/review-rubric.md`](contracts/review-rubric.md):

`F-NNN | skill | rule | loc | sev (S1 correctness > S2 design > S3 style/naming/docs > S4 perf) | status`

Status: `fixed` (this branch) · `deferred(reason)` · `wontfix(justification)` · `reviewed-clean`.

## Findings

| ID | skill / rule | location | sev | status |
|----|--------------|----------|-----|--------|
| F-001 | design-patterns / no `init()`+mutable global | `internal/eval/builtins.go:70,72` | S2 | **fixed** — `sync.OnceValue`; function set identical (eval/diag/corpus goldens unchanged) |
| F-002 | concurrency / goroutine guaranteed exit | `mcpserver/transport.go` `ServeHTTP` | S1 | **fixed** — watcher goroutine now exits via a `done` channel if `Serve` returns first (no leak); return value unchanged |
| F-003 | pro / don't lose errors | `mcpserver/transport.go` `StartHTTP` | S1 | **deferred** — `_ = srv.Serve(ln)` error can't be surfaced through the frozen `stop func()` signature (INV-7); needs an additive API. No leak (goroutine exits when `stop` closes the server) |
| F-004 | pro / surface actionable errors | `internal/cli/run.go:252` | S1 | **fixed** — silent `cache.Store` failure now logged to stderr (Principle I "caching is visible"); caching/exit semantics unchanged |
| F-005 | concurrency / `ctx.Done()` in blocking select | `internal/cli/watch.go` | S1 | **fixed** — `--watch` select ignored the cancelled context (Ctrl-C **hung** under `signal.NotifyContext`); added `ctx.Done()` → clean exit 130 |
| F-006 | code-style / no dead assignments | `internal/parser/attribute.go:25-26` | S2 | **fixed** — removed dead `_ = lb/_ = rb` (tokens were captured then discarded); parse output unchanged (corpus golden) |
| F-007 | naming / no package-name shadowing | `internal/cache/cache.go:118` | S3 | **fixed** — global `unsafe` (shadowed the stdlib package) → `unsafeKeyChars` |
| F-008 | design / context on external calls (FR-008) | `internal/cli/choose.go:45` | S1 | **fixed** — `exec.Command` → `exec.CommandContext(opts.ctx(), …)` |
| F-009 | design / context on external calls (FR-008) | `mcpserver/transport.go` ×2 | S2 | **fixed** — `net.Listen` → `(&net.ListenConfig{}).Listen(ctx, …)` (ServeHTTP uses the request ctx; StartHTTP uses Background) |
| F-010 | concurrency / leak detection | scheduler, mcpserver, cli | S1 | **fixed** — `goleak.VerifyTestMain` added; suite passes leak-clean |
| F-011 | concurrency / accurate cancellation | `internal/runtime/scheduler/parallel.go` | S3 | **deferred** — comment says "first-error cancellation" but `new(errgroup.Group)` waits for all; switching to `WithContext` would change *which* deps run on error (behavior change, out of scope). Doc-only correction recommended |
| F-012 | code-style / redundant loop-var copy | `internal/runtime/scheduler/parallel.go:22` | S3 | **deferred** — `dep := dep` redundant on go1.22+; harmless |
| F-013 | performance / measure first | lexer, parser, eval, scheduler | S4 | **partial** — `Benchmark*` added for lexer + parser (baseline recorded); eval + scheduler benchmarks deferred (need evaluator/DAG harness) |
| F-014 | pro / wrap errors with `%w` | 7 `fmt.Errorf` sites | S2 | **reviewed-clean** — all create *new* errors (e.g. "a bearer token is required"), not wrapping an existing one; `errorlint` reports 0 production issues |
| F-015 | code-style / docs / DI sweep | multiple packages | S2/S3 | **deferred** — broad early-return / field-named-literal / exported-doc pass (US3 T019/T021); lower severity, larger surface |

## Per-package coverage (SC-001)

| Package | Result |
|---------|--------|
| `cmd/rune` | reviewed — clean (build vars are the idiomatic ldflags pattern) |
| `internal/token` | reviewed — `kindNames`/`keywords` are immutable lookup tables (acceptable, A1) |
| `internal/lexer` | reviewed — clean; benchmark added |
| `internal/ast` | reviewed — clean |
| `internal/parser` | F-006 (fixed); benchmark added |
| `internal/analyzer` | reviewed — clean |
| `internal/diag` | reviewed — clean |
| `internal/eval` | F-001 (fixed) |
| `internal/config` | reviewed — `ErrNotFound` sentinel + `candidateNames` immutable (acceptable) |
| `internal/dotenv` | reviewed — clean |
| `internal/cache` | F-007 (fixed); `unsafeKeyChars` regex global immutable (acceptable) |
| `internal/cli` | F-004, F-005, F-008 (fixed); F-015 (deferred: 1857 LOC, file-split candidate) |
| `internal/runtime` | reviewed — clean |
| `internal/runtime/scheduler` | F-010 (fixed); F-011, F-012 (deferred, S3); errgroup bounded ✓ |
| `internal/runtime/shell` | reviewed — clean |
| `internal/runtime/interp` | reviewed — temp-file `defer Close()`/`Remove()` correct |
| `internal/runtime/agent` | reviewed — clean |
| `mcpserver` | F-002, F-009 (fixed); F-003 (deferred, API-frozen); secure-by-default preserved (INV-6) |

## Summary

- **S1 correctness/safety**: 6 found → **5 fixed**, 1 deferred (F-003, API-frozen; no leak, only error-surfacing).
- **S2 design**: 4 found → 3 fixed, 1 broad sweep deferred (F-015).
- **S3 style/naming**: 3 found → 1 fixed (naming), 2 deferred (doc/loop-var, harmless).
- **S4 performance**: benchmarks partial (lexer+parser done; eval/scheduler deferred).
- **Verification**: full suite + `-race` + goleak green in Docker; **no golden regenerated**;
  expanded `golangci-lint` set reports 0 issues.

Deferred items (F-003, F-011, F-012, F-013 eval/scheduler, F-015) are documented for a
follow-up pass; none are S1 leaks or races.
