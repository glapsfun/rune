---
description: "Task list for 004-rich-documentation"
---

# Tasks: Rich, Example-Driven Documentation & Easy-Start Contributing

**Input**: Design documents from `/specs/004-rich-documentation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: This feature's "tests" are the **documentation verification harness** (`test/docs`).
The constitution (Principle VI, NON-NEGOTIABLE) mandates test-first, multi-layer verification,
so the harness is built in the Foundational phase **before** the content it verifies. The
harness legitimately starts RED for not-yet-written content (Red→Green→Refactor); each story
drives it to green.

**Organization**: Tasks are grouped by user story (US1–US4 from spec.md) so each is
independently implementable, testable, and shippable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no incomplete dependencies)
- **[Story]**: US1/US2/US3/US4 (setup, foundational, polish carry no story label)
- Exact file paths are included in every task.

## Path Conventions

Documentation lives under `docs/`; runnable examples under `docs/examples/<use-case>/` (each =
`Runefile` + `README.md` per `contracts/example-contract.md`); the verification harness is the
Go package `test/docs/` (mirrors `test/integration/`). Entry docs: `README.md`,
`CONTRIBUTING.md`. Dev-workflow wiring: repo-root `Runefile`, `.github/workflows/ci.yml`. This
feature changes **docs + tooling only** — no product code (FR-024).

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Scaffolding the content tree and harness package depend on. Non-blocking, no
behavior.

- [X] T001 Create the documentation skeleton per plan.md: directories `docs/guides/`, `docs/examples/` (with placeholder group dirs), and `test/docs/`; add empty placeholder pages `docs/overview.md`, `docs/troubleshooting.md`, and `docs/examples/README.md`.
- [X] T002 [P] Add `test/docs/doc.go` with a package comment describing the documentation verification harness (idiomatic Go, per Principle VIII / golang-pro).

**Checkpoint**: Directory tree and harness package exist; nothing verified yet.

---

## Phase 2: Foundational (Verification Harness — Blocking Prerequisite)

**Purpose**: The cross-cutting test infrastructure every story relies on to stay honest
(FR-015, FR-016, FR-017; SC-003, SC-005). Built test-first per Constitution Principle VI;
mirrors `test/integration/harness_test.go`. Stdlib-only (Principle V/VIII).

**⚠️ CRITICAL**: No user-story content is considered "verified" until this harness exists.

- [X] T003 Implement `test/docs/harness_test.go`: `TestMain` finds module root via `go.mod`, builds `./cmd/rune` once into a temp dir (`CGO_ENABLED=0`), runs `m.Run()`, cleans up; add helpers `runeCheck(t, runefilePath)` and `runTask(t, dir, args...)` returning `{stdout, stderr, exit}`, each using a per-call `context.Context` timeout and wrapping errors with `%w`. Follow `test/integration/harness_test.go` exactly.
- [X] T004 [P] Implement Tier-A example static check in `test/docs/examples_test.go`: discover every `docs/examples/*/Runefile`, table-driven subtests (`t.Run` per example, range var captured) asserting **`rune --file <path> --list`** exits 0 — parse+analyze, runs nothing. **NB (finding F1):** there is no built-in `rune check`; use `--list` (`--dry-run`/`--dump` also force analysis but `--list` needs no task arg). Use `exec.CommandContext` with a timeout; `t.Helper()` on helpers; wrap errors with `%w`.
- [X] T005 [P] Implement `test/docs/codeblocks_test.go`: extract every **complete** fenced ` ```rune ` block from `docs/**/*.md`, write each to a temp file, assert `rune --file <tmp> --list` exits 0. Per the fenced-block convention (research D2), only ` ```rune ` blocks that are complete Runefiles are validated — deliberate fragments are fenced ` ```text `; a parse/analyze failure in a ` ```rune ` block is a test failure.
- [X] T006 [P] Implement `test/docs/links_test.go`: scan `docs/**/*.md`, `README.md`, `CONTRIBUTING.md` for relative Markdown links and `#anchors`; fail on any unresolved file target or missing heading anchor; skip `http(s)` links (offline/deterministic — research D4). Fix any currently-broken internal links surfaced.
- [X] T007 [P] Implement contract + safety checks in `test/docs/contract_test.go`: (a) every `docs/examples/*/` has `Runefile` + `README.md` with the required sections (`contracts/example-contract.md`); (b) no secret/token/key-shaped literal appears in any example or doc (Principle VII); (c) no forbidden glossary alias appears in any page body (`contracts/information-architecture.md`, FR-016).
- [X] T008 Repoint the `docs-check` task in the repo-root `Runefile` from the single-file `rune check` to `docker-compose run --rm test go test ./test/docs/...` (research D11).
- [X] T009 Update `.github/workflows/ci.yml` so documentation verification runs as a required check (finding N1): run `go test ./test/docs/...` **natively on each OS leg** (Linux/macOS/Windows) — Docker is Linux-only and is the **local** flow (the `docs-check` task and `-race` runs), not the cross-OS CI mechanism, exactly as `test/integration` already runs. Add `docs-verify` to the required status checks.

