# Implementation Plan: Rich, Example-Driven Documentation & Easy-Start Contributing

**Branch**: `004-rich-documentation` | **Date**: 2026-06-09 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/004-rich-documentation/spec.md`

## Summary

Raise Rune's existing-but-thin documentation to a **rich, example-driven set** a stranger
can understand and act on: a clear overview (the "main idea" + when to use / not use), a
friction-free getting-started path, a **use-case-organized library of runnable examples**
(at least one per headline capability and per common project shape), task-oriented capability
guides, an accurate CLI reference, and a reworked **easy-start CONTRIBUTING**.

The technical approach treats **every example as a tested fixture**, not prose that can rot.
A new Go verification harness (`test/docs/`, modeled on the existing `test/integration`
harness) builds the `rune` binary and asserts that (1) every example Runefile passes static
validation (`rune --file <path> --list` — no built-in `check` subcommand; see Post-Analysis
Refinements), (2) every complete fenced ` ```rune ` block in the docs validates, (3)
shell-only examples run and produce their documented output, and (4) internal doc links
resolve. The repo-root `docs-check` task is rewired to run this harness inside Docker. This
makes "docs never drift" a CI gate (Constitution Principle VI), not a hope. The work changes
**only** documentation, examples, and supporting verification tooling — no product behavior,
DSL, or CLI changes (FR-024). Authoring follows technical-writing best practice
(`technical-writer` skill); the harness follows the bundled `golang-*` skills (Principle VIII).

## Technical Context

**Language/Version**: Documentation in GitHub-Flavored Markdown (CommonMark). Supporting
verification tooling in Go 1.25 (matches `go.mod`). Example task files are Runefiles using
the shipped Rune DSL only.

**Primary Dependencies**: The compiled `rune` binary (built from `./cmd/rune`). The
verification harness uses the Go standard library only (`os/exec`, `testing`, `path/filepath`,
`regexp`) plus the existing in-repo integration-harness pattern — **no new third-party
dependency**, preserving single-binary distribution (Principle V).

**Storage**: Files in the repository — Markdown under `docs/`, runnable examples under
`docs/examples/<use-case>/`, the reworked `README.md` and `CONTRIBUTING.md`. No database.

**Testing**: Go tests run **inside Docker** (`docker-compose run --rm test go test ./...`),
per project policy and the constitution. The new `test/docs` package builds the binary once in
`TestMain` (as `test/integration` does) and runs table-driven, per-example subtests asserting
stdout/stderr/exit code. Examples needing an interpreter (python3/node), a container runtime,
or an agent CLI are still **statically checked always**; their **run** tier is skipped with a
logged reason when the prerequisite is absent (no silent pass).

**Target Platform**: Linux, macOS, and Windows. Shell examples use the default pure-Go
`mvdan.cc/sh` executor so they behave identically across all three (Principle V); any
unavoidable per-OS difference is called out at the point it matters (FR-018).

**Project Type**: Documentation set + a Go verification harness for an existing single-binary
CLI tool. No web/mobile components.

**Performance Goals**: The documentation verification suite completes within the existing CI
budget (target < ~2 min on CI hardware; the static-check tier is near-instant per example).
Reader-facing: time-to-first-task < 5 minutes (SC-002); locate any topic within 2 navigation
steps (SC-006).

**Constraints**: No changes to Rune's runtime behavior, DSL, or CLI surface (FR-024). Examples
must run via the default shell executor wherever possible for cross-platform parity; examples
requiring external tooling must declare prerequisites up front (FR-008) and be CI-skippable.
Secrets never appear in any example or doc (Principle VII). Verification must be repeatable and
wired into `docs-check` + CI.

**Scale/Scope**: ~9 headline capabilities and ~8 common project shapes → **≥16 runnable
examples** (the minimum coverage bar, FR-007/SC-004); ~10–12 documentation pages (overview,
getting-started, installation, CLI reference, language guide, ~9 capability guides,
troubleshooting, examples index); 1 verification harness package; full `README.md` +
`CONTRIBUTING.md` rework.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The Rune Constitution v1.1.0 (8 principles). This is a documentation + test-tooling feature;
it touches no product code paths. Evaluation:

