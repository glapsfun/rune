# Contract — Expanded Lint Gate (US4)

The durable enforcement layer: the additions to `.golangci.yml` that encode the skill rules
mechanically, so the review can't silently rot. Realizes FR-015/FR-016, SC-005.

## Baseline (already present, from feature 002)

Linters: `errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`, `misspell`.
Formatters: `gofmt`, `gofumpt`, `goimports` (local-prefixes set). Existing `errcheck`
exclusions for buffer/stdout writes are retained.

## Additions (this feature)

| Linter | Encodes (skill rule) | Would have caught (scout) |
|--------|----------------------|----------------------------|
| `errorlint` | wrap with `%w`; `errors.Is/As` over `==`/type switch | the 7 `fmt.Errorf` w/o `%w` |
| `contextcheck` | don't drop/replace inherited `context.Context` | context-propagation findings |
| `noctx` | HTTP/external calls carry a context | missing-timeout calls |
| `bodyclose` | response bodies closed (resource lifecycle) | resource-leak findings |
| `predeclared` | no shadowing predeclared/builtin identifiers | global named `unsafe` |
| `wastedassign` | no dead assignments | `_ = lb/_ = rb` |
| `revive` (focused) | naming + exported-doc + early-return | naming/doc findings |
| `gocritic` (stable tags) | idiom/diagnostic style | misc style findings |

### Configuration notes

- `revive`: enable a **curated** rule subset (e.g. `exported`, `early-return`,
  `var-naming`, `context-as-argument`) — **not** its full default — to avoid churn.
- `gocritic`: enable `diagnostic` + `style` stable tags; disable opinionated/experimental
  checks that conflict with `gofumpt`.
- Keep findings actionable: `issues.max-issues-per-linter: 0`, `max-same-issues: 0` (as
  today) so nothing is silently truncated.

## Sequencing & enforcement

1. Add the linters **after** US2/US3 remediation so the tree is already clean (a red gate
   would block the branch — research §2).
2. `golangci-lint run` MUST report **0 issues** on the refactored tree (SC-005).
3. The gate runs in the existing CI `lint` job (002's `ci.yml`, `golangci-lint-action`) — no
   workflow change needed; the config change is picked up automatically. A reintroduced
   violation fails CI and blocks merge (FR-016).

## Acceptance

- `golangci-lint run` → `0 issues` on the refactored code.
- A deliberately-seeded violation (e.g. an unwrapped error, a dead assignment) is flagged by
  the corresponding linter (proves the gate is live).

## Non-goals

- Not enabling every available linter (churn/contradiction) — only those that encode a named
  skill rule (research §2).
- Not changing the CI workflow file — only `.golangci.yml` (+ `go.mod` for the goleak test
  dep, which is unrelated to lint).
