---
description: "Task list for Minimum Rune Version"
---

# Tasks: Minimum Rune Version

**Input**: Design documents from `/specs/010-minimum-rune-version/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: INCLUDED — the spec's acceptance criteria explicitly require unit, integration,
golden-diagnostic, and cross-platform tests, and Constitution Principle VI mandates
test-first development. Write each test task and confirm it FAILS before implementing.

**Testing policy**: All Go tests run inside Docker only:
`docker-compose run --rm test go test ./...` (race: `-e CGO_ENABLED=1 … -race`).

**Organization**: Tasks are grouped by user story (US1–US4) for independent implementation
and testing. MVP = User Story 1.

> **Revision note**: This list incorporates `/speckit-analyze` findings — added a
> root-ownership test (G1/FR-012/SC-007), a `rune_version`+`minimum_version` coexistence
> test (G2/FR-013), the two distinct guard messages (B1), and cross-platform/docs notes
> (C1/I2). Test file consolidated to `test/integration/minimum_version_test.go` (I1).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: US1–US4; setup/foundational/polish carry no story label

### Error-message contract (locked, per analysis B1)

- **Non-static value** → `minimum_version must be a static semantic version`
- **Invalid semantic version literal** (incl. `"0.8"`, `"latest"`, `"v0.8.0"`, ranges like
  `">=0.8,<1.0"`) → `minimum_version must be a valid semantic version` (message names the
  offending value)
- **Incompatible** → `this Runefile requires Rune >= <req>` + `installed`/`required`/`upgrade` notes + `nothing was executed`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Baseline and new-package scaffolding.

- [x] T001 Confirm baseline is green: run `docker-compose run --rm test go test ./...` and record it passes before any changes
- [x] T002 [P] Create the `internal/semver` package with package declaration and a `Version` struct stub (Major/Minor/Patch int, Prerelease/Build []string) in `internal/semver/semver.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: SemVer comparator, the root-only setting extractor, installed-version
plumbing, and the shared upgrade-URL constant. Every user story depends on these.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T003 [P] Write table-driven unit tests for the SemVer comparator in `internal/semver/semver_test.go`: `Parse` accepts `MAJOR.MINOR.PATCH[-pre][+build]` and rejects malformed / `v`-prefixed / partial (`0.8`) input; `Compare` orders core numerically; prerelease ranks below the equal release (`0.9.0-rc.1 < 0.9.0`); build metadata ignored (`0.9.0+abc == 0.9.0`); `Satisfies(min)` is `Compare(min) >= 0`. Confirm the tests FAIL.
- [x] T004 Implement `Parse`, `(Version).Compare`, and `(Version).Satisfies` per SemVer 2.0.0 in `internal/semver/semver.go` to make T003 pass
- [x] T005 Add an `IgnoreVersion bool` field to `Options` in `internal/cli/dispatch.go` (installed version is already carried as `Options.Version`); no behavior change yet
- [x] T006 Add a test-only installed-version hook in `cmd/rune/main.go`: when `RUNE_TEST_VERSION` is set, use it as the reported `version` before `newRootCmd` (documented as test-only, never read from `internal/`)
- [x] T007 Create `internal/config/minimum_version.go` with the `UpgradeURL` constant (`https://github.com/glapsfun/rune/releases`) and `MinimumVersion(f *ast.File)` that scans the ROOT `f.Settings` for `minimum_version` and returns its raw literal value + `token.Span` + a "present" flag (mirroring `RuneVersion` in `internal/config/version.go`); add the test file `internal/config/minimum_version_test.go` with a parse helper. No validation/comparison logic yet.

**Checkpoint**: SemVer compare works and is unit-tested; the setting can be located on the root file before imports; installed version is injectable.

---

## Phase 3: User Story 1 - Reject an incompatible binary before execution (Priority: P1) 🎯 MVP

**Goal**: A root Runefile's `minimum_version` causes an older binary to be refused before
imports, analysis, or any execution, with a caret-anchored required/installed/upgrade
diagnostic and a non-zero exit; equal/newer versions run normally; no setting = unchanged;
only the ROOT file's value governs.

