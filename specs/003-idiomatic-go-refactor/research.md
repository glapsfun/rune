# Phase 0 Research — Idiomatic Go Refactor

Decisions for the review methodology, the automated gate, leak detection, benchmarking, and
the concrete remediation approaches surfaced by the code scout. No `NEEDS CLARIFICATION`
remain. Format: **Decision / Rationale / Alternatives considered**.

---

## 1. Review methodology & rubric (US1)

**Decision**: Conduct the review **per package × per skill**, emitting findings in the fixed
schema from `contracts/review-rubric.md`: `{skill, rule, file:line, severity, recommendation,
status}`. Severity scale (drives ordering & US2/US3 split): **S1 correctness/safety**
(concurrency, error loss, resource leak, context) → **S2 design** (init/globals, DI,
options, interface placement) → **S3 style/naming/docs** → **S4 performance**. The report
lives at `specs/003-idiomatic-go-refactor/review.md` (generated during implementation).

**Rationale**: A fixed schema makes findings countable (SC-001), traceable to commits
(SC-010), and directly partitionable into the US2/US3/US5 work. Mapping each finding to a
specific skill rule keeps the audit objective rather than a matter of taste (Principle VIII
intent).

**Alternatives considered**: One free-form review doc — rejected: not countable or
traceable. Per-file instead of per-package×skill — rejected: misses cross-file concerns
(e.g., a goroutine started in one file, joined in another).

---

## 2. Expanded lint gate (US4) — which linters encode which skills

**Decision**: Extend `.golangci.yml` (currently `errcheck, govet, ineffassign, staticcheck,
unused, misspell` + `gofmt/gofumpt/goimports`) with a curated set that mechanically encodes
the skills, introduced **in one PR after remediation** so the tree is already clean:

| Linter | Skill rule it encodes |
|--------|-----------------------|
| `errorlint` | error wrapping with `%w`, correct `errors.Is/As` (golang-pro errors) |
| `contextcheck` | context propagation / non-inherited ctx (golang-design-patterns) |
| `bodyclose` | resource lifecycle — response bodies closed (resource mgmt) |
| `noctx` | external calls carry a context (timeouts/cancellation) |
| `revive` | naming + exported-doc + early-return style (golang-naming/code-style/documentation) |
| `gocritic` | idiom/diagnostic checks (golang-code-style) |
| `predeclared` | shadowing predeclared identifiers (catches the `unsafe` global, golang-naming) |
| `wastedassign` | dead assignments (catches `_ = lb/_ = rb`) |

`revive` is configured with a focused rule set (not its full opinionated default) to avoid
churn; `gocritic` uses the stable diagnostic/style tags only.

**Rationale**: These linters turn the most enforceable skill rules into CI gates so the
review can't silently rot (the durable half of the work). Several would have *caught the
scout's findings* (`predeclared`→`unsafe`, `wastedassign`→dead vars, `errorlint`→`%w` gaps,
`noctx`→missing-context calls), which is the proof they're worth adding.

**Alternatives considered**: Enabling `golangci-lint`'s "all" set — rejected: enormous churn,
many opinionated/contradictory linters, would stall on style noise. Hand-written custom
analyzers — rejected: high cost; the off-the-shelf linters cover the rules. Adding linters
*before* remediation — rejected: a red gate blocks the whole branch; add after the tree is
clean (US4 sequences after US2/US3).

---

## 3. Goroutine-leak detection (US2)

**Decision**: Add `go.uber.org/goleak` as a **test-only** dependency and a
`TestMain(m)` calling `goleak.VerifyTestMain(m)` in the concurrency-bearing packages:
`internal/runtime/scheduler`, `mcpserver`, and `internal/cli` (watch). Use
`goleak.IgnoreTopFunction` only for documented, framework-owned goroutines.

**Rationale**: The constitution SHOULD-guards concurrency tests with goleak. Goroutines live
only in `mcpserver` + scheduler + watch (per the scout), so three `TestMain`s cover the
surface cheaply. This directly verifies "every goroutine has a guaranteed exit" (FR-006).

**Alternatives considered**: Manual goroutine accounting — rejected: brittle. Skipping leak
detection and relying on `-race` — rejected: `-race` finds data races, not leaks; they're
complementary.

---

## 4. Benchmark strategy (US5)

**Decision**: Add `Benchmark*` functions for the pipeline hot paths — `internal/lexer`
(tokenize a representative Runefile), `internal/parser` (parse), `internal/eval` (evaluate
expressions/interpolation), and `internal/runtime/scheduler` (DAG topo + fan-out) — seeded
from existing `testdata` fixtures. Record a baseline with `benchstat` (run via
`go run golang.org/x/perf/cmd/benchstat@latest`). **No optimization merges without a
`benchstat` win**; perf changes ship under `perf(scope):` commits with a comment.

