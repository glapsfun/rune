# Phase 0 Research — Best-Practices Refactor

All Technical Context unknowns are resolved below. No `NEEDS CLARIFICATION` remain. Each entry:
**Decision / Rationale / Alternatives considered**, grounded in the repo's actual state and the
constitution.

---

## 1. Go toolchain version — single source of truth

**Decision**: Keep `go.mod` at `go 1.25.0` (the language baseline) and have CI resolve its
toolchain via `actions/setup-go` with `go-version-file: go.mod` instead of a hard-coded
`go-version: "1.26"`. Document the supported floor (1.24+) in `CONTRIBUTING.md`.

**Rationale**: The spec flags the `1.26` (CI) vs `1.25.0` (`go.mod`) drift (FR-014, edge case).
A hard-coded CI version inevitably drifts from `go.mod`; deriving from `go.mod` makes the module
file the one place the version lives. `go 1.25.0` already satisfies the constitution's 1.24+
minimum, and the installed `go1.26.2` toolchain builds it fine.

**Alternatives considered**: (a) Bump `go.mod` to `go 1.26` — rejected: needlessly raises the
floor with no feature need, and the constitution targets broad compatibility. (b) Pin CI to a
concrete `1.26.x` — rejected: reintroduces a second version to keep in sync.

---

## 2. CI job topology & the constitutional gate set

**Decision**: Restructure `.github/workflows/ci.yml` into parallel jobs:

- **lint** (ubuntu): `gofmt -l` empty, `gofumpt -l` empty, `goimports -l` empty (run via
  `golangci-lint` formatters where possible), `go vet ./...`, `golangci-lint run`.
- **test** (matrix: ubuntu/macos/windows): `go test -race ./...` with **`CGO_ENABLED=1`**, plus
  `setup-node`/`setup-python` (already present) for multi-language executor integration tests.
- **build** (matrix: ubuntu/macos/windows): `go build ./...` with **`CGO_ENABLED=0`** — proves
  the shipped static artifact compiles everywhere.
- **golden** (ubuntu): regenerate goldens with `-update` for the golden-bearing packages, then
  `git diff --exit-code` (see §4).
- **fuzz-smoke** (ubuntu): unchanged — `FuzzLexer`/`FuzzParser` for ~20s each.
- **release-dryrun** (ubuntu): `goreleaser release --snapshot --clean` to prove config + file
  refs resolve before any real tag (supports FR-027/SC-006).

**Rationale**: Mirrors the constitution's required gate set verbatim (gofmt/gofumpt/goimports +
vet + golangci-lint clean; full suite **with `-race`** on all three OSes; fuzz build+smoke;
golden consistency; cross-platform build). Splitting build (CGO off) from test (CGO on) keeps
the two CGO modes honest — see §3. Lint is OS-independent so one runner suffices.

**Alternatives considered**: A single mega-job — rejected: loses parallelism and makes failure
attribution muddy (SC-003 wants the failing gate named). Running `-race` only on Linux —
rejected: constitution explicitly requires Linux/macOS/Windows.

---

## 3. Race detector (CGO) vs. CGO-free static binary

**Decision**: Race-enabled test jobs set `CGO_ENABLED=1`; the build/release jobs and the
production `Dockerfile` set `CGO_ENABLED=0`. The two never mix in one invocation. GitHub's
`ubuntu-latest` (gcc), `macos-latest` (clang), and `windows-latest` (mingw gcc on PATH) all ship
a C toolchain, so `go test -race` runs on all three without extra setup.

**Rationale**: The race detector requires CGO + a C compiler; the shipped single binary must be
CGO-free/static (Principle V). These are different invocations of the same code, so there is no
conflict as long as they stay in separate jobs. This satisfies the spec edge case directly and
keeps the release artifact pure-Go.

**Alternatives considered**: Skipping `-race` on Windows to avoid the C-toolchain question —
rejected: the runner already has mingw, and the constitution requires `-race` on Windows. If a
future runner regresses, the fallback is to drop Windows `-race` **explicitly with a logged
note** (never silently), per the spec edge case.

---

## 4. Golden-file consistency guard (FR-012)