**Independent Test**: With a Runefile declaring `set minimum_version := "0.8.0"`, run
`RUNE_TEST_VERSION=0.7.2 rune build` → incompatibility diagnostic, exit 3, task never runs;
`0.8.0` and `0.9.1` → task runs; no setting → unchanged; an imported child's value never
imposes or relaxes the requirement.

### Tests for User Story 1 ⚠️ (write first, confirm they FAIL)

- [x] T008 [P] [US1] Unit tests for `CheckMinimumVersion(file, installed)` in `internal/config/minimum_version_test.go`: older→error, equal→ok, newer→ok, absent-setting→ok; assert the error carries the value literal's span. Confirm FAIL.
- [x] T009 [P] [US1] Unit test in `internal/config/minimum_version_test.go` proving `minimum_version` and `rune_version` coexist independently (both set → both honored, neither affects the other) — covers FR-013. Confirm FAIL.
- [x] T010 [P] [US1] Golden diagnostic fixture for the incompatible-version error (caret on the value literal, `installed`/`required`/`upgrade` lines, `nothing was executed` trailer) under `testdata/diag/minimum_version_incompatible.golden` per `contracts/diagnostics.md`. Confirm FAIL.
- [x] T011 [P] [US1] Integration tests in `test/integration/minimum_version_test.go` using `RUNE_TEST_VERSION`: reject-old (exit 3, stderr diagnostic, stdout empty), allow-equal, allow-new, and no-setting-unchanged. Confirm FAIL.
- [x] T012 [P] [US1] Root-ownership integration test in `test/integration/minimum_version_test.go`: a root Runefile importing a child that declares a different `minimum_version` (both a higher `9.9.9` and a lower value) — assert the ROOT's value governs, and that a child's value with no root setting imposes no requirement. Covers FR-012 / SC-007. Confirm FAIL.

### Implementation for User Story 1

- [x] T013 [US1] Implement `CheckMinimumVersion(file *ast.File, installed string) diag.List` in `internal/config/minimum_version.go`: parse the (well-formed) literal via `internal/semver`, compare against `installed`, and on failure emit `diag.New(valueSpan, …)` with required/installed/upgrade content (makes T008/T010 pass)
- [x] T014 [US1] Wire the gate in `internal/cli/run.go` between `parser.Parse` (l.57) and `config.Compose` (l.61): call `CheckMinimumVersion(file, opts.Version)` on the ROOT `file.Settings` (pre-Compose ⇒ root ownership), and on errors `renderDiags` + return `&ValidationError{}` (exit `ExitValidation`=3) so nothing else runs (makes T011/T012 pass)
- [x] T015 [US1] Wire the same gate in `internal/cli/serve.go` `loadModule` between `parser.Parse` (l.40) and `config.Compose` (l.42) so the MCP/agent static-load path is gated identically

**Checkpoint**: US1 fully functional and independently testable — the MVP.

---

## Phase 4: User Story 2 - Static-value guarding (Priority: P1)

**Goal**: `minimum_version` must be a static string literal that is a valid single semantic
version; non-literal values and invalid/range values are rejected with the locked distinct
diagnostics and zero execution.

**Independent Test**: `set minimum_version := required` (a variable) → `minimum_version must
be a static semantic version`, caret at the value, exit 3; `"0.8"` / `">=0.8,<1.0"` /
`"v0.8.0"` → `minimum_version must be a valid semantic version`, caret at the literal.

### Tests for User Story 2 ⚠️ (write first, confirm they FAIL)

- [x] T016 [P] [US2] Unit tests in `internal/config/minimum_version_test.go`: non-`*ast.StringLit` value → `minimum_version must be a static semantic version` (caret at `s.Value.Span()`); literal that fails `semver.Parse` (`"0.8"`, `"latest"`, `"v0.8.0"`, `">=0.8,<1.0"`) → `minimum_version must be a valid semantic version` (caret at the literal `Sp`). Confirm FAIL.
- [x] T017 [P] [US2] Integration/golden tests for both static-guard diagnostics in `test/integration/minimum_version_test.go` (and goldens under `testdata/diag/` if renderer output is pinned). Confirm FAIL.

### Implementation for User Story 2

