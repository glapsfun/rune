# Implementation Plan: Best-Practices Refactor — Structure, Docs, CI, and Docker

**Branch**: `002-best-practices-refactor` | **Date**: 2026-06-08 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-best-practices-refactor/spec.md`

## Summary

This is a **non-functional hardening** feature: it improves how Rune is documented, verified,
and distributed without changing what Rune does (Runefile semantics, CLI contracts, exit codes,
and diagnostics stay byte-for-byte backward-compatible — enforced by the existing behavioral/
golden/corpus suite as the safety net). Five workstreams, mapped 1:1 to the spec's prioritized
user stories:

1. **Docs (US1, P1)** — add the missing `README.md` and a `docs/` usage set (install,
   getting-started, CLI reference, Runefile language guide, MCP guide, Docker guide) on top of
   the existing `docs/GRAMMAR.md`.
2. **CI gates (US2, P1)** — close the gap between the current pipeline and the constitution:
   add `gofumpt`/`goimports`, run the suite **with `-race`** on all three OSes (CGO enabled for
   the race detector, separate from the CGO-free release build), add a **golden-consistency**
   job and a **cross-platform build** job, and make the Go toolchain a single source of truth
   (`go-version-file: go.mod`).
3. **Docker (US3, P2)** — add a multi-stage **production image** (distroless static, the
   CGO-free binary) plus a `.dockerignore`; keep and document the existing in-container test
   harness.
4. **Structure & refactor (US4, P2)** — add the missing hygiene files (`LICENSE`,
   `CONTRIBUTING.md`, `SECURITY.md`, `.editorconfig`), enforce `gofumpt`/`goimports` in
   `.golangci.yml`, run a behavior-preserving formatting/lint pass, and **dogfood Rune** with a
   root `Runefile` of dev tasks (`lint`, `test`, `fmt`, `build`, `docker`).
5. **Releases (US5, P3)** — add a tag-triggered release workflow wiring the existing
   `.goreleaser.yaml` (binaries + checksums) and publishing the production image to GHCR.

## Technical Context

**Language/Version**: Go — `go.mod` declares `go 1.25.0`; installed toolchain `go1.26.2`;
constitution minimum 1.24+. **Decision**: keep `go 1.25.0` as the language baseline and make CI
derive its version from `go.mod` (`actions/setup-go` `go-version-file: go.mod`) so the
`1.26`-vs-`1.25` drift (spec edge case / FR-014) cannot recur. Pure Go; the **shipped artifact
stays `CGO_ENABLED=0`**.

**Primary Dependencies**: No new Go module dependencies. Tooling/infra only:
`golangci-lint` v2 (config present; add `gofumpt` + `goimports` formatters), `goreleaser`
(config present; wire a workflow), Docker **buildx** + a distroless static base
(`gcr.io/distroless/static-debian12:nonroot`), GitHub Actions, GHCR.

**Storage**: N/A (no runtime data). Artifacts produced: docs (Markdown), CI workflow YAML,
`Dockerfile`, release archives + checksums, a published OCI image.

**Testing**: Reuse the existing suite as the behavior-preservation oracle — table-driven unit
tests, golden files (per-package `-update` flag, **compare-by-default**), binary-level
integration tests, the compatibility corpus, and lexer/parser fuzz targets. New verification is
process-level: race-enabled matrix, golden-consistency guard, build job, a release **dry-run**,
and a docs-example smoke check. Per global policy, the suite runs **inside Docker**
(`docker compose run --rm test ...`).

**Target Platform**: Linux, macOS, Windows on amd64 + arm64 (unchanged). New: a Linux container
image (`linux/amd64`, `linux/arm64`).

**Project Type**: Single Go module — a CLI tool (internally a small compiler + runtime). This
feature touches repo-level tooling/docs, not the compiler/runtime packages (beyond a
formatting/lint pass).

**Performance Goals**: Image size budget **< 30 MB** (SC-005). Time-to-first-task for a new
user **< 10 min** following docs only (SC-001). CI wall-clock kept reasonable via Go build
caching and parallel jobs (advisory, not gated).

**Constraints**: Behavior-preserving — **zero** observable behavior change; no golden file may be
regenerated to accommodate changed behavior (SC-009). The shipped binary stays CGO-free/static
(Principle V) even though the **race-test jobs** enable CGO. No secrets in images, docs, logs,
or release config (Principle VII). Docs are in-repo Markdown only (no hosted site).

**Scale/Scope**: ~6 new docs, ~5 new root files (`README`, `LICENSE`, `CONTRIBUTING`,
`SECURITY`, `Dockerfile`, `.dockerignore`, `.editorconfig`, root `Runefile`), 2 modified config
files (`.github/workflows/ci.yml`, `.golangci.yml`, `.goreleaser.yaml`), 1 new workflow
(`release.yml`), plus a repo-wide `gofumpt`/`goimports` pass.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principle | Gate for this feature | Status |
|---|-----------|------------------------|--------|
| I | Command Runner, Not a Build System | No change to run/cache semantics; dogfooded root `Runefile` uses only normal task features | ✅ N/A — no semantic change |
| II | Errors Are a Feature | No front-end change; diagnostics/spans untouched; behavior-preserving | ✅ N/A |
| III | Minimal, Total DSL | No DSL/grammar change; `docs/GRAMMAR.md` stays the source of grammar truth | ✅ N/A |
| IV | Hand-Written Front End, Idiomatic Go | No parser change; refactor is formatting/lint only, no new parser libs | ✅ Honored |
| V | Boringly Portable (single static binary) | Shipped binary stays `CGO_ENABLED=0`/static; production image is distroless-static; race jobs use CGO **for tests only**, never the release artifact | ✅ Honored — see `contracts/docker-image.md`, `research.md` §3 |
| VI | Test-First, Multi-Layer Verification | This feature **strengthens** VI: `-race` matrix on 3 OSes, golden-consistency guard, fuzz smoke, build job. Refactor is guarded by the existing suite (no new behavior to TDD); any code touch runs the full suite | ✅ Directly advances — `contracts/ci-gates.md` |
| VII | AI-Native, Secure by Default | Docs document secure-by-default MCP (read-only default, env-only secrets, destructive opt-in); CI/release/image contain **no secrets**; image runs as non-root | ✅ Honored — `contracts/docs-structure.md`, `release-pipeline.md` |
| VIII | Idiomatic Go Engineering Discipline (skill-governed) | `.golangci.yml` gains `gofumpt`+`goimports`; CI enforces `golangci-lint run` + `-race`; behavior-preserving cleanup brings code to the bar | ✅ Directly advances — `contracts/ci-gates.md` |

**Workflow & Quality Gates (constitution §"Development Workflow & Quality Gates")**: the new CI
realizes the required gate set verbatim (gofmt/gofumpt/goimports + vet + `golangci-lint` clean;
full suite green **with `-race`** on Linux/macOS/Windows; fuzz build+smoke; golden consistency).
This feature does **not** add or change DSL surface, so the "update `docs/GRAMMAR.md` + golden
fixtures per PR" rule is satisfied vacuously.

**Result**: **PASS** — no violations; the feature implements constitutional requirements.
Complexity Tracking is empty.

## Project Structure

### Documentation (this feature)

```text
specs/002-best-practices-refactor/
├── plan.md              # This file (/speckit-plan output)
├── spec.md              # Feature specification
├── research.md          # Phase 0 — decisions: base image, race+CGO, golden guard, version SoT
├── data-model.md        # Phase 1 — deliverable-artifact inventory + states (no runtime data)
├── quickstart.md        # Phase 1 — runnable validation scenarios (verify each US)
├── contracts/           # Phase 1 — the "interfaces" this feature ships
│   ├── ci-gates.md      #   gate matrix, job topology, pass/fail contract, golden guard
│   ├── docker-image.md  #   production image contract: base, entrypoint, labels, tags, size
│   ├── release-pipeline.md #  tag → binaries+checksums+GHCR image; archive contents
│   └── docs-structure.md   #  documentation IA + required contents per page (US1 coverage)
├── checklists/
│   └── requirements.md  # spec quality checklist (12/12)
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root) — files this feature adds (+) or changes (~)