| Principle | Applies how | Status |
|-----------|-------------|--------|
| **I. Command Runner, Not a Build System** | Docs/examples MUST teach always-run semantics; caching examples MUST show the explicit `[cache(...)]` opt-in and the visible "cached" log, never imply timestamp skipping. | ✅ Pass — encoded in the caching guide/example contract. |
| **II. Errors Are a Feature** | The troubleshooting/guides content MUST show real `file:line:col` + caret diagnostics; error-behavior claims are verified against actual output. | ✅ Pass — failure-mode docs (FR-011) assert real diagnostics. |
| **III. Minimal, Total DSL** | Examples MUST use only shipped DSL surface — no invented syntax, no implied loops/recursion. | ✅ Pass — every example Runefile is `rune --list`-validated. |
| **IV. Hand-Written Front End, Idiomatic Go** | The verification harness is idiomatic Go in a focused `test/docs` package. | ✅ Pass — mirrors `test/integration` layout. |
| **V. Boringly Portable** | Shell examples use the pure-Go `sh` executor for cross-OS parity; no new runtime dependency; harness is stdlib-only. | ✅ Pass — single-binary distribution untouched. |
| **VI. Test-First, Multi-Layer Verification (NON-NEGOTIABLE)** | Examples become **tested fixtures**: static-check + run-and-assert tiers, plus link checking, all in CI. Harness written test-first (expectations before examples added). | ✅ Pass — this principle actively shapes the design; doc drift becomes a merge blocker. |
| **VII. AI-Native, Secure by Default** | The agents/MCP guide MUST state read-only default, env-only secrets, gated destructive tasks; **no secret literal appears in any example/doc**. | ✅ Pass — example contract forbids secrets; verified by scan. |
| **VIII. Idiomatic Go Engineering Discipline (Skill-Governed)** | Harness Go code MUST be `gofumpt`/`golangci-lint`-clean, errors wrapped with `%w`, no goroutine leaks, table-driven tests. | ✅ Pass — authored under `golang-pro`/`golang-cli`; subject to the same CI gates. |

**Result: PASS — no violations.** Complexity Tracking is therefore empty. The feature is
strongly *aligned* with Principle VI: turning examples into verified fixtures is the mechanism
that keeps documentation trustworthy, exactly as the constitution demands for a tool other
teams' CI depends on.

## Project Structure

### Documentation (this feature)

```text
specs/004-rich-documentation/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output — decisions (IA, harness, conventions)
├── data-model.md        # Phase 1 output — the documentation content model
├── quickstart.md        # Phase 1 output — validation scenarios for the docs set
├── contracts/           # Phase 1 output — durable contracts
│   ├── example-contract.md           # required shape/metadata of every example
│   ├── doc-verification-harness.md   # what test/docs verifies & how it's run
│   ├── information-architecture.md   # page map, per-page structure, navigation, glossary
│   └── cli-reference.md              # authoritative CLI surface the reference must mirror
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root)

```text
docs/
├── overview.md            # NEW — the "main idea": problem, mental model, when to use / not
├── getting-started.md     # REWORK — single linear install→write→run, expected output shown
├── installation.md        # REFINE — per-OS, cross-platform call-outs
├── cli.md                 # REWORK — complete reference: flags, mcp/serve/completion, exit codes
├── runefile.md            # REFINE — language guide; cross-link to guides + examples
├── troubleshooting.md     # NEW — common failure modes + the exact diagnostic (FR-011)
├── GRAMMAR.md             # KEEP — formal grammar
├── guides/                # NEW — task-oriented capability guides (concept+syntax+example+pitfalls)
│   ├── dependencies-and-hooks.md
│   ├── parameters.md
│   ├── caching.md
│   ├── parallelism.md
│   ├── executors.md       # multi-language bodies (sh/python/node/agent)
│   ├── settings-and-dotenv.md
│   ├── imports-and-modules.md
│   ├── os-filtering.md
│   └── agents-and-mcp.md  # supersedes/expands docs/mcp.md; security model front-and-center
└── examples/              # NEW library — use-case-organized; each dir = Runefile + README
    ├── README.md          # the index: groups labeled by use case (entry to SC-006)
    ├── getting-started/   # EXISTING — keep; add README to meet example contract
    ├── go-service/        # project shape: compiled-language service
    ├── node-project/      # project shape: Node/JavaScript
    ├── python-project/    # project shape: Python
    ├── monorepo/          # project shape: imports/namespaced modules
    ├── ci-cd/             # project shape: CI pipeline (list/dry-run/exit codes)
    ├── docker-workflow/   # project shape: containerized workflow
    ├── polyglot/          # project shape: mixed shell/python/node bodies
    ├── caching/           # capability spotlight: [cache(...)] opt-in + "cached" log
    ├── parallel/          # capability spotlight: parallel prerequisites
    └── agent-driven/      # capability spotlight: agent task + MCP exposure

