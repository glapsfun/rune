# Tasks: Accurate Version Reporting for Go-Toolchain Installs

**Input**: Design documents from `/specs/019-fix-version-reporting/`

**Prerequisites**: plan.md, spec.md, research.md (D1–D7), data-model.md (E1–E2, S1–S4), contracts/version-output.md (C1–C8), quickstart.md (V1–V6)

**Tests**: INCLUDED — the constitution (Principle VI) mandates test-first Red-Green-Refactor. Every test task must be written and observed failing before its implementation task.

**Organization**: Tasks are grouped by user story. All Go tests run inside Docker (`docker-compose run --rm test go test ./...`), never on the host.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)

## Path Conventions

Single Go module at repo root; all code changes in `cmd/rune/` (plan Structure Decision — `internal/` is locked and untouched).

---

## Phase 1: Setup

**Purpose**: Confirm a green baseline so every later red/green signal is attributable to this feature.

- [X] T001 Verify baseline: `docker-compose run --rm test go test ./...` passes on branch `019-fix-version-reporting` before any code change (Docker via Rancher Desktop — start it if needed)

---

## Phase 2: Foundational

**Purpose**: N/A — no blocking prerequisites. The feature is self-contained in `cmd/rune` (one new file + a one-line call-site change); no schema, framework, or shared infrastructure precedes the stories.

*(No tasks.)*

---

## Phase 3: User Story 1 - Installed Version Is Reported Truthfully (Priority: P1) 🎯 MVP

**Goal**: A binary installed via `go install …/cmd/rune@vX.Y.Z` (or `@latest`) reports `X.Y.Z` from `rune --version` / `rune version`, while ldflags-stamped release artifacts stay byte-identical (C1–C4; FR-001, FR-002, FR-005, FR-007).

**Independent Test**: quickstart V2 pre-release form (`go install …@<pushed-fix-commit>` → pseudo-version, not `dev`) and V4 (ldflags-stamped binary output unchanged); clean-version check post-release via `@latest`.

### Tests for User Story 1 (write first, watch them fail)

- [X] T002 [US1] RED: Create `cmd/rune/buildinfo_test.go` with table-driven tests for a pure `versionFromBuildInfo(version string, bi *debug.BuildInfo, ok bool) string` covering: ldflags-stamped version (`"9.9.9"`) wins with build info ignored; module-cache build (`Main.Version "v0.4.3"`, no `vcs.*` settings) → `"0.4.3"`; module pseudo-version (`"v0.4.4-0.20260812092550-432dbb8de56b"`, no vcs) → same minus leading `v`; build info absent (`ok == false`) → `"dev"`; empty `Main.Version` → `"dev"`. Run `docker-compose run --rm test go test ./cmd/rune/` and confirm it FAILS (function does not exist yet)

### Implementation for User Story 1

- [X] T003 [US1] GREEN: Create `cmd/rune/buildinfo.go` implementing `versionFromBuildInfo` per research D1–D5 (precedence: non-`dev` ldflags version short-circuits; else trusted module version with `v` trimmed; else `"dev"`) plus a thin `resolveVersion(version string) string` wrapper that calls `runtime/debug.ReadBuildInfo()`; doc comments in repo style (see `cmd/rune/versionhook.go` for tone). `docker-compose run --rm test go test ./cmd/rune/` now PASSES
- [X] T004 [US1] Wire the call site in `cmd/rune/main.go`: `newRootCmd(&opts, installedVersion(resolveVersion(version)), commit)` — the runetest-only `RUNE_TEST_VERSION` hook stays outermost (D2) — and update the adjacent comment block to describe the fallback; full suite `docker-compose run --rm test go test ./...` PASSES
- [X] T005 [US1] Validate independently per `specs/019-fix-version-reporting/quickstart.md` V2 pre-release form (`go install …@<pushed-fix-commit>` reports the pseudo-version per C5, commit `none` — NOT `dev`; the clean `X.Y.Z` check runs post-release via `@latest`) and V4 (ldflags-stamped build prints `rune version 9.9.9 (commit abc1234)`; Homebrew 0.4.3 output unchanged) — record results in the PR description

**Checkpoint**: The reported bug is fixed and release output proven unchanged — MVP deliverable.

---

## Phase 4: User Story 2 - Compatibility Gate Works for Go-Toolchain Installs (Priority: P2)

**Goal**: The `minimum_version` gate, `rune version --check`, and `--check --json` consume the true version for go-install binaries: older releases blocked with the standard diagnostic, satisfying ones pass (C7; FR-004, FR-006). Per research D6 this requires ZERO `internal/` changes — these tasks prove the behavior rather than add code.

**Independent Test**: quickstart V5 — a stamped-old binary against a `minimum_version` Runefile aborts with exit 3 and `installed version:` naming the real version; nothing executes.

### Tests for User Story 2 (write first, watch them fail only if behavior is missing)

- [X] T006 [US2] Extend `cmd/rune/buildinfo_test.go` with gate-compatibility cases tying US1 output to `internal/config`: for a resolved module version (`"0.4.3"`) and a resolved pseudo-version, assert `config.MinimumRequirement.Satisfied` compares them as real versions (older-than-required → not ok, dev=false; newer/equal → ok) and that `"dev"` still returns dev=true — expected to pass immediately if T003 output is correct; investigate any failure before touching `internal/`

### Verification for User Story 2

