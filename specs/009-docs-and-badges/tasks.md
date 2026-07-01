---
description: "Task list for Modern, Example-Rich Documentation & README Status Badges"
---

# Tasks: Modern, Example-Rich Documentation & README Status Badges

**Input**: Design documents from `/specs/009-docs-and-badges/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: INCLUDED. Rune treats documentation as tested fixtures (Constitution Principle VI;
CI gate `docs-verify`). The verification harness tasks below are written test-first, per the
Q2 clarification (accuracy enforced as an ongoing gate).

**Organization**: Grouped by user story. After Phase 2, US1/US2/US3 are independently testable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- All docs commands run **inside Docker** (`rune docs-check` → `docker-compose run --rm test go test ./test/docs/...`). Use `go run ./cmd/rune <task>` if `rune` isn't installed.

## Path Conventions

Single project. Docs live under `docs/`; the verification harness under `test/docs/`; the
front door is `README.md`. All in-repo doc links are **relative**.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish a green baseline and the new directory skeleton.

- [X] T001 Confirm green baseline — run `rune docs-check` and `docker-compose run --rm test go test ./...`; record that they pass before any change (no files modified).
- [X] T002 [P] Create empty directories `docs/how-to/`, `docs/user-guide/`, `docs/use-cases/` (add a temporary `.gitkeep` in each; removed in T026).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The mechanical reorganize-in-place that establishes final page paths. Both US2
(use-cases link into how-to) and US3 (index + how-to polish) depend on this.

**⚠️ CRITICAL**: US2 and US3 cannot begin until this phase is complete. (US1 is independent and
may start in parallel with Phase 2.)

- [X] T003 `git mv` the eight capability pages `docs/guides/{dependencies-and-hooks,parameters,caching,parallelism,executors,settings-and-dotenv,imports-and-modules,os-filtering}.md` → `docs/how-to/` (history preserved).
- [X] T004 Update every internal link to the moved pages, repo-wide and relative, in `README.md` (docs table, lines ~78-88), `docs/overview.md`, `docs/getting-started.md`, `docs/examples/README.md`, each `docs/examples/*/README.md`, `docs/mcp.md`, `docs/runefile.md`, and `docs/troubleshooting.md`.
- [X] T005 [P] Leave one-line redirect stubs at each old path `docs/guides/{...}.md` per `contracts/docs-structure.md` §C5 (`# Moved` + `> [!NOTE]` pointer to `../how-to/<page>.md`).
- [X] T006 [P] Convert `docs/guides/README.md` into a stub pointing to the new `docs/README.md` index and `docs/how-to/`.
- [X] T007 Verify reorg integrity — `rune docs-check` passes (`links_test.go` reports zero broken internal links across moved pages + stubs).

**Checkpoint**: Final doc paths exist and all links resolve — US2 and US3 can start.

---

## Phase 3: User Story 1 - Size up the project in seconds from the README (Priority: P1) 🎯 MVP

**Goal**: A README front door — light header refresh + a live status-badge row (CI, release/
tag, license, Go version, Go Report Card, docs, Go Reference).

**Independent Test**: Open the rendered README on GitHub; badge row is above the fold, every
badge resolves to live state and links to the correct source, legible in light & dark themes;
`badges_test.go` passes.

### Tests for User Story 1 (write first — must fail before T009)

- [X] T008 [P] [US1] Add `test/docs/badges_test.go` asserting the badge integrity rules in `contracts/badges.md` (canonical repo `glapsfun/rune` vs module `rune-task-runner/rune`, no placeholder tokens, non-empty `alt` on every badge image, image wrapped in the correct link, real `.github/workflows/ci.yml` referenced; shape only, no live HTTP). Confirm it FAILS against the current README.

### Implementation for User Story 1

- [X] T009 [US1] Add the light header refresh + centered `<p align="center">` badge row to `README.md` using the exact snippets in `contracts/badges.md` (`<a><img alt=…></a>` pairs, one shields style), plus quick-nav links into `docs/README.md`; keep existing prose/tables.
- [X] T010 [US1] Verify — `rune docs-check` passes (`badges_test.go` green, links green). Confirm the module path `github.com/rune-task-runner/rune` actually resolves on `pkg.go.dev` and `goreportcard.com` (visit both once); if it does NOT (the module is served from `glapsfun/rune` without a vanity-import redirect), either configure the vanity import or replace the Go Reference + Go Report Card badges with repo-scoped alternatives BEFORE shipping. Note the first-visit indexing step in `quickstart.md`.

**Checkpoint**: README front door complete and enforced. MVP shippable.

---

## Phase 4: User Story 2 - Follow a use-case walkthrough for my kind of project (Priority: P2)

**Goal**: Project-shaped walkthroughs for Python, Node, and MCP/AI-agent, anchored to the
existing examples, showing features + why + expected output.

**Independent Test**: Using only the relevant use-case page, copy the example and run its task
— succeeds first try, output matches; the page names the Rune features it used.

### Implementation for User Story 2

- [X] T011 [P] [US2] Author `docs/use-cases/python-project.md` — anchored to `docs/examples/python-project/`, per `contracts/docs-structure.md` §C3 (features paired, why, command+output blocks, ≤2 GitHub Alerts, next-steps footer).
- [X] T012 [P] [US2] Author `docs/use-cases/node-project.md` — anchored to `docs/examples/node-project/`, same shape.
- [X] T013 [P] [US2] Author `docs/use-cases/mcp-agents.md` — anchored to `docs/examples/agent-driven/`; state the security posture (read-only default, `[confirm]` gating, env-only secrets) at the relevant point (FR-012).
- [X] T014 [US2] Cross-link the three use-cases from `docs/examples/README.md` and add any self-contained `rune` blocks to `test/docs/codeblocks_test.go` `selfContainedPages`.
- [X] T015 [US2] Verify — `rune docs-check` passes (use-case pages validate, backing examples run Tier A/B, links resolve).

**Checkpoint**: A Python/Node/MCP user can succeed from these pages alone.

---

## Phase 5: User Story 3 - Learn any capability from a how-to recipe or the user guide (Priority: P2)

**Goal**: A goal-oriented index, a readable user-guide tour that links out, and consistent
modern presentation across the how-to pages.

**Independent Test**: From `docs/README.md`, reach any target page in ≤2 clicks; each page has a
consistent structure, output-bearing snippets show output, and a next-steps footer.

### Implementation for User Story 3

- [X] T016 [P] [US3] Author `docs/README.md` as the intent-first router — an `| I want to… | Go to |` table + a "by document type" section per `contracts/docs-structure.md` §C4 (FR-013, ≤2 clicks).
- [X] T017 [P] [US3] Author `docs/user-guide/README.md` as an ordered tour that links into how-to/reference/explanation without duplicating them (FR-008).
- [X] T018 [US3] Polish the eight `docs/how-to/*.md` pages to the per-page shape (§C2): title/intro, next-steps footer, GitHub Alerts for pitfalls, command+expected-output block pairs.
- [X] T019 [US3] Add the now-compliant how-to and user-guide pages to `test/docs/codeblocks_test.go` `selfContainedPages` to tighten coverage.
- [X] T020 [US3] Update the top-level `README.md` Documentation section to route through `docs/README.md`, and refresh `docs/examples/README.md` to reference the new structure.
- [X] T021 [US3] Verify — `rune docs-check` passes; manually confirm ≤2-click navigation and that GitHub Alerts / `<details>` render.

**Checkpoint**: The docs read as one navigable, modern handbook.

---

## Phase 6: User Story 4 - Trust that the docs are accurate and won't rot (Priority: P3)

**Goal**: Accuracy enforced on every change (examples run, links resolve, badges target the
right repo) — extending the existing gate.

**Independent Test**: Intentionally break a link / a badge target / an example; `rune docs-check`
FAILS; revert → PASS.

### Implementation for User Story 4

- [X] T022 [US4] Confirm the gate covers the new tree — verify `test/docs/links_test.go` walks `docs/how-to/`, `docs/use-cases/`, and `docs/user-guide/` (plus `README.md`), and that `badges_test.go` (T008) runs under `rune docs-check` / CI `docs-verify`. If any new directory is outside the walk root, widen it.
- [X] T023 [US4] Prove drift is caught — following `quickstart.md` §2, verify a broken internal link, a crossed badge target, and a broken example each make `rune docs-check` FAIL; document the result.
- [ ] T024 [US4] Confirm Constitution gate #6 `docs-verify` is green across the reorganized set on the CI matrix (Tier-A everywhere; Tier-B where interpreters exist).

**Checkpoint**: Accuracy is self-enforcing going forward.

---

## Phase 7: Polish & Cross-Cutting Concerns

- [X] T025 [P] Run the full `quickstart.md` validation end-to-end.
- [X] T026 Remove temporary `.gitkeep` files from T002 and do a final relative-link sweep.
- [ ] T027 [P] Manual GitHub render check — push the branch; verify README badge row + key pages render in light & dark themes and at mobile width (SC-006); click Go Report Card / Go Reference links once to trigger indexing.
- [X] T028 Non-regression — `docker-compose run --rm test go test ./...` and `-race` pass with **zero golden changes** (SC-008).
- [X] T029 [P] Lint the new Go — run `rune lint` (golangci-lint / gofumpt / goimports) so `test/docs/badges_test.go` is clean (Constitution Principle VIII; CI gate #1).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (P1)**: no dependencies.
- **Foundational (P2)**: after Setup. Blocks US2 and US3. (US1 does not depend on it.)
- **US1 (P3 phase)**: after Setup; independent of Foundational — can run in parallel with P2.
- **US2 (P4)** and **US3 (P5)**: after Foundational.
- **US4 (P6)**: after US1 (badges_test) + the docs content it enforces (US2/US3) exist.
- **Polish (P7)**: after all desired stories.

### User Story Dependencies

- **US1 (P1)**: independent — touches `README.md` + `test/docs/badges_test.go` only.
- **US2 (P2)**: needs final paths (Foundational). Independent of US1/US3.
- **US3 (P2)**: needs final paths (Foundational). Independent of US1/US2.
- **US4 (P3)**: verifies the gate over US1–US3 output.

### Within Each User Story

- US1: test (T008, fails) → implement (T009) → verify (T010).
- US2/US3: authoring tasks marked [P] are independent files; allowlist/link updates come after.

### Parallel Opportunities

- T005 + T006 (stubs) in parallel within Foundational.
- **US1 can run entirely in parallel with Foundational** (different files).
- After Foundational: US2 and US3 proceed in parallel; within US2, T011/T012/T013 in parallel; within US3, T016/T017 in parallel.

---

## Parallel Example: User Story 2

```bash
# After Foundational, author the three walkthroughs together (different files):
Task: "Author docs/use-cases/python-project.md"
Task: "Author docs/use-cases/node-project.md"
Task: "Author docs/use-cases/mcp-agents.md"
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Phase 1 Setup → 2. Phase 3 US1 (README badges + `badges_test.go`) → 3. **STOP & VALIDATE**
on GitHub → 4. Ship the credible front door. (US1 needs no reorg.)

### Incremental Delivery

1. Setup + Foundational (reorg) → paths final, links green.
2. US1 badges → MVP front door.
3. US2 use-cases → Python/Node/MCP walkthroughs.
4. US3 index + how-to polish + user-guide → the handbook.
5. US4 → lock in accuracy enforcement.
6. Polish → render checks + zero-golden non-regression.

### Notes

- `rune docs-check` (Runefile task) is the local entry point to the CI `docs-verify` gate — the same `test/docs` harness.
- [P] = different files, no dependencies. Every task names its file(s).
- Commit after each task or logical group; keep all in-repo links relative.
- Never touch engine/CLI/DSL code or golden files — this feature is docs + README only.
