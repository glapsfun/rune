# Phase 1 Data Model — Review Findings

This feature has **no runtime/domain data model** (it changes no Rune behavior and stores no
data). The "model" is the structured shape of the **review** that drives all downstream work
— the `Finding` entity, its lifecycle, and the derived artifacts. This is what makes the
review countable (SC-001), partitionable into US2/US3/US5, and traceable to commits (SC-010).

## Entity: Finding

The atomic unit of the review (US1). Recorded in the review report (`review.md`).

| Field | Type | Rules |
|-------|------|-------|
| `id` | string | Stable, e.g. `F-001`; referenced by remediation commits (SC-010). |
| `skill` | enum | One of: `golang-code-style`, `golang-naming`, `golang-concurrency`, `golang-design-patterns`, `golang-performance`, `golang-documentation`, `golang-pro`. |
| `rule` | string | The specific rule within the skill (e.g. "no init()/mutable globals", "wrap errors with %w", "ctx.Done() in blocking select"). |
| `location` | string | `package/file.go:line` (required, must resolve). |
| `severity` | enum | `S1` correctness/safety · `S2` design · `S3` style/naming/docs · `S4` performance. |
| `recommendation` | string | Concrete fix. |
| `status` | enum | `open → fixed` \| `deferred(reason)` \| `wontfix(justification)`. |
| `commit` | string? | Set when `status=fixed`: the remediating commit SHA (SC-010). |

**Validation rules** (from spec):
- Every Finding MUST have a non-empty `skill`, `rule`, `location`, `severity` (FR-002).
- The report MUST cover every Go package (SC-001) — a package with no findings is recorded
  as "reviewed, clean".
- `S1` Findings MUST reach `status=fixed` (SC-002, FR-005). `S2/S4` may be `fixed` or
  `deferred` with a reason. `wontfix` requires a documented justification (e.g. skill ⟂
  constitution → constitution governs, FR-019).

### State transitions

```text
open ──fix──▶ fixed (records commit)
  │
  ├──defer(reason)──▶ deferred        (allowed for S2/S3/S4; NOT S1)
  └──justify──────▶ wontfix(reason)   (requires documented justification)
```

## Severity → User Story mapping

| Severity | Handled by | Acceptance |
|----------|-----------|------------|
| S1 correctness/safety | **US2** | 100% fixed; `-race` + goleak clean (SC-002/003) |
| S2 design | **US3** | fixed or justified-deferred; suite green (SC-004) |
| S3 style/naming/docs | **US3** + **US4** | fixed; encoded in lint gate where possible (SC-005/007) |
| S4 performance | **US5** | benchmarked; only benchstat-proven fixes merge (SC-008) |

## Derived artifacts (relationships)

- **Review report** (`review.md`) = the ordered set of Findings (US1). Source of truth for
  the work list.
- **Remediation changeset** = commits, each referencing ≥1 Finding `id` (US2/US3/US5);
  closes the `open → fixed` transition and stamps `commit`.
- **Lint gate** (`.golangci.yml`) = the subset of S2/S3 rules that are *mechanically
  enforceable*, encoded as linters (US4) — see `contracts/lint-gate.md`.
- **Benchmark record** = `Benchmark*` functions + a `benchstat` baseline; gates S4 Findings
  (US5).
- **Preservation invariants** = the externally-observable contracts every Finding's fix MUST
  NOT violate — see `contracts/preservation-invariants.md`.

## Representative findings (from the Phase-0 scout — seed the report, not exhaustive)

| Seed | skill / rule | location | sev |
|------|--------------|----------|-----|
| `init()` builds mutable `builtins` global | design-patterns / no init+global | `internal/eval/builtins.go:70,72` | S2 |
| swallowed server error | pro / don't lose errors | `mcpserver/transport.go:75` | S1 |
| silent cache-store discard | pro / surface actionable errors | `internal/cli/run.go:252` | S1 |
| dead `_ = lb/_ = rb` | code-style / no dead assignments | `internal/parser/attribute.go:25-26` | S2 |
| global named `unsafe` shadows stdlib pkg | naming / no predeclared/pkg shadowing | `internal/cache/cache.go:118` | S3 |
| 7× `fmt.Errorf` without `%w` | pro / wrap errors | (audit) | S2 |
| no benchmarks on hot paths | performance / measure first | lexer/parser/eval/scheduler | S4 |

> These are *seeds* surfaced before the full review; US1 produces the complete, authoritative
> set. The full review may reclassify or add to them.
