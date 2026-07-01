# Implementation Plan: Modern, Example-Rich Documentation & README Status Badges

**Branch**: `009-docs-and-badges` | **Date**: 2026-07-01 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/009-docs-and-badges/spec.md`

## Summary

Turn Rune's existing documentation into a coherent, modern handbook read on GitHub and give
the repo a credible front door. Three tracks: (1) **README** — a light header refresh
(centered title/tagline, quick-nav) topped with a live status-badge row (CI, release/tag,
license, Go version, Go Report Card, docs, Go Reference); (2) **docs reorganization in
place** — fold `docs/guides/` into a Diátaxis-shaped layout (`how-to/`, `user-guide/`,
`use-cases/`) with the flat reference pages kept and all internal links updated, fronted by a
goal-oriented ("I want to…") index, adding use-case walkthroughs for **Python**, **Node**, and
**MCP/AI-agent** projects anchored to the existing runnable examples; (3) **verification** —
extend the already-present `test/docs` harness (CI gate `docs-verify`, run via `rune
docs-check`) to cover the reorganized pages and add badge-integrity checks, so accuracy is
enforced on every change.

This is a documentation-and-README feature. No Rune engine, CLI, or DSL code changes; the
shipped binary and golden/byte-exact CLI output are untouched.

## Technical Context

**Language/Version**: GitHub-Flavored Markdown for all docs + `README.md`. The verification
harness is Go 1.25 (existing `test/docs` package); no engine/CLI Go code changes.

**Primary Dependencies**: No new runtime dependencies. README-render-time only: shields.io
badge images, goreportcard.com, and pkg.go.dev — external image/link providers, never linked
into the binary. Verification reuses the existing `test/docs` harness and the
`docker-compose` test env.

**Storage**: N/A (documentation artifacts only).

**Testing**: Existing `test/docs` Go harness — `rune docs-check` →
`docker-compose run --rm test go test ./test/docs/...` (Docker-only per project policy).
Extended with: badge-integrity test, an internal-link pass over the reorganized tree, and a
growing `selfContainedPages` allowlist for fenced `rune` blocks.

**Target Platform**: GitHub markdown rendering — light & dark themes, desktop & mobile widths.

**Project Type**: Single project (CLI + in-repo docs). No hosted documentation site.

**Performance Goals**: N/A. The extended `docs-verify` gate must stay within the existing CI
time budget (static Tier-A checks run on every OS; interpreter-dependent Tier-B checks skip
when a runtime is absent).

**Constraints**: No hosted site; no new runtime deps; forward-slash paths in all docs
(Principle V); examples stay cross-platform on the pure-Go shell; **zero change to CLI
behavior or golden output** (SC-008); badges must degrade gracefully and target the canonical
repo (`glapsfun/rune`) / module (`rune-task-runner/rune`).

**Scale/Scope**: ~20 existing doc pages reorganized in place; 3 new use-case walkthroughs
(Python, Node, MCP); 1 goal-oriented index; 1 README header + badge refresh; harness
extension (badge-integrity + allowlist growth). Builds on the existing 15-example library.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment | Status |
|-----------|------------|--------|
| I. Command Runner, Not a Build System | No behavior change; docs only. | ✅ Pass |
| II. Errors Are a Feature | Diagnostics unchanged; docs may illustrate them accurately. | ✅ Pass |
| III. Minimal, Total DSL | DSL surface untouched. | ✅ Pass |
| IV. Hand-Written Front End, Idiomatic Go | No engine/parser changes; `test/docs` package already exists and stays put. | ✅ Pass |
| V. Boringly Portable | Forward-slash paths; no runtime deps added; badges are render-time external images, not binary deps; examples stay cross-platform. | ✅ Pass |
| VI. Test-First, Multi-Layer Verification | **Directly reinforced** — docs are tested fixtures; new harness checks (badge-integrity, allowlist, links) are written test-first and gate in CI (`docs-verify`). | ✅ Pass |
| VII. AI-Native, Secure by Default | MCP use-case walkthrough documents read-only default, `[confirm]` gating, and env-only secrets exactly as implemented. | ✅ Pass |
| VIII. Go Engineering Discipline | Any harness Go code is gofumpt/goimports-clean and passes `golangci-lint run`. | ✅ Pass |

**Engineering constraints**: Docker-only testing honored (`docs-check` runs in the compose
harness). "Surface changes carry their docs" is N/A — there is no DSL/CLI surface change.

**Result**: No violations. Complexity Tracking table is empty (below).

## Project Structure

### Documentation (this feature)

```text
specs/009-docs-and-badges/
├── plan.md              # This file (/speckit-plan output)
├── research.md          # Phase 0 output — badge providers + docs-structure framework
├── data-model.md        # Phase 1 output — doc-set model (page types, index, badge set)
├── quickstart.md        # Phase 1 output — how to validate the feature end-to-end
├── contracts/           # Phase 1 output — badge contract + page/nav contract
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root)

```text
README.md                     # Header refresh + status-badge row (US1, FR-001..006a)

docs/
├── README.md                 # Goal-oriented "I want to…" index (FR-013) — the nav surface
├── overview.md               # Kept (explanation) — linked from index
├── getting-started.md        # Kept (tutorial) — linked from index
├── how-to/                   # Task-oriented recipes (folded from guides/) — FR-007
│   ├── dependencies-and-hooks.md
│   ├── parameters.md
│   ├── caching.md
│   ├── parallelism.md
│   ├── executors.md
│   ├── settings-and-dotenv.md
│   ├── imports-and-modules.md
│   └── os-filtering.md
├── user-guide/               # Ordered, readable capability tour (FR-008)
│   └── README.md             # + narrative chapters weaving the how-tos together
├── use-cases/                # Project-shaped walkthroughs (FR-009..012)
│   ├── python-project.md
│   ├── node-project.md
│   └── mcp-agents.md
├── cli.md                    # Kept (reference)
├── runefile.md               # Kept (reference)
├── GRAMMAR.md                # Kept (reference)
├── mcp.md                    # Kept (reference) — deep MCP/security model
├── docker.md                 # Kept
├── installation.md           # Kept
├── releasing.md              # Kept
├── troubleshooting.md        # Kept
├── guides/                   # Replaced by short redirect stubs → how-to/ (no broken links)
│   └── *.md                  # each: one-line "moved to ../how-to/<page>.md"
└── examples/                 # Existing 15-example library (unchanged content; index refresh)
    └── README.md

test/docs/                    # Existing verification harness — EXTENDED
├── badges_test.go            # NEW — badge-integrity: canonical repo/module targets, well-formed URLs
├── codeblocks_test.go        # selfContainedPages allowlist grown to new/reorganized pages
├── links_test.go             # already scans docs/**/*.md + README.md; covers reorg + stubs
└── …                         # examples_test, contract_test, cli_test unchanged
```

**Structure Decision**: Single-project layout. Documentation is reorganized *in place* under
`docs/` following Diátaxis (tutorial = getting-started; how-to = folded guides; reference =
cli/runefile/grammar/mcp; explanation = overview + user-guide narrative; use-cases as guided
walkthroughs). Old `docs/guides/*` paths become one-line redirect stubs so inbound links keep
resolving (there is no server-side redirect on GitHub). Verification stays in the existing
`test/docs` Go package, extended rather than replaced.

## Complexity Tracking

> No constitution violations — no entries required.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
