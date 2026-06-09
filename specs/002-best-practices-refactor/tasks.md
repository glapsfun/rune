---
description: "Task list for Best-Practices Refactor — Structure, Docs, CI, and Docker"
---

# Tasks: Best-Practices Refactor — Structure, Docs, CI, and Docker

**Input**: Design documents from `/specs/002-best-practices-refactor/`

**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅ (ci-gates, docker-image, release-pipeline, docs-structure), quickstart.md ✅

**Tests**: No NEW unit/TDD tests are generated. This is a **behavior-preserving infrastructure** feature; its correctness oracle is the **existing** test suite, which MUST stay green **with no golden file regenerated** (SC-009). Per global policy and the constitution, the Go suite runs **inside Docker** (`docker compose run --rm test …`), never on the host. Each story therefore carries explicit **verification/validation** tasks (drawn from `quickstart.md`) instead of red-green test tasks.

**Organization**: Tasks are grouped by user story (US1–US5 from spec.md). Priorities: US1 (docs) and US2 (CI) are **P1**; US3 (Docker) and US4 (structure) are **P2**; US5 (release) is **P3**. Suggested MVP = **US1 + US2** (the P1 baseline).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on an incomplete task)
- **[Story]**: US1–US5 (user-story phases only; Setup/Foundational/Polish carry no label)
- Exact file paths are included in every task

## Path Conventions

Single Go module rooted at the repo (per plan.md). This feature touches repo-level tooling/docs (`README.md`, `docs/`, `.github/workflows/`, `Dockerfile`, `.goreleaser.yaml`, `.golangci.yml`, hygiene files, a dogfooded `Runefile`) plus a behavior-preserving format/lint pass over `cmd/`, `internal/`, `mcpserver/`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Lightweight shared baselines (the repo is already scaffolded).

- [X] T001 Audit the repository package layout against the constitution's locked layout (`cmd/rune`, `internal/{token,lexer,ast,parser,analyzer,diag,eval,config,dotenv,cache,cli}`, `internal/runtime/{scheduler,shell,interp,agent}`, public `mcpserver/`); record any deviation in plan notes — verification only, no code change expected (FR-021)
- [X] T002 [P] Add `.editorconfig` at repo root (utf-8, LF, final newline, trailing-whitespace trim; tabs for `*.go`, spaces for `*.{yml,yaml,md,json}`) (FR-022)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Cross-cutting repo hygiene that unblocks the P1 CI gates and the release file references. The format/lint pass must land before US2's strengthened `lint` gate, and `LICENSE` must exist before US2's `release-dryrun` job and US5.

**⚠️ CRITICAL**: Complete this phase before starting US2 or US5.

- [X] T003 Add `gofumpt` and `goimports` formatters to `.golangci.yml` under `formatters.enable` (set `goimports.local-prefixes: github.com/rune-task-runner/rune`), keeping the existing linters and `errcheck` exclusions (FR-008, FR-022)
- [X] T004 Run a repo-wide, behavior-preserving format normalization (`gofumpt -w` + `goimports -w`) across `cmd/`, `internal/`, `mcpserver/`, and fix any `golangci-lint run` findings the stricter config surfaces (FR-022, SC-008) — depends on T003
- [X] T005 Verify behavior preserved after T004: `docker compose run --rm test go test ./...` is green AND `git status --porcelain testdata/` is empty (no golden regenerated) (SC-009, FR-023) — depends on T004
- [X] T006 [P] Add `LICENSE` at repo root so `.goreleaser.yaml` `files: LICENSE*` resolves and US5 archives are complete (FR-024, FR-027)

**Checkpoint**: tree is `gofumpt`/`goimports`/`golangci-lint`-clean with behavior unchanged, and `LICENSE` exists — CI gates and release refs can now be wired.

---

## Phase 3: User Story 1 - Adopt Rune from documentation (Priority: P1) 🎯 MVP

**Goal**: A newcomer installs Rune and runs their first task using only the docs, and every CLI command and Runefile construct is documented with a runnable example.

**Independent Test**: Follow the docs on a clean machine → reach a running task in <10 min (SC-001); every flag/construct has an example (SC-010). See `quickstart.md` §US1.