**Checkpoint**: Harness compiles and runs in Docker; it correctly verifies existing content and
reds-out gaps. `rune docs-check` now drives the harness. Stories below turn each gate green.

---

## Phase 3: User Story 1 — Grasp the idea and run your first task (Priority: P1) 🎯 MVP

**Goal**: A newcomer understands what Rune is (the "main idea") and reaches a first successful
task run quickly.

**Independent Test**: Read only `docs/overview.md` and state Rune's purpose/distinction in <3
min (SC-001); follow only `docs/getting-started.md` from nothing to a run in <5 min (SC-002).
The harness verifies the getting-started example, its ` ```rune ` blocks, and link integrity.

- [X] T010 [P] [US1] Write `docs/overview.md`: problem, core mental model ("named blocks of commands, always run when asked, shared by humans and agents"), differentiators, and explicit **"use Rune when… / don't use when…"** (FR-001, FR-002); include an orientation opener and a next-step footer (FR-014).
- [X] T011 [US1] Rework `docs/getting-started.md` into one linear install→write→run path with every command/file copy-pasteable, **expected output shown inline**, cross-platform call-outs where they matter (FR-003, FR-018), ending with a named next step to the examples library/language guide (FR-004). Ensure all ` ```rune ` blocks pass the codeblocks check (T005).
- [X] T012 [US1] Add `docs/examples/getting-started/README.md` conforming to `contracts/example-contract.md` (purpose, `Prerequisites: none`, run command `rune greet`, expected output, link to the relevant guide) so the existing example passes T007's contract check and Tier-B run+assert.
- [X] T013 [US1] Update `README.md` to point to `docs/overview.md` and `docs/examples/`, keep the hero example/install accurate, and use only canonical glossary terms (FR-016); fix any links flagged by T006.

**Checkpoint**: MVP — a stranger can understand Rune and run a first task; harness green for US1
content.

---

## Phase 4: User Story 2 — Find a runnable example for my use case (Priority: P2)

**Goal**: A use-case-organized library of runnable, verified examples covering every headline
capability and common project shape (the heart of the request).

**Independent Test**: Run `docker-compose run --rm test go test ./test/docs/...` — every example
passes Tier-A `--list` validation, shell-only ones pass Tier-B run+assert, interpreter/agent ones skip with
a logged reason, and the coverage gate is green (SC-003, SC-004). Browse `docs/examples/README.md`
and locate an example for a stated use case by group label alone (FR-005).

> Each example task delivers a directory `docs/examples/<id>/` = `Runefile` + `README.md` per
> `contracts/example-contract.md` (purpose, prerequisites, capability + guide link, run command,
> expected output). All are `[P]` — distinct directories, no shared files.

- [X] T014 [US2] Write `docs/examples/README.md` as the library index, grouping examples by use case ("Project shapes", "Capability spotlights") with a one-line label per group and per example (FR-005); orientation opener + links to every example.

**Project shapes**