test/
└── docs/                  # NEW — documentation verification harness (Go, stdlib-only)
    ├── harness_test.go    # TestMain builds the binary once (mirrors test/integration)
    ├── examples_test.go   # per-example subtests: `rune --list` validate always; run+assert when able
    ├── codeblocks_test.go # extract complete ```rune blocks from docs/*.md → `rune --list` validate
    └── links_test.go      # internal Markdown links resolve to existing targets

README.md                  # REWORK — entry point; links to overview + examples; consistent terms
CONTRIBUTING.md            # REWORK — easy-start: what to contribute, clone→verify, repo map, gates
Runefile                   # UPDATE — `docs-check` runs the new harness inside Docker
.github/workflows/ci.yml   # UPDATE — ensure docs verification runs as a gate
```

**Structure Decision**: Keep the established conventions — prose under `docs/`, runnable
examples under `docs/examples/<use-case>/` (the existing `getting-started` example already
lives there, and `docs-check` already points into it). The verification harness is a new Go
test package `test/docs/` that mirrors the proven `test/integration` harness (build-once in
`TestMain`, run the binary, assert stdout/stderr/exit). No new top-level directories or
third-party dependencies are introduced. Each example directory is self-contained: a
`Runefile` plus a `README.md` carrying the metadata the example contract requires (purpose,
prerequisites, capability demonstrated, run command, expected output, guide cross-link).

## Complexity Tracking

> No Constitution Check violations — this table is intentionally empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| _(none)_  | —          | —                                    |

## Post-Analysis Refinements (golang-grounded)

`/speckit-analyze` surfaced six findings; the fixes below were applied to the design artifacts
and are reflected in `tasks.md`. The central one (F1) is a Go-harness correctness fix, grounded
in the `golang-pro` / `golang-cli` skills (Principle VIII). No new Constitution issues; gate
re-checked → still PASS.

| # | Finding | Resolution (where) |
|---|---------|--------------------|
| **F1** (HIGH) | The harness assumed a built-in `rune check`; none exists (`cmd/rune` special-cases only `mcp`/`serve`/`completion`, so `check` is a *task name* — today's `docs-check` works only because the getting-started Runefile defines a `check` task). | **Static validation now uses `rune --file <path> --list`** — parse+analyze, runs nothing, needs no task arg (`internal/cli/run.go` analyzes before the `--list` branch); exit `0`=valid, `3`=validation, `2`=usage. Updated research §D2, `contracts/doc-verification-harness.md`, `contracts/example-contract.md` (E1), `contracts/cli-reference.md`, tasks T004/T005, and the Notes. |
| **F1-adjacent** | Not every fenced ` ```rune ` block is a complete Runefile (some are expression/partial snippets) — `--list` would reject them. | **Fenced-block convention**: only *complete* ` ```rune ` blocks are validated; deliberate fragments use ` ```text `. Codified in §D2, the harness contract, and T005. |
| **I1** (MED) | The glossary forbade "prerequisite" as an alias of "dependency", but the example contract uses **"Prerequisites:"** for required external tooling — the alias check (T007c) would false-positive on every example. | Scoped the rule: `prerequisite` is forbidden **only for task ordering**; the `Prerequisites:` tooling field is allowed and excluded from the check. Updated `contracts/information-architecture.md`. |
| **I2** (MED) | T038 supersedes `docs/mcp.md` with `docs/guides/agents-and-mcp.md`; the `README.md` doc-table link (and others) to `docs/mcp.md` would break the links check / strand readers. | T038 now requires deciding remove-vs-redirect and updating all `docs/mcp.md` references. |
| **C1** (LOW) | FR-024 ("docs/tooling only, no product change") had no positive verification task. | Added **T052** — a diff-scope guard asserting only `docs/`, `test/docs/`, `Runefile`, `README.md`, `CONTRIBUTING.md`, `.github/workflows/ci.yml` change. |
| **N1** (LOW) | T009 implied Docker on all CI OS legs, but Docker is Linux-only; cross-OS runs are native (as `test/integration` already does). | T009 now specifies **native `go test ./test/docs` per OS** in CI; Docker stays the local flow. |

**Idiomatic-Go harness decisions** (from `golang-pro` testing reference, now in
`contracts/doc-verification-harness.md`): table-driven `t.Run` subtests (one per example, range
var captured) so a failure names the offending example; `exec.CommandContext` with a bounded
timeout and `exec.ExitError` inspection (never `os.Exit` in tests); `t.Skip(reason)` via
`exec.LookPath` for missing interpreters (logged, never silent); `t.Helper()` on helpers;
`%w`-wrapped errors; `-update`-gated goldens only for output-centric examples; and the package
held to `gofumpt`/`golangci-lint`-clean + `go test -race` (T049).