- [x] T018 [US2] Extend `MinimumVersion`/`CheckMinimumVersion` in `internal/config/minimum_version.go` to enforce the `*ast.StringLit` guard (static message) and the semver-validity check (valid-semver message) before any comparison (makes T016/T017 pass)

**Checkpoint**: US1 + US2 both work; malformed requirements can never reach comparison or execution.

---

## Phase 5: User Story 3 - Inspect and check compatibility from the CLI (Priority: P2)

**Goal**: `rune version` reports installed + language version; `rune version --check` reports
required/installed/status and exits non-zero when incompatible without running a task;
`--json` emits `{installed, required, compatible, runefile}`; no-Runefile/no-requirement is
reported as "no requirement declared", not incompatible.

**Independent Test**: `rune version` prints two lines; `RUNE_TEST_VERSION=0.7.2 rune version
--check` in a project requiring `0.8.0` → incompatible, non-zero exit, no task; `--json`
prints the documented object; empty dir → "no requirement declared", exit 0.

### Tests for User Story 3 ⚠️ (write first, confirm they FAIL)

- [x] T019 [P] [US3] Integration tests in `test/integration/minimum_version_test.go`: `rune version` two-line output; `--check` compatible/incompatible (exit codes, no task run); `--check --json` shape and values; no-requirement case. Confirm FAIL.

### Implementation for User Story 3

- [x] T020 [US3] Add the language-version line (`runefile language <config.CurrentVersion>`) to bare `rune version` in `cmd/rune/version.go`
- [x] T021 [US3] Add the `--check` flag to `cmd/rune/version.go`: resolve the Runefile via `config.Resolve`, read/validate `minimum_version`, print required/installed/status, exit `ExitValidation` when incompatible, and report "no requirement declared" (exit 0) when none
- [x] T022 [US3] Add the `--json` flag (a `CompatibilityResult` DTO with `installed`/`required`/`compatible`/`runefile` tags) to `cmd/rune/version.go`, modeled on the `internal/cli/dump.go` JSON pattern (makes T019 pass)

**Checkpoint**: US1–US3 independently functional; CI/scripts can query compatibility.

---

## Phase 6: User Story 4 - Emergency override (Priority: P3)

**Goal**: A CLI-only `--ignore-version` bypasses the gate with a loud stderr warning and
proceeds; it can never be enabled from a Runefile; the MCP/agent path refuses by default and
only ignores when the operator explicitly enables it.

**Independent Test**: `RUNE_TEST_VERSION=0.7.2 rune --ignore-version build` prints
`warning: ignoring Runefile minimum Rune version 0.8.0; running 0.7.2` and runs the task; no
Runefile setting enables it; MCP with an unmet requirement refuses unless `AllowIgnoreVersion`.

### Tests for User Story 4 ⚠️ (write first, confirm they FAIL)

- [x] T023 [P] [US4] Integration tests in `test/integration/minimum_version_test.go`: `--ignore-version` emits the warning to stderr and runs the task; confirm no Runefile mechanism enables it; MCP/agent path default-refuses on incompatibility. Confirm FAIL.

### Implementation for User Story 4

- [x] T024 [US4] Add the global `--ignore-version` bool flag in `cmd/rune/root.go` (bound to `Options.IgnoreVersion`) and honor it in the `internal/cli/run.go` gate: when set and incompatible, print the warning to stderr (styled `warning:` / `diag.Warn`) and continue instead of aborting
- [x] T025 [US4] Add `AllowIgnoreVersion bool` (default false) to `mcpserver.Options` in `mcpserver/server.go`, honor it in the `serve.go loadModule` gate, and feed it from an explicit `rune serve` operator flag/env in `cmd/rune/serve.go` (never from a Runefile) (makes T023 pass)

**Checkpoint**: All four user stories independently functional.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, gate compliance, and final verification.