- [X] T015 [P] [US2] `docs/examples/go-service/` — a compiled-language service (fetch→build→test dependencies, params); `Prerequisites: none` (uses `sh`); demonstrates dependencies + parameters.
- [X] T016 [P] [US2] `docs/examples/node-project/` — a Node/JavaScript project's tasks; `Prerequisites: node`; Tier-B skips when node absent.
- [X] T017 [P] [US2] `docs/examples/python-project/` — a Python project's tasks; `Prerequisites: python3`.
- [X] T018 [P] [US2] `docs/examples/monorepo/` — composition via imports + a namespaced module; `Prerequisites: none`; demonstrates imports/modules.
- [X] T019 [P] [US2] `docs/examples/ci-cd/` — a CI pipeline using `--list`, `--dry-run`, and exit codes; `Prerequisites: none`; demonstrates operational tooling + predictable exit codes.
- [X] T020 [P] [US2] `docs/examples/docker-workflow/` — running Rune in a container; `Prerequisites: docker`; Tier-B skips when docker absent.
- [X] T021 [P] [US2] `docs/examples/polyglot/` — one Runefile mixing `sh`, `python`, and `node` bodies; `Prerequisites: python3, node`; demonstrates executors.
- [X] T022 [P] [US2] `docs/examples/agent-driven/` — an agent task + MCP exposure; `Prerequisites: an agent CLI`; **no secret literals**, read-only-by-default, gated destructive task shown (Principle VII).

**Capability spotlights**

- [X] T023 [P] [US2] `docs/examples/dependencies/` — prerequisites run-once memoization + post-hooks; `Prerequisites: none`.
- [X] T024 [P] [US2] `docs/examples/parameters/` — defaults + required + variadic params with interpolation; `Prerequisites: none`.
- [X] T025 [P] [US2] `docs/examples/caching/` — explicit `[cache(inputs=…, outputs=…)]` opt-in showing the visible "cached" log on re-run (Principle I); `Prerequisites: none`.
- [X] T026 [P] [US2] `docs/examples/parallel/` — independent prerequisites run with bounded concurrency; `Prerequisites: none`.
- [X] T027 [P] [US2] `docs/examples/settings-dotenv/` — project settings + `.env` loading + exported vars; `Prerequisites: none`.
- [X] T028 [P] [US2] `docs/examples/os-filtering/` — an OS-restricted task and its listing/dispatch behavior; `Prerequisites: none`; cross-platform call-outs (FR-018).
- [X] T029 [US2] Add the **coverage-completeness assertion** to `test/docs/contract_test.go`: fail unless every Coverage-Matrix capability and project shape (`data-model.md`) has ≥1 conforming example — turning SC-004 into a hard gate (depends on T014–T028).

**Checkpoint**: Library complete and verified; coverage gate green; readers can find a starting
point for any common job.

---

## Phase 5: User Story 3 — Learn a capability in depth (Priority: P3)

**Goal**: Task-oriented guides (concept→syntax→runnable example→pitfalls) for every capability,
an accurate CLI reference, and a troubleshooting page mapping failures to diagnostics.

**Independent Test**: Each guide contains all four required elements (SC-008); a reader can
answer a realistic "how do I…/what happens when…" from one guide; `docs/cli.md` matches the
binary (drift check green, FR-013). Guides cross-link to their US2 examples (so best sequenced
after US2 to avoid dead-end links — FR-014).

> Each guide page MUST follow `contracts/information-architecture.md` (orient opener, next-step
> footer, ≥1 example link) and is `[P]` — distinct files.

