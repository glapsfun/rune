# Contract — Review Rubric (US1 output)

The review's "interface": the fixed taxonomy and finding schema so the audit is objective,
countable (SC-001), and traceable (SC-010). Realizes FR-001…FR-004.

## Finding schema (every finding)

```text
F-NNN | skill=<skill> | rule="<rule>" | loc=<pkg/file.go:line> | sev=S1..S4
      | rec="<concrete fix>" | status=open|fixed|deferred|wontfix [| commit=<sha>]
```

All fields required except `commit` (set on fix). See `../data-model.md` for validation +
lifecycle.

## Skills → rule families (the audit checklist)

| Skill | Rule families the review MUST check |
|-------|-------------------------------------|
| `golang-concurrency` | goroutine has clear owner + guaranteed exit; `ctx.Done()` in every blocking `select`; only sender closes a channel; channel params directional; concurrency bounded (`errgroup.SetLimit`); no data races |
| `golang-pro` (errors) | every error handled; wrap with `%w`; `errors.Is/As` not `==`/type-assert; no silent `_` discards of actionable errors; `panic` only for impossible states |
| `golang-design-patterns` | functional options for wide/optional constructors; no `init()`/mutable globals; `defer Close()` after open; timeout/ctx on external calls; accept interfaces, inject deps; bounded pools/queues |
| `golang-code-style` | early-return / no `else` after terminal branch; ≤4 params (else options struct); `context.Context` as the first parameter; explicit slice/map init; field-named composite literals; no dead assignments |
| `golang-naming` | idiomatic names; no shadowing predeclared identifiers or package names (e.g. `unsafe`); minimized/unexported surface |
| `golang-documentation` | exported identifiers carry doc comments; package docs present |
| `golang-performance` | (review only; changes gated) identify hot paths; no unprofiled optimization |

## Severity scale (drives ordering + story split)

| Sev | Meaning | Owner story |
|-----|---------|-------------|
| **S1** | correctness/safety — concurrency, lost errors, resource/goroutine leaks, missing context | US2 (must fix) |
| **S2** | design — init/globals, DI, options, interface placement, dead code | US3 |
| **S3** | style / naming / documentation | US3 + US4 |
| **S4** | performance | US5 (benchmark-gated) |

## Coverage rule

Every Go package under `cmd/`, `internal/**`, `mcpserver/` MUST appear in the report — either
with findings or recorded "reviewed, clean" (SC-001). The report is ordered by severity
(FR-003).

## Non-goals

- Not a style-bikeshed: only rules traceable to a named skill are findings.
- Not behavior change: a "finding" never proposes altering observable behavior (see
  `preservation-invariants.md`).
