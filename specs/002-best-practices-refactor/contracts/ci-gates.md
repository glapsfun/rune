# Contract — CI Quality Gates

The CI pipeline's "interface" to contributors: the set of gates, their pass/fail semantics, and
the merge contract. Realizes spec FR-007…FR-015, SC-002/003/004, and the constitution's
"Development Workflow & Quality Gates". Implemented in `.github/workflows/ci.yml`.

## Triggers

- `push` to `main`
- `pull_request` (any branch → blocks merge until green)

`permissions: { contents: read }` for CI (release escalates separately — see
`release-pipeline.md`).

## Toolchain source of truth

`actions/setup-go@v5` with `go-version-file: go.mod` (NOT a hard-coded version). Resolves the
`1.26`-vs-`1.25` drift (FR-014); the module file is the only place the version is declared.

## Job topology & gate contract

| Job | Runner(s) | CGO | Command(s) | Fails when |
|-----|-----------|-----|------------|-----------|
| **lint** | ubuntu | n/a | `gofmt -l .` empty; `golangci-lint run` (incl. `gofumpt`,`goimports` formatters); `go vet ./...` | any unformatted file, lint finding, or vet error |
| **test** | ubuntu, macos, windows | **1** | `go test -race ./...` (with node20 + python3.12 for executor tests) | any test failure or **data race** |
| **build** | ubuntu, macos, windows | **0** | `go build ./...` | the static binary fails to compile on any OS |
| **golden** | ubuntu | 0 | regen + `git diff --exit-code` (see below) | committed goldens differ from freshly generated |
| **fuzz-smoke** | ubuntu | 0 | `go test -run=xxx -fuzz=FuzzLexer -fuzztime=20s ./internal/lexer`; same for `FuzzParser` ./internal/parser | a fuzz target fails to build or crashes |
| **release-dryrun** | ubuntu | 0 | `goreleaser release --snapshot --clean` | config invalid or a referenced file (LICENSE/README) is missing |

**Merge contract**: every job above is a required status check. A red job blocks merge; the job
name identifies the failing gate (SC-003). `fail-fast: false` on the test/build matrices so one
OS failing still reports the others.

## Golden-consistency guard (FR-012)

Goldens compare-by-default (drift already fails `test`). The `golden` job additionally proves
the committed goldens equal a deliberate regeneration — scoped to the packages that declare the
`-update` flag (a blanket `go test ./... -update` errors on packages without it):

```sh
go test ./internal/diag ./internal/lexer ./internal/parser ./internal/cli ./test/corpus -update
git diff --exit-code testdata/ docs/GRAMMAR.md
```

Non-empty diff → fail with: *"goldens out of date — regenerate deliberately with `-update` and,
if the grammar changed, update docs/GRAMMAR.md (never hand-edit goldens)."*

## Race detector / CGO rule

`test` runs with `CGO_ENABLED=1` (race detector needs a C toolchain — present on all three
GitHub runners). `build` and all release/image builds run with `CGO_ENABLED=0` so the **shipped
artifact stays static** (Principle V). The two modes never share an invocation. If a platform
ever cannot run `-race`, that exception MUST be encoded explicitly with a logged comment — never
silently dropped (spec edge case).

## Non-goals (this contract)

- Code-coverage thresholds (may be reported, not gated).
- Performance/benchmark gates (Principle VIII governs perf separately; out of scope).
- Changing what any test asserts (behavior-preserving — see `release-pipeline.md` / SC-009).