- [X] T030 [P] [US3] `docs/guides/dependencies-and-hooks.md` — links to `docs/examples/dependencies/`.
- [X] T031 [P] [US3] `docs/guides/parameters.md` — links to `docs/examples/parameters/`.
- [X] T032 [P] [US3] `docs/guides/caching.md` — explains always-run vs. opt-in cache + the fingerprint; links to `docs/examples/caching/` (Principle I).
- [X] T033 [P] [US3] `docs/guides/parallelism.md` — links to `docs/examples/parallel/`.
- [X] T034 [P] [US3] `docs/guides/executors.md` — sh/python/node/agent bodies; links to `docs/examples/polyglot/`.
- [X] T035 [P] [US3] `docs/guides/settings-and-dotenv.md` — links to `docs/examples/settings-dotenv/`.
- [X] T036 [P] [US3] `docs/guides/imports-and-modules.md` — name resolution + collisions; links to `docs/examples/monorepo/`.
- [X] T037 [P] [US3] `docs/guides/os-filtering.md` — links to `docs/examples/os-filtering/`.
- [X] T038 [P] [US3] `docs/guides/agents-and-mcp.md` — supersede/expand `docs/mcp.md`; lead with the **security model** (read-only default, env-only secrets, gated destructive tasks — FR-012); links to `docs/examples/agent-driven/`. **Finding I2:** decide whether `docs/mcp.md` is removed or kept as a one-line redirect stub, and update the `README.md` documentation-table row (and any other refs) that currently point to `docs/mcp.md` — otherwise the links check (T006) fails / readers hit a moved page.
- [X] T039 [US3] Rework `docs/cli.md` to mirror `contracts/cli-reference.md` exactly: all global flags, reserved subcommands (`mcp`/`serve`/`completion`), and exit codes `0/1/2/3/130` (FR-013).
- [X] T040 [US3] Add the **CLI-reference drift check** to `test/docs` (e.g. `cli_test.go`): assert every flag in `docs/cli.md` ⇔ the binary's `--help`, and the documented exit-code set equals `{0,1,2,3,130}` (research D7) — depends on T039.
- [X] T041 [P] [US3] Write `docs/troubleshooting.md`: map each failure mode (unknown task, undefined variable, dependency cycle, arity mismatch, missing interpreter, cache miss) to its expected behavior, the real `file:line:col`+caret diagnostic, and exit code (FR-011, Principle II); verify wording against actual binary output.
- [X] T042 [P] [US3] Refine `docs/runefile.md` and `docs/installation.md`: add orientation openers + next-step footers, cross-link to the new guides/examples, and add cross-platform call-outs (FR-014, FR-018).

**Checkpoint**: Every capability has a complete guide; CLI reference is drift-checked; failure
modes documented.

---

## Phase 6: User Story 4 — Make my first contribution without friction (Priority: P4)

**Goal**: A reworked `CONTRIBUTING.md` that takes a newcomer from "I'd like to help" to a
verified change with minimal friction.

**Independent Test**: Following only `CONTRIBUTING.md` from a clean clone, a newcomer adds a
small change (e.g. a new example) and runs the checks the supported way in <15 min (SC-007),
without asking a maintainer how to begin.

- [X] T043 [US4] Rework `CONTRIBUTING.md` around a newcomer's path (research D10): **What to contribute** (lead with low-barrier wins — doc fixes, new examples — FR-019) → **Set up** (prereqs, clone, build) → **Verify** (Docker-only testing policy explained accessibly, exact commands, and the `go run ./cmd/rune <task>` fallback when `rune` isn't installed — FR-021) → **Repo map** (where things live) → **Propose it** (PR + which CI gates run — FR-020, FR-022).
- [X] T044 [US4] Add an "Add a new example" walkthrough to `CONTRIBUTING.md` that points at `contracts/example-contract.md` and the example library, positioning it as the easiest first contribution (FR-019) and noting the harness verifies it automatically.
- [X] T045 [US4] Consistency pass: reconcile `CONTRIBUTING.md` with the user docs and `.specify/memory/constitution.md` so commands, policies (Docker-only testing, CI gates), and terminology agree — zero contradictions (FR-023, SC-009).

**Checkpoint**: A first-time contributor can onboard and ship a small change unaided.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: End-to-end validation, consistency, and engineering-quality gates across the set.