- [X] T007 [P] [US1] Create a minimal runnable example at `docs/examples/getting-started/Runefile` (a dir-scoped, discoverable `Runefile` with one or two simple `sh` tasks) used by both the quickstart and the freshness check (FR-003)
- [X] T008 [US1] Write `docs/getting-started.md` — zero→first-task quickstart referencing `docs/examples/getting-started/Runefile`, with the exact command and expected output (FR-003, SC-001) — depends on T007
- [X] T009 [P] [US1] Write `docs/installation.md` — install per OS via prebuilt binary, `go install`/build-from-source, and container; link the container section forward to `docs/docker.md` (authored in US3, T026) (FR-002, D1)
- [X] T010 [P] [US1] Write `docs/cli.md` — every command (`mcp`/`serve`, `completion`) and every global flag registered in `cmd/rune/main.go` (`-f/--file`, `--list`, `--dry-run`, `--summary`, `--dump`, `--format`, `--set`, `--watch`, `--choose`, `--yes`, `--quiet`, `--fmt`, `--clear-cache`), each with ≥1 example (FR-004, SC-010)
- [X] T011 [P] [US1] Write `docs/runefile.md` — language usage (task defs, dependencies, parameters/arity, `[cache(inputs,outputs)]`, executors sh/python/node/agent, dotenv, `[confirm]`, `--fmt`), linking `docs/GRAMMAR.md` as the formal grammar (FR-004, SC-010)
- [X] T012 [P] [US1] Write `docs/mcp.md` — how non-private tasks become MCP tools; secure-by-default (read-only default, env-only secrets, destructive `[confirm]` opt-in, remote endpoint opt-in/localhost/token-gated) (FR-005, Principle VII)
- [X] T013 [US1] Write root `README.md` — one-line "what is Rune", who it's for, comparison to make/just, install one-liner, copy-paste quickstart snippet, and links to every `docs/` page; resolves `.goreleaser.yaml` `README*` ref (FR-001) — depends on T008–T012
- [X] T014 [US1] Add a docs-example freshness check that runs `docs/examples/getting-started/Runefile` and asserts its output (used by the dogfooded `Runefile docs-check` and CI) (FR-006) — depends on T007
- [X] T015 [US1] Coverage check: confirm every `cmd/rune/main.go` flag and every `docs/GRAMMAR.md` construct appears in the docs with an example; fill gaps (SC-010) — depends on T010, T011

**Checkpoint**: A newcomer can install Rune and run their first task from the docs alone.

---

## Phase 4: User Story 2 - Constitutional CI gates (Priority: P1)

**Goal**: Every push/PR runs the full constitutional gate set; violations are blocked and the failing gate is named.

**Independent Test**: A deliberately-broken PR (bad format, lint error, failing test, data race, hand-edited golden) is blocked per gate; a clean PR passes (SC-002/003/004). See `quickstart.md` §US2 and `contracts/ci-gates.md`.