**Decision**: Goldens already **compare-by-default** (a drifted golden fails `go test`), so the
normal `test` job catches most drift. Add a dedicated **golden** job that *regenerates* and
diffs, scoped to the packages that define `-update`:

```sh
go test ./internal/diag ./internal/lexer ./internal/parser ./internal/cli ./test/corpus -update
git diff --exit-code testdata/ docs/GRAMMAR.md
```

A non-empty diff fails with guidance to regenerate deliberately and update `docs/GRAMMAR.md` if
the grammar changed.

**Rationale**: `go test ./... -update` would **fail** on packages that don't declare the
`-update` flag (`flag provided but not defined`), so the regen command must target the five
golden-bearing packages explicitly (confirmed by grep: `internal/diag`, `internal/lexer`,
`internal/parser`, `internal/cli`, `test/corpus`). This enforces the constitution's "golden
files regenerated deliberately, never hand-edited" rule and gives a deterministic "regenerate X"
failure (spec edge case). It is belt-and-suspenders over the compare-by-default behavior.

**Alternatives considered**: Relying solely on compare-by-default — rejected: it catches a
drifted golden but not a *hand-edited* golden that happens to match wrong output; the
regenerate-and-diff guard closes that hole. A bespoke regen script — rejected: the `-update`
flag is the project's existing, documented mechanism.

---

## 5. `.golangci.yml` formatter additions