- [X] T046 Run the full `quickstart.md` validation (Scenarios 1–9) inside Docker; record results.
- [X] T047 [P] Verify the navigation guarantee: from `README.md`, every guide and example is reachable in ≤2 clicks (SC-006) — confirm via the link graph + manual spot-check.
- [X] T048 [P] Run the consistency review against `specs/004-rich-documentation/checklists/requirements.md` (SC-009) and resolve any contradictions.
- [X] T049 Ensure `test/docs` passes `golangci-lint run` + `gofumpt` clean and `docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./test/docs/...` is green (Principle VIII, VI).
- [X] T050 [P] Final terminology sweep: confirm zero forbidden glossary aliases across all pages (FR-016) — the T007 check is green set-wide.
- [X] T051 Manual timing/comprehension checks: SC-001 (<3 min to grasp from overview), SC-002 (<5 min to first task), SC-007 (<15 min first contribution); note outcomes and fix any blockers.
- [X] T052 [P] Enforce the scope boundary (FR-024, finding C1): add a guard that asserts the feature's diff touches **only** `docs/`, `test/docs/`, `Runefile`, `README.md`, `CONTRIBUTING.md`, and `.github/workflows/ci.yml` — no product packages (`cmd/`, `internal/`) change. Implement as a small CI step (`git diff --name-only` against the merge base, filtered) or a documented review check; if doc work surfaces a real product defect, log it separately rather than editing product code here.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: no dependencies — start immediately.
- **Foundational (Phase 2)**: depends on Setup; **blocks meaningful verification** of all
  stories. The harness can be authored before content exists (test-first).
- **User Stories (Phase 3–6)**: depend on Foundational. Recommended order P1→P2→P3→P4:
  - US1 is the MVP and is fully independent.
  - US2 is independent of US1/US3/US4.
  - US3 guides **cross-link to US2 examples** (FR-014), so US3 is best sequenced **after US2**
    to avoid dead-end links; otherwise independent.
  - US4 references the whole set, so it is best **last**.
- **Polish (Phase 7)**: depends on the desired stories being complete.

### Key task dependencies

- T008/T009 (wiring) depend on T003 (harness exists).
- T029 (coverage gate) depends on T014–T028 (all examples exist).
- T040 (CLI drift check) depends on T039 (cli.md reworked).
- US3 guide tasks (T030–T038) depend on their matching US2 example existing (cross-link target).

### Parallel Opportunities

- T004–T007 (separate harness test files) run in parallel after T003.
- **All US2 example tasks T015–T028 run in parallel** (distinct directories) — the biggest
  fan-out; then T029.
- **All US3 guide tasks T030–T038 + T041 + T042 run in parallel** (distinct files).
- Different stories can be staffed in parallel once Foundational is done (mind the US2→US3
  cross-link note).

---

## Parallel Example: User Story 2 (the example library)

```bash
# After Foundational, launch every example directory in parallel (distinct dirs):
Task: "Create docs/examples/go-service/ (Runefile + README per example-contract)"
Task: "Create docs/examples/node-project/ ..."
Task: "Create docs/examples/python-project/ ..."
Task: "Create docs/examples/monorepo/ ..."
Task: "Create docs/examples/ci-cd/ ..."
Task: "Create docs/examples/docker-workflow/ ..."
Task: "Create docs/examples/polyglot/ ..."
Task: "Create docs/examples/agent-driven/ ..."
Task: "Create docs/examples/dependencies/ , parameters/ , caching/ , parallel/ , settings-dotenv/ , os-filtering/"
# Then (barrier): add the coverage-completeness assertion (T029).
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Phase 1 Setup → Phase 2 Foundational (the harness) → Phase 3 US1.
2. **STOP and VALIDATE**: read overview (SC-001), run getting-started (SC-002); harness green
   for US1 content. This alone makes Rune understandable + immediately usable.

### Incremental Delivery

1. Setup + Foundational → harness in place (CI gate live).
2. US1 → understandable + first-run (MVP) → ship.
3. US2 → the rich example library + coverage gate → ship.
4. US3 → deep guides + accurate reference + troubleshooting → ship.
5. US4 → easy-start CONTRIBUTING → ship.
6. Polish → validate quickstart, consistency, timings, race/lint.

### Notes

- `[P]` = different files, no incomplete dependencies.
- The harness is the test layer (Principle VI); expect it RED until each story's content lands.
- Tests run **inside Docker** only (project policy); use `go run ./cmd/rune <task>` if `rune`
  isn't installed.
- Commit after each task or logical group; stop at any checkpoint to validate a story.
- Avoid: invented DSL in examples (must pass `rune --file <path> --list`), secrets in any
  doc/example, Linux-only examples, and cross-story dependencies that break independence.