**Rationale**: There are currently **no benchmarks**, so the baseline is net-new and is the
gate for any future perf claim (Principle VIII forbids unprofiled optimization). Benchmarks
on the hot path also document expected complexity.

**Alternatives considered**: Profiling (`pprof`) only — useful but not a regression gate;
benchmarks give a repeatable, CI-checkable baseline. Benchmark everything — rejected:
diminishing returns; the lex→parse→eval→schedule path dominates real workloads.

---

## 5. Removing the `init()` + mutable global in `eval` (US3, headline design fix)

**Decision**: Replace `internal/eval/builtins.go`'s `var builtins map[...]` populated by
`func init()` with a package-level `var builtins = sync.OnceValue(newBuiltins)` (Go 1.21+
`sync.OnceValue`) returning an immutable map built by a pure `newBuiltins()` function — or,
if cleaner, build the map once in the `Evaluator` constructor and hold it on the evaluator.
`IsBuiltin`/`callBuiltin` call the accessor. The **function set and behavior are identical**
(guarded by eval/diag/corpus golden tests).

**Rationale**: Eliminates the `init()` and the mutable package global (Principle VIII), keeps
lazy one-time construction, and is a localized, behavior-preserving change. `sync.OnceValue`
(available on go1.25) is the idiomatic replacement for `init()`-populated singletons.

**Alternatives considered**: Leaving `init()` (it's "just a lookup table") — rejected: the
constitution explicitly forbids `init()`/globals, and SC-006 targets zero. A plain
package-`var` map literal (no init) — viable if the entries are expressible as a literal, but
several entries are closures referencing helpers; a constructor/`OnceValue` is cleaner.

---

## 6. Error-discard audit — fix vs. legitimately-ignored (US2)

**Decision**: Triage the discards the scout found:

- **Fix/surface**: `mcpserver/transport.go:75 _ = srv.Serve(ln)` (server error lost — capture
  and report via the returned shutdown path / log); `internal/cli/run.go:252
  _ = cache.Store(...)` (a cache-write failure should be logged, not silent — consistent with
  Principle I's "caching is visible"); `internal/parser/attribute.go:25-26 _ = lb/_ = rb`
  (dead assignments — investigate whether `lb`/`rb` spans should be used in the AST/diagnostic
  or removed).
- **Keep (legitimate), with a clarifying comment**: `defer func(){ _ = f.Close() }()` /
  `_ = os.Remove(tmp)` / `_ = os.Chmod(...)` cleanup discards (already covered by the
  `errcheck` exclusions), and `_, _ = rand.Read(b[:])` (`crypto/rand.Read` never returns a
  non-nil error in practice) — add a one-line comment rather than churn.

**Rationale**: Not every `_ =` is a defect; the constitution forbids **silent** loss of
*actionable* errors. Surfacing the server/cache errors improves reliability; the cleanup
discards are idiomatic. The `attribute.go` dead vars are the one genuine smell to resolve.

**Alternatives considered**: Mechanically wrapping every discard — rejected: adds noise to
legitimate cleanup paths and fights the existing `errcheck` exclusions.

---

## 7. `internal/cli` decomposition (US3)

**Decision**: `internal/cli` is the largest package (1857 LOC / 11 files). Review it for
single-responsibility and split **files** (not the package) where a file mixes concerns —
keeping the package name `cli` (within the locked layout). Driven by US1 findings; only split
where it improves clarity, not for its own sake (Principle III's "no ceremony" ethos applied
to code).

**Rationale**: Keeps changes within the locked layout (Principle IV) while improving
navigability. File-level reorg is behavior-neutral and low-risk.

**Alternatives considered**: Splitting `cli` into sub-packages — rejected: changes the locked
layout and risks import cycles / exported-surface growth; not worth it for an 1857-LOC
package.

---

## 8. Behavior-preservation strategy (cross-cutting, SC-004/SC-009)

**Decision**: Treat the full existing suite (unit + golden + integration + corpus + fuzz +
the `mcpserver` authz/transport tests + `cli` mcp tests) as the oracle. Acceptance: the suite
passes **without** any golden `-update`, and the `mcpserver`/CLI contract tests pass
unmodified (SC-009). Run via `docker-compose run --rm test go test ./...` and, for races,
`docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...`. Keep each remediation
commit small and finding-scoped (SC-010) so the oracle pinpoints any accidental drift.

**Rationale**: The golden + binary-level + MCP contract tests already assert stdout/stderr/
exit codes and tool schemas, giving a crisp, machine-checkable definition of "behavior
preserved." Small commits keep the blast radius reviewable.

**Alternatives considered**: A big-bang refactor — rejected: behavior drift would be hard to
localize and would violate the small-reviewable-change ethos.