- [x] T026 [P] Document the `minimum_version` setting in `docs/GRAMMAR.md` (settings list, static-literal + `>=` semantics)
- [x] T027 [P] Document `minimum_version`, `--ignore-version`, and `rune version --check`/`--json` in `docs/runefile.md` (and any settings reference); note the intentional `--json` vs `--format json` convention divergence (analysis I2); keep added code blocks self-contained so `docs-verify` passes
- [x] T028 Run the full Docker suite and race variant (`docker-compose run --rm test go test ./...`; `-e CGO_ENABLED=1 … -race`), confirm the new tests are picked up by the CI OS matrix (Linux/macOS/Windows) for cross-platform coverage (analysis C1 / SC-008), regenerate goldens and confirm no diff, and run `rune docs-check`
- [x] T029 [P] `golangci-lint run` + gofumpt/goimports clean pass over new/changed files
- [x] T030 Execute the `quickstart.md` validation scenarios end-to-end (Scenarios 1–5)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: no dependencies.
- **Foundational (Phase 2)**: depends on Setup; BLOCKS all user stories. T004 depends on T003; T007 depends on the package existing (T002).
- **User Stories (Phase 3–6)**: all depend on Foundational.
  - US1 (P1) is the MVP. US2 (P1) hardens the same `config/minimum_version.go` file US1 creates, so US2's T018 depends on US1's T013.
  - US3 (P2) and US4 (P3) depend only on Foundational; they don't depend on US1/US2 implementation, though US4's `run.go` warning (T024) touches the same gate wired in T014.
- **Polish (Phase 7)**: after the desired stories are complete.

### Within Each User Story

- Tests first and failing, then implementation.
- Same-file tasks are sequential: T013 → T018 (both edit `internal/config/minimum_version.go`); T014 → T024 (both edit `internal/cli/run.go`); T020 → T021 → T022 (all edit `cmd/rune/version.go`).

### Parallel Opportunities

- Setup: T002 alongside T001.
- Foundational: T003 independent; T005/T006/T007 touch different files → parallel after T002.
- US1 tests T008/T009/T010/T011/T012 are file-disjoint (config unit vs golden vs integration) → parallel; then T013 → (T014, T015 parallel: different files).
- US2: T016/T017 parallel; then T018.
- US3: T019 first; T020–T022 all edit `version.go` → sequential.
- Cross-story: once Foundational is done, US3 (`version.go`) and US4 (`root.go`/`mcpserver`/`serve.go`) are largely file-disjoint from US1/US2 (`internal/config`) and can run in parallel.
- Polish: T026/T027/T029 parallel; T028/T030 last.

---

## Parallel Example: User Story 1

```bash
# Write the failing test groups together (different files):
Task: "Unit tests for CheckMinimumVersion in internal/config/minimum_version_test.go"
Task: "Coexistence test (rune_version + minimum_version) in internal/config/minimum_version_test.go"
Task: "Golden diagnostic fixture in testdata/diag/minimum_version_incompatible.golden"
Task: "Integration reject/allow tests in test/integration/minimum_version_test.go"
Task: "Root-ownership integration test in test/integration/minimum_version_test.go"

# After T013, wire the two gate call sites together (different files):
Task: "Gate in internal/cli/run.go"
Task: "Gate in internal/cli/serve.go loadModule"
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Phase 1 Setup → 2. Phase 2 Foundational → 3. Phase 3 US1 → **STOP & VALIDATE**: an older
binary is refused before execution with the correct diagnostic; equal/newer run; no-setting
unchanged; the root owns the requirement. This alone delivers the core protective value.

### Incremental Delivery

MVP (US1) → add US2 (static guarding) → add US3 (`rune version --check`) → add US4
(`--ignore-version` + MCP operator opt) → Polish (docs, gates, quickstart). Each story is an
independently testable, shippable increment.

### Parallel Team Strategy

After Foundational: Dev A takes US1→US2 (`internal/config`, gate wiring); Dev B takes US3
(`cmd/rune/version.go`); Dev C takes US4 (`cmd/rune/root.go`, `mcpserver`, `serve.go`).
Coordinate on the shared `internal/cli/run.go` gate (T014 before T024).

---

## Notes

- [P] = different files, no dependency on an incomplete task.
- Verify each test FAILS before implementing (Red-Green-Refactor, Principle VI).
- Golden files are regenerated deliberately, never hand-edited to pass (Principle VI).
- Commit after each task or logical group.
- Keep `minimum_version` independent from `rune_version` throughout (FR-013; verified by T009).