- [X] T007 [US2] Regression: run `docker-compose run --rm test go test ./...` (the `test/integration` harness builds its own `-tags runetest` binary, incl. `minimum_version_test.go`) with ZERO edits to `test/integration/` or `internal/` — confirms C8 (dev bypass) and the diagnostic contract are untouched
- [X] T008 [US2] Validate independently per `specs/019-fix-version-reporting/quickstart.md` V5: stamped binary vs `minimum_version` Runefile → standard mismatch diagnostic, `installed version: <real>`, `nothing was executed`, exit 3; and `rune version --check` / `--check --json` report the same version (C7)

**Checkpoint**: The minimum_version safety net provably engages for the go-install install base.

---

## Phase 5: User Story 3 - Development Builds Stay Honestly Labeled (Priority: P3)

**Goal**: Builds from a source checkout (`go build` / `go run`, clean-at-tag, post-tag, or dirty) keep reporting `dev` and keep bypassing the gate — the fallback must never let a checkout build claim a release version (C6, C8; FR-003). Guard: presence of `vcs.*` build settings (research D3).

**Independent Test**: quickstart V3 — `go run ./cmd/rune --version` and a fresh `go build` binary both print `rune version dev (commit none)`.

### Tests for User Story 3 (write first, watch them fail)

- [X] T009 [US3] RED: Add checkout-build cases to `cmd/rune/buildinfo_test.go`: `Main.Version "v0.4.3"` WITH `vcs.revision`/`vcs.modified` settings (clean-at-tag checkout — the dangerous case, D3) → `"dev"`; pseudo-version `"v0.4.4-0.2026…+dirty"` with vcs settings → `"dev"`; `Main.Version "(devel)"` → `"dev"`. Confirm the clean-at-tag case FAILS if T003 didn't yet guard on vcs settings
- [X] T010 [US3] GREEN: Implement (or confirm) the vcs-settings guard in `cmd/rune/buildinfo.go`: the module version is trusted only when `bi.Settings` contains no `vcs.*` keys (D3); `docker-compose run --rm test go test ./cmd/rune/` PASSES with all US1+US3 cases
- [X] T011 [US3] Validate independently per `specs/019-fix-version-reporting/quickstart.md` V3: `go run ./cmd/rune --version` and `go build`-from-checkout binary both report `dev`, even on a clean tagged commit

**Checkpoint**: All three stories independently verified; contributor workflow unchanged.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T012 [P] Update `docs/installation.md`: note that `go install` binaries report the installed release version (and that source builds report `dev`); keep the docs harness green (`rune docs-check`) — surface S4
- [X] T013 Full quality gates per quickstart V1+V6: `rune fmt`, `rune lint` (golangci-lint zero issues, Principle VIII), `docker-compose run --rm test go test ./...`, `docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...`, `rune docs-check`, `rune release-dryrun` — all green (SC-004)
- [X] T014 Sweep the quickstart scenarios V1–V6 end-to-end once on the finished branch and check off each contract C1–C8 in `specs/019-fix-version-reporting/contracts/version-output.md` against observed output

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: none — start immediately
- **Foundational (Phase 2)**: empty — no blocker
- **US1 (Phase 3)**: after T001. Creates `buildinfo.go`/`buildinfo_test.go`
- **US2 (Phase 4)**: after US1 (consumes the resolved version; no code of its own)
- **US3 (Phase 5)**: after US1 (tightens the same function in `buildinfo.go`); independent of US2
- **Polish (Phase 6)**: T012 anytime; T013–T014 after all stories

### Task Dependencies

- T002 → T003 → T004 → T005 (strict red-green-wire-validate chain)
- T006, T007, T008 after T004 (T006/T007/T008 mutually independent)
- T009 → T010 → T011, with T009 after T003 (same test file as T002's)
- T013 after T004, T010, T012; T014 after T013

### Parallel Opportunities

This feature funnels through two files (`buildinfo.go`, `buildinfo_test.go`), so intra-story parallelism is intentionally near-zero. Genuine parallel lanes:

- T012 (docs) with any of Phases 3–5 — different files, no dependency
- Within US2: T006 (unit cases), T007 (regression run), T008 (manual V5) can run concurrently once T004 lands
- US2 (verification-only) and US3 (code) can proceed in parallel after US1

## Parallel Example: after US1 completes

```bash
# Lane A (US3, code):      T009 → T010 → T011 in cmd/rune/buildinfo*.go
# Lane B (US2, verify):    T006 + T007 + T008 concurrently
# Lane C (docs):           T012 in docs/installation.md
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. T001 baseline → T002 red → T003 green → T004 wire → T005 validate
2. **STOP and VALIDATE**: quickstart V2/V4 prove the bug fixed and releases untouched — shippable on its own

### Incremental Delivery

1. US1 → the reported bug is gone (MVP)
2. US2 → safety-net behavior proven (no code, pure verification)
3. US3 → over-correction guard locked in by tests
4. Polish → docs note + full gate sweep (T013) before PR

### Notes

- Realistically T003 and T010 may land as one function body; keep the RED tasks (T002, T009) honest anyway — write each test set and see it fail (or, for T009 after a complete T003, document that it passed immediately and why that's expected)
- Commit after each task or logical group; never hand-edit goldens (none should change — if one does, that's a C3 violation, stop and investigate)
- Tests always in Docker (`docker-compose`, standalone binary — no compose plugin)