**Decision**: Under `formatters.enable`, add `gofumpt` and `goimports` alongside the existing
`gofmt` (golangci-lint v2 runs formatters and can fail on unformatted files). Keep the current
linter set (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`, `misspell`) and the
existing `errcheck` exclusions. Configure `goimports.local-prefixes:
github.com/rune-task-runner/rune` so local imports group last.

**Rationale**: Principle VIII requires `gofumpt`/`goimports`-clean code; the current config only
enables `gofmt`. Adding them to the single `.golangci.yml` keeps "format" defined in one place
(used by both CI and the dogfooded `Runefile fmt` task). A one-time repo-wide `gofumpt -w` +
`goimports -w` pass brings existing code to the bar (behavior-preserving — formatting only,
guarded by the test suite).

**Alternatives considered**: Separate `gofumpt`/`goimports` CLI steps in CI outside
golangci-lint — rejected: duplicates configuration and version pinning; golangci-lint v2 already
orchestrates formatters.

---

## 6. Production container image — base & shape

**Decision**: Multi-stage `Dockerfile`. Stage 1 (`golang:1.25-bookworm` or matching `go.mod`)
builds the binary with `CGO_ENABLED=0 -trimpath -ldflags "-s -w -X main.version=… -X
main.commit=…"`. Final stage = `gcr.io/distroless/static-debian12:nonroot` containing only the
`rune` binary; `ENTRYPOINT ["/rune"]`, `WORKDIR /work`, runs as the nonroot user. Build for
`linux/amd64,linux/arm64` via buildx. Add OCI labels (source, version, license).

**Rationale**: The binary is CGO-free/static, so distroless-static (no libc, no shell) is the
smallest secure base that still ships CA certificates and `/etc/passwd` for the nonroot user.
Total image ≈ base (~2–3 MB) + binary (~10–15 MB) → **well under the 30 MB budget** (SC-005).
Non-root + no shell shrinks attack surface (aligns with Principle VII's secure-by-default
posture). The default `sh` executor is pure-Go (`mvdan/sh`), so **no system shell is needed** for
self-contained tasks.

**Alternatives considered**: `scratch` — rejected: lacks CA certs and a passwd entry, complicating
HTTPS (remote MCP/agent providers) and nonroot. `alpine` — rejected: adds a libc/shell and ~5 MB
for capabilities this minimal image intentionally omits; documented as the "fuller image" users
can build themselves if they need system tools. **Known limitation** (FR-019, documented in
`docs/docker.md`): tasks invoking external programs or the `python`/`node`/agent executors won't
find those runtimes in the minimal image — they must bind-mount a richer environment or install
natively. The pure-Go shell still runs, and missing-runtime failures surface as clear errors.

---

## 7. Release pipeline (US5) & image publishing

**Decision**: Add `.github/workflows/release.yml` triggered on `v*` tags: (1) run `goreleaser
release --clean` (uses the existing config → cross-platform archives + `checksums.txt`,
GitHub-sourced changelog); (2) `docker buildx build --push` the production image to
`ghcr.io/rune-task-runner/rune` tagged with the version + `latest`, using the repo's
`GITHUB_TOKEN` for GHCR auth. Ensure `.goreleaser.yaml` archives include `LICENSE` + `README`
(refs now resolve once those files exist). `permissions: { contents: write, packages: write }`.

**Rationale**: `.goreleaser.yaml` already defines builds/archives/checksums but nothing triggers
it and no image is published — this wires both with one tag action (SC-007). GHCR + `GITHUB_TOKEN`
needs no extra secrets (Principle VII: no secret material in config). A `release --snapshot`
dry-run in CI (job in §2) catches unresolved file refs before a real tag (SC-006).

**Alternatives considered**: GoReleaser's built-in `dockers:`/`kos:` image build — viable, but a
separate `buildx` step gives clearer multi-arch control and keeps image concerns out of the
binary-release config; revisit if duplication grows. Publishing to Docker Hub — rejected: adds a
secret and a second registry; GHCR is native to the repo.

---

## 8. Documentation information architecture (US1)

**Decision**: `README.md` is the entry point (what/why, 60-second pitch vs `make`/`just`,
install one-liner, a runnable quickstart snippet, and links). Deep content lives under `docs/`:
`installation.md`, `getting-started.md`, `cli.md`, `runefile.md`, `mcp.md`, `docker.md`, with
`GRAMMAR.md` remaining the formal grammar. Source the CLI surface from `cmd/rune/main.go` flags
and feature-001 `contracts/cli.md`; source the language guide from `contracts/grammar.md` /
`docs/GRAMMAR.md`. Every CLI command/flag and every language construct gets ≥1 runnable example
(SC-010). A docs-example smoke check runs the getting-started Runefile in CI (or via the
dogfooded `Runefile docs-check` task) to catch staleness (FR-006, edge case).

**Rationale**: Splits a scannable landing page from reference depth — the standard, low-friction
docs IA for a CLI tool; keeps `GRAMMAR.md` authoritative and avoids duplicating the grammar.
Examples are checked by execution so docs can't silently rot (FR-006).

**Alternatives considered**: A hosted docs site (mkdocs/Docusaurus) — explicitly out of scope
this iteration (spec assumption); in-repo Markdown renders on GitHub with zero infra. One giant
`README` — rejected: poor scannability and fails the "find any command/construct" test (SC-010).

---

## 9. Dogfooding — root `Runefile`

**Decision**: Add a repo-root `Runefile` defining the project's own dev workflow: `fmt`
(gofumpt+goimports via golangci-lint), `lint` (`golangci-lint run`), `test` (the Docker harness:
`docker compose run --rm test`), `test-race`, `build`, `docker` (buildx the image), `docs-check`,
and `release-dryrun`. Document it in `CONTRIBUTING.md`.

**Rationale**: A task runner should run its own tasks — the strongest "best practice" signal and
a living usage example for US1. It uses only existing, stable Runefile features (no DSL change,
Principle I/III honored). It also de-duplicates command definitions shared by humans, CI, and
docs.

**Alternatives considered**: A `Makefile` — rejected: the project *is* a task runner;
dogfooding is both better practice here and free documentation. Scripting everything inline in
CI — rejected: duplicates commands across CI, docs, and local dev.

---

## 10. Behavior-preservation strategy (cross-cutting, SC-009)

**Decision**: Treat the entire existing suite (unit + golden + integration + corpus + fuzz) as
the refactor oracle. The only code edits are `gofumpt`/`goimports` normalization and any
`golangci-lint` fixes that are provably semantics-neutral. Acceptance: the suite passes
**without** running any golden `-update`. Per global policy, run it via
`docker compose run --rm test`.

**Rationale**: The backward-compatibility promise forbids non-opt-in behavior change; the golden
+ binary-level integration tests already assert stdout/stderr/exit codes, so "no golden
regenerated" is a crisp, machine-checkable definition of "behavior preserved" (spec edge case).

**Alternatives considered**: A broader logic refactor — rejected: out of scope and risks
behavior change; performance work is separately gated by Principle VIII (needs benchmarks) and
is explicitly out of scope.