- [X] T016 [US2] In `.github/workflows/ci.yml`, confirm the `on:` triggers run on `pull_request` + pushes to `main` (satisfies FR-007's merge-gating intent), then rewrite the `lint` job: `actions/setup-go` with `go-version-file: go.mod`; `gofmt -l .` empty; `golangci-lint run` (now incl. gofumpt+goimports); `go vet ./...` (FR-007, FR-008, FR-014) — depends on T003
- [X] T017 [US2] Update the `test` job in `.github/workflows/ci.yml` to a 3-OS matrix (ubuntu/macos/windows) running `go test -race ./...` with `CGO_ENABLED=1`, keeping the node20 + python3.12 setup for executor integration tests (FR-009, FR-010, SC-004)
- [X] T018 [US2] Add a `build` job to `.github/workflows/ci.yml`: 3-OS matrix `CGO_ENABLED=0 go build ./...` to prove the static artifact compiles everywhere (FR-013) — same file as T016/T017 (sequential)
- [X] T019 [US2] Add a `golden` job to `.github/workflows/ci.yml`: `go test ./internal/diag ./internal/lexer ./internal/parser ./internal/cli ./test/corpus -update` then `git diff --exit-code testdata/ docs/GRAMMAR.md`, failing with a "regenerate deliberately" message (FR-012) — same file (sequential)
- [X] T020 [US2] Keep the `fuzz-smoke` job (FuzzLexer/FuzzParser ~20s each) and align it to `go-version-file: go.mod` in `.github/workflows/ci.yml` (FR-011) — same file (sequential)
- [X] T021 [US2] Add a `release-dryrun` job to `.github/workflows/ci.yml`: `goreleaser release --snapshot --clean` to prove the release config + file refs resolve (SC-006) — depends on T006 (LICENSE) and T013 (README)
- [X] T022 [US2] Confirm each failure clearly names the gate; validate by pushing a deliberately-broken branch (5 violations) and a clean branch. **Manual repo-admin step (not a file)**: enable GitHub branch protection on `main` with these CI jobs as **required status checks** so non-conforming PRs cannot merge — without it SC-003 is unenforceable; document this in CONTRIBUTING.md (T028) (FR-007, FR-015, SC-002, SC-003)

**Checkpoint**: CI enforces the constitution's gate set; non-conforming PRs are blocked.

---

## Phase 5: User Story 3 - Run Rune via Docker (Priority: P2)

**Goal**: Users run Rune install-free via a minimal official image; the test harness is retained and documented.

**Independent Test**: Build/run the image with no Go toolchain → `sh`-task parity with native, image <30 MB, missing-runtime tasks fail cleanly (SC-005, FR-019). See `quickstart.md` §US3 and `contracts/docker-image.md`.

- [X] T023 [P] [US3] Add `.dockerignore` at repo root (exclude `.git`, `dist/`, `.rune/`, `specs/`, `.idea/`, `.vscode/`, `*.test`, `*.out`) to keep the build context minimal/deterministic (FR-017)
- [X] T024 [P] [US3] Add a multi-stage `Dockerfile`: declare `ARG VERSION=dev` / `ARG COMMIT=none`; build stage `golang:1.25-bookworm` with `CGO_ENABLED=0 -trimpath -ldflags "-s -w -X main.version=$VERSION -X main.commit=$COMMIT"`; final stage `gcr.io/distroless/static-debian12:nonroot` with `COPY --from=build /out/rune /rune`, `WORKDIR /work`, `ENTRYPOINT ["/rune"]`, and OCI image labels (FR-016, FR-017, U1)
- [X] T025 [US3] Build and verify: `docker buildx build -t rune:local --load .`; assert image size < 30 MB, `rune --version`/`--list` parity with native, and a python/node/missing-tool task fails with a clear error (FR-019, SC-005) — depends on T024
- [X] T026 [P] [US3] Write `docs/docker.md` — `docker run --rm -v "$PWD":/work ghcr.io/rune-task-runner/rune …` usage, argument passing, and the minimal-image limitations + workarounds; link it from `README.md` and `docs/installation.md` (FR-018, FR-019)
- [X] T027 [P] [US3] Align `Dockerfile.test`'s base image to go.mod's major.minor (`golang:1.25-bookworm`, currently `1.26`) so the harness matches the CI/production Go version; review/improve `Dockerfile.test` + `docker-compose.yml` comments and confirm the in-container test command is the documented path (FR-020, I1)

**Checkpoint**: The official image runs Rune install-free; the test harness is documented.

---

## Phase 6: User Story 4 - Best-practice structure & hygiene (Priority: P2)

**Goal**: Standard hygiene files present, the project dogfoods its own task runner, and layout/discipline conform — all behavior-preserving.

**Independent Test**: Hygiene files exist; `golangci-lint run` reports zero issues; the full suite passes with no golden regenerated; `rune --list` shows the dev tasks (SC-008, SC-009). See `quickstart.md` §US4.

- [X] T028 [P] [US4] Add `CONTRIBUTING.md` — dev setup, the Docker-only test policy (`docker compose run --rm test …`), the CI gate set, how to regenerate goldens deliberately (`-update`), the required **branch-protection** setup (CI jobs as required status checks, per SC-003 / T022), and PR/branch expectations (FR-024)
- [X] T029 [P] [US4] Add `SECURITY.md` — coordinated-disclosure contact/process, restating env-only secrets and secure-by-default agent access (Principle VII) (FR-024)
- [X] T030 [P] [US4] Add a repo-root `Runefile` dogfooding the dev workflow: `fmt` (golangci-lint formatters), `lint` (`golangci-lint run`), `test` (`docker compose run --rm test go test ./...`), `test-race`, `build`, `docker` (buildx the image), `docs-check` (runs the getting-started example), `release-dryrun` (research §9)
- [X] T031 [US4] Wire `docs-check` (T014) into the dogfooded `Runefile` and optionally invoke it from `ci.yml`; verify `rune --list` shows every dev task (FR-006) — depends on T030, T014
- [X] T032 [US4] Final structure/discipline verification: `golangci-lint run` reports zero findings and `docker compose run --rm test go test ./...` is green with `git status --porcelain testdata/` empty (SC-008, SC-009) — depends on T030

**Checkpoint**: Repo hygiene complete; Rune runs its own tasks; zero behavior change.

---

## Phase 7: User Story 5 - One-tag release (Priority: P3)

**Goal**: A single `v*` tag produces cross-platform binaries + checksums + a published GHCR image with no manual assembly.

**Independent Test**: A `--snapshot` dry-run yields the full artifact set with LICENSE+README in each archive; a tag publishes the Release + multi-arch image (SC-007). See `quickstart.md` §US5 and `contracts/release-pipeline.md`.

- [X] T033 [P] [US5] Verify/adjust `.goreleaser.yaml`: `archives.files` include `LICENSE*` + `README*` (now resolvable) and `builds` cover linux/macos/windows × amd64/arm64 (FR-027)
- [X] T034 [US5] Add `.github/workflows/release.yml` triggered on `v*` tags: checkout (full history) + `setup-go` (`go-version-file: go.mod`) + `goreleaser release --clean` producing binaries, checksums, and the GitHub Release; `permissions: contents: write` (FR-025) — depends on T006, T013, T033
- [X] T035 [US5] Add a multi-arch image-publish step to `.github/workflows/release.yml`: GHCR login via `GITHUB_TOKEN` + `docker buildx build --push` to `ghcr.io/rune-task-runner/rune:<version>` and `:latest` for `linux/amd64,linux/arm64`, passing `--build-arg VERSION=<tag> --build-arg COMMIT=<sha>` so the published image reports the release version (consistent with the goreleaser binaries); `permissions: packages: write` (FR-026, U1) — depends on T024, T034 (same file)
- [X] T036 [US5] Validate locally: `goreleaser release --snapshot --clean`; assert `dist/` archives include `LICENSE`+`README` and `checksums.txt` is present (SC-006, SC-007) — depends on T033

**Checkpoint**: One tag yields the complete, reproducible release artifact set.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: End-to-end validation and final consistency across all stories.

- [X] T037 [P] Run the full `quickstart.md` validation (all US scenarios) and check off its Done-when matrix
- [X] T038 [P] Verify `README.md` links resolve to every `docs/` page and that cross-references (installation↔docker, runefile↔GRAMMAR) are correct (SC-010)
- [X] T039 Final green gate: `golangci-lint run` clean + `docker compose run --rm test go test -race ./...` green + `git status --porcelain testdata/` empty (SC-004, SC-008, SC-009)
- [X] T040 [P] Update `CLAUDE.md` / contributor docs to surface the dogfooded `Runefile` dev tasks as the canonical local workflow (maintainability; complements FR-006 and CONTRIBUTING.md)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: no dependencies — start immediately.
- **Foundational (Phase 2)**: depends on Setup. **Blocks US2 (lint/release-dryrun) and US5 (archives).** US1, US3 do not strictly require it (docs/docker don't depend on the Go format pass).
- **US1 (Phase 3, P1)**: can start after Setup. Produces `README.md` (needed by US2's `release-dryrun` job and US5's archives).
- **US2 (Phase 4, P1)**: depends on Foundational (T003 → lint; T006 → release-dryrun) and on US1's `README.md` (T013) for `release-dryrun` (T021).
- **US3 (Phase 5, P2)**: independent — can start after Setup.
- **US4 (Phase 6, P2)**: independent for hygiene files; `docs-check` wiring (T031) depends on US1's T014.
- **US5 (Phase 7, P3)**: depends on Foundational (T006 LICENSE), US1 (T013 README), and US3 (T024 Dockerfile, for the image-publish step).
- **Polish (Phase 8)**: depends on all targeted stories being complete.

### User Story Dependencies

- **US1 (P1)**: depends only on Setup. Headline MVP slice.
- **US2 (P1)**: depends on Foundational + US1's README (for `release-dryrun`).
- **US3 (P2)**: fully independent (after Setup).
- **US4 (P2)**: independent (hygiene); `docs-check` wiring touches US1's example.
- **US5 (P3)**: depends on LICENSE (Foundational), README (US1), Dockerfile (US3).

### Within Each User Story

- US1: `examples/getting-started/Runefile` (T007) before getting-started.md (T008) and docs-check (T014); doc pages before README links them (T013); coverage check last (T015).
- US2: all jobs edit `.github/workflows/ci.yml` → **sequential** (T016→T017→T018→T019→T020→T021→T022), not parallel.
- US3: `.dockerignore` + `Dockerfile` parallel; build/verify after Dockerfile.
- US5: goreleaser verify (T033) before/parallel to release.yml; image-publish step after the Dockerfile and the base release.yml job.

### Parallel Opportunities

- Setup: T002 [P].
- Foundational: T006 [P] (LICENSE) alongside T003 (config); T004/T005 are sequential.
- **US1 docs**: T007, T009, T010, T011, T012 are different files → run [P]; then T008 (needs T007), then T013 (needs the pages), then T014/T015.
- **US3**: T023, T024, T026, T027 are different files → [P]; T025 after T024.
- **US4**: T028, T029, T030 are different files → [P]; T031/T032 after T030.
- **Cross-story**: once Foundational + Setup are done, US1 and US3 and US4's hygiene files can all proceed in parallel by different contributors.
- **US2 is NOT internally parallel** (single workflow file).

---

## Parallel Example: User Story 1

```bash
# After T007 lands, author the independent doc pages together (different files):
Task: "Write docs/installation.md"   # T009
Task: "Write docs/cli.md"            # T010
Task: "Write docs/runefile.md"       # T011
Task: "Write docs/mcp.md"            # T012
# Then T008 (getting-started.md), then T013 (README links them all).
```

## Parallel Example: User Story 3

```bash
# Different files, no inter-dependency:
Task: "Add .dockerignore"            # T023
Task: "Add multi-stage Dockerfile"   # T024
Task: "Write docs/docker.md"         # T026
Task: "Improve Dockerfile.test/compose comments"  # T027
# Then T025 (build + verify) once the Dockerfile exists.
```

---

## Implementation Strategy

### MVP First (P1 baseline = US1 + US2)

1. Phase 1 Setup → Phase 2 Foundational (format/lint clean + LICENSE).
2. Phase 3 US1 (docs) → **STOP & VALIDATE**: a newcomer reaches first task from docs (SC-001). Shippable on its own.
3. Phase 4 US2 (CI gates) → **STOP & VALIDATE**: broken PR blocked, clean PR passes (SC-002/003). P1 baseline complete.

### Incremental Delivery

1. Setup + Foundational → hygiene baseline ready.
2. US1 (docs) → demo: install + first task from docs (MVP!).
3. US2 (CI) → every change now gated.
4. US3 (Docker) → install-free usage.
5. US4 (structure/dogfood) → contributor-ready repo.
6. US5 (release) → one-tag distribution.
7. Polish → full quickstart validation.

### Parallel Team Strategy

After Setup + Foundational:

- Dev A: US1 (docs)
- Dev B: US3 (Docker) + US4 hygiene files
- Dev C: US2 (CI) — starts once US1's README exists (for the release-dryrun job)
- US5 integrates last (needs LICENSE + README + Dockerfile).

---

## Requirements Coverage (traceability)

| Req | Tasks | Req | Tasks |
|-----|-------|-----|-------|
| FR-001 | T013 | FR-015 | T022 |
| FR-002 | T009 | FR-016 | T024 |
| FR-003 | T007,T008 | FR-017 | T023,T024 |
| FR-004 | T010,T011 | FR-018 | T026 |
| FR-005 | T012 | FR-019 | T025,T026 |
| FR-006 | T014,T031 | FR-020 | T027 |
| FR-007 | T016–T022 | FR-021 | T001 |
| FR-008 | T003,T016 | FR-022 | T002,T003,T004 |
| FR-009 | T017 | FR-023 | T005,T032 |
| FR-010 | T017 | FR-024 | T006,T028,T029 |
| FR-011 | T020 | FR-025 | T034 |
| FR-012 | T019 | FR-026 | T035 |
| FR-013 | T018 | FR-027 | T033,T036 |
| FR-014 | T016 | | |
| SC-001 | T008,T013,T037 | SC-006 | T021,T036 |
| SC-002 | T016–T022 | SC-007 | T034,T035,T036 |
| SC-003 | T022,T028 | SC-008 | T004,T032,T039 |
| SC-004 | T017,T039 | SC-009 | T005,T032,T039 |
| SC-005 | T025 | SC-010 | T010,T011,T015,T038 |

---

## Notes

- [P] = different files, no dependency on an incomplete task. The US2 CI jobs all edit one file → sequential.
- This feature changes **no Rune behavior**; "done" for the refactor means the existing suite stays green with **no golden regenerated** (SC-009).
- All Go test runs go through Docker (`docker compose run --rm test …`) per global policy.
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
- Avoid: editing the same file in "parallel", regenerating goldens to mask behavior change, embedding secrets in CI/images/docs/release config (Principle VII).