```text
README.md                      # (+) US1 — what/why, install, quickstart, doc links
LICENSE                        # (+) US4 — resolves .goreleaser file refs (FR-024/FR-027)
CONTRIBUTING.md                # (+) US4 — dev setup, Docker test policy, gate expectations
SECURITY.md                    # (+) US4 — coordinated disclosure; reinforces Principle VII
.editorconfig                  # (+) US4 — cross-editor whitespace/charset baseline
.dockerignore                  # (+) US3 — keep build context minimal & deterministic
Dockerfile                     # (+) US3 — multi-stage: build (CGO=0) → distroless-static-nonroot
Runefile                       # (+) US4 — dogfood: lint/test/fmt/build/docker dev tasks
docs/
  GRAMMAR.md                   # (=) unchanged source of grammar truth
  examples/getting-started/Runefile  # (+) US1 — runnable, freshness-checked doc example
  installation.md              # (+) US1 — binary / source / container install per OS
  getting-started.md           # (+) US1 — zero → first task quickstart (FR-003)
  cli.md                       # (+) US1 — every command/flag with examples (FR-004)
  runefile.md                  # (+) US1 — language usage: deps, params, cache, executors, dotenv
  mcp.md                       # (+) US1 — expose tasks to agents; secure-by-default (FR-005)
  docker.md                    # (+) US1/US3 — run Rune via container; mounts; limits (FR-018/019)
.golangci.yml                  # (~) US2/US4 — add gofumpt + goimports formatters
.goreleaser.yaml               # (~) US5 — ensure LICENSE/README archived; add image hooks if used
.github/workflows/
  ci.yml                       # (~) US2 — gofumpt/goimports, -race matrix (CGO on), golden
                               #          guard, build job, go-version-file: go.mod
  release.yml                  # (+) US5 — on tag: goreleaser + buildx push to GHCR
Dockerfile.test                # (=) keep (improve comments if needed)
docker-compose.yml             # (=) keep — documented test harness
# internal/**, cmd/**, mcpserver/**  → only gofumpt/goimports normalization + lint fixes
#                                       (NO behavior change; golden/integration suite must stay green)
```

**Structure Decision**: The Go package layout already matches the constitution's locked layout
(`cmd/rune`, `internal/{token,lexer,ast,parser,analyzer,diag,eval,config,dotenv,runtime/*,cache,
cli}`, public `mcpserver/`) — confirmed by inspection. So US4 changes **no package boundaries**;
"structure as best practice" here means *repository* structure (hygiene files, dogfooded
`Runefile`) and *code hygiene* (gofumpt/goimports/lint), not re-architecting. All compiler/
runtime edits are mechanical formatting normalizations validated by the unchanged test suite.

## Complexity Tracking

> No Constitution Check violations. No entries.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
