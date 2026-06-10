---
description: "Task list for Release Automation"
---

# Tasks: Release Automation

**Input**: Design documents from `/specs/006-release-automation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: This is release/distribution **infrastructure** (no application unit tests were
requested). Per Constitution Principle VI, each story ends with a **verification task** using
the `goreleaser` dry-run + the verify commands in `quickstart.md` / `contracts/verification.md`.

**Organization**: Tasks are grouped by user story (P1–P5) for independent, incremental delivery.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no incomplete dependency)
- **[Story]**: US1–US5 (user-story phases only; Setup/Foundational/Polish carry no label)
- Exact file paths are included in each task

## ⚠️ Shared-file note (affects [P] markers)

Two files are edited by **multiple** stories and therefore serialize across phases:
`.goreleaser.yaml` and `.github/workflows/release.yml`. Tasks touching them are **not** marked
`[P]` relative to each other. New, distinct files (`cliff.toml`, `CHANGELOG.md`,
`scripts/install.sh`, `pr-title.yml`, docs) are parallelizable.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Provision the accounts, repos, and secrets the pipeline depends on. Mostly
out-of-repo operational steps; can proceed in parallel.

- [ ] T001 [P] Create the public `rune-task-runner/homebrew-tap` repo (empty + README) for the Homebrew cask (research §E, FR-018)
- [ ] T002 [P] Create the public `rune-task-runner/scoop-bucket` repo (empty + README) for the Scoop manifest (research §E, FR-018)
- [ ] T003 [P] Create a GitHub App install token (preferred) or fine-grained PAT scoped to `homebrew-tap` + `scoop-bucket` with `contents: write`, **and** allow it to bypass `main` branch protection (for the changelog commit); store as repo secret `TAP_GITHUB_TOKEN` (research §F, §A operational note)
- [ ] T004 [P] Create a protected `release` GitHub Environment with required reviewers so only authorized maintainers can publish (FR-029, research §J)
- [ ] T005 [P] Enable repo setting **Settings → General → "Default to PR title for squash merge commits"** so the validated PR title becomes the squash commit subject (research §H1, FR-013a)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The release-orchestration skeleton + config validity that every user story plugs
into: a valid GoReleaser baseline, the dispatch workflow up through tag creation, and the CI
dry-run gate.

**⚠️ CRITICAL**: No user-story publishing can occur until this phase is complete.

- [ ] T006 Validate and normalize the existing `.goreleaser.yaml` against GoReleaser v2.16: run `goreleaser check`; confirm `version: 2`, the 6-target `builds:` matrix (`CGO_ENABLED=0`, `-trimpath`, version/commit ldflags), `archives.formats:` (plural), and `checksums.txt` (research §C, §K)
- [ ] T007 Add `release:` block with `prerelease: auto` to `.goreleaser.yaml` so `-rc.N` tags are auto-flagged as GitHub pre-releases (FR-005, research §H5)
- [ ] T008 Create `.github/workflows/release.yml` skeleton: `workflow_dispatch` with inputs `bump` (choice: patch|minor|major, required) and `prerelease` (boolean, default false); `environment: release`; `concurrency` guard; `permissions: { contents: write }`; `actions/checkout` with `fetch-depth: 0` + `fetch-tags: true`; `actions/setup-go@v5` (go 1.25) — SHA-pinned (contracts/release-workflow.md, research §J)
- [ ] T009 Add preconditions to `.github/workflows/release.yml`: refuse unless `github.ref == refs/heads/main`, the tree is clean (`git status --porcelain`), and all check-runs for `HEAD` are `success` (gh query) (FR-025, FR-026, research §A/J)
- [ ] T010 Add version-computation + tagging steps to `.github/workflows/release.yml`: derive `<next>` from the latest `v*` tag + `bump` (+ `-rc.N` if `prerelease`), **refuse if the tag exists**, then create & push an annotated tag (FR-002, FR-003, FR-004, FR-005; contracts/version.md; data-model state machine)
- [ ] T011 Add a `release-dryrun` job to `.github/workflows/ci.yml` running `goreleaser check` + `goreleaser release --snapshot --clean` (operationalizes Constitution gate #7, FR-027; quickstart S1/S2)

**Checkpoint**: The workflow can guard, compute a version, and create a tag; CI validates the
release config and a full snapshot. User stories can now add publishing.

---

## Phase 3: User Story 1 - Cut a versioned, cross-platform binary release (Priority: P1) 🎯 MVP

**Goal**: A maintainer-initiated run turns the new tag into a published GitHub Release with
self-contained binaries for all 6 OS/arch targets + a checksums file.

**Independent Test**: Run the workflow (or `--snapshot` dry-run) → 6 downloadable archives +
`checksums.txt`; a downloaded binary reports the exact released version (quickstart S2/S3; SC-002/006).

- [ ] T012 [US1] Add the GoReleaser publish step to `.github/workflows/release.yml`: `goreleaser/goreleaser-action@v7` (`version: "~> v2"`, `args: release --clean`) with `env: GITHUB_TOKEN` (SHA-pinned) (FR-016, contracts/release-workflow.md)
- [ ] T013 [US1] Confirm `.goreleaser.yaml` archives include `LICENSE*` + `README*` and the Windows `.exe`/zip override, so each of the 6 archives matches the naming contract `rune_<version>_<os>_<arch>.<ext>` (FR-007, FR-009, contracts/artifacts.md)
- [ ] T014 [US1] Verify US1: run `goreleaser release --snapshot --clean`; assert 6 archives + `checksums.txt` in `dist/`, and the built binary prints `rune version <v> (commit <sha>)` (quickstart S2/S3, FR-006, SC-002/006)

**Checkpoint**: MVP — a real run publishes a complete cross-platform binary release. **STOP & VALIDATE.**

---

## Phase 4: User Story 2 - Publish multi-architecture container images (Priority: P2)

**Goal**: The same release builds and pushes one multi-arch GHCR image (linux amd64+arm64),
tagged with the version and (for stable) `latest`.

**Independent Test**: `docker buildx imagetools inspect ghcr.io/rune-task-runner/rune:<version>`
lists both platforms; the image runs and reports the version on each arch (quickstart S7; US2 scenarios).

- [ ] T015 [US2] Add `dockers:` to `.goreleaser.yaml` — one entry per goarch (amd64, arm64), `use: buildx`, reusing `Dockerfile` with `--platform=linux/<arch>` and `--build-arg VERSION=/COMMIT=`; images `ghcr.io/rune-task-runner/rune:{{.Version}}-<arch>` (FR-010, research §D, contracts/artifacts.md)
- [ ] T016 [US2] Add `docker_manifests:` to `.goreleaser.yaml` combining the per-arch images into `:{{.Version}}` and a `:latest` manifest, with `latest` gated off prereleases (`skip_push: '{{ .Prerelease }}'`) (FR-010, FR-020, research §D)
- [ ] T017 [US2] Add to `.github/workflows/release.yml`: `docker/setup-qemu-action@v3`, `docker/setup-buildx-action@v3`, and `docker/login-action@v3` to `ghcr.io` (user `${{ github.actor }}`, password `GITHUB_TOKEN`) — SHA-pinned (research §D/J)
- [ ] T018 [US2] Add `packages: write` to the `permissions:` block in `.github/workflows/release.yml` (FR-017, research §J)
- [ ] T019 [US2] Verify US2: in a `--snapshot` run confirm local `…-amd64`/`…-arm64` images build; document the post-release `imagetools inspect` + dual-arch `docker run --platform` check (quickstart S7, FR-010/017)

**Checkpoint**: US1 + US2 both deliver independently — binaries and a multi-arch image.

---

## Phase 5: User Story 3 - Automated changelog (Priority: P3)

**Goal**: Each release generates a Keep-a-Changelog section in a committed `CHANGELOG.md` and
identical GitHub Release notes, from Conventional-Commit PR titles.

**Independent Test**: After merges, a release adds a dated grouped `CHANGELOG.md` section and
the release notes match it, with no manual editing (quickstart S4; FR-012/014/015).

- [ ] T020 [P] [US3] Create `cliff.toml` mapping Conventional-Commit types to Keep-a-Changelog groups (`feat`→Added, `fix`→Fixed, `perf`/`refactor`→Changed, …), configured to parse only commits after the first managed baseline tag (research §H2/H4, contracts/changelog.md)
- [ ] T021 [P] [US3] Create the curated first-managed `CHANGELOG.md` entry for the current state (Keep a Changelog format); git-cliff manages sections from the next tag forward (research §H4, FR-012)
- [ ] T022 [P] [US3] Create `.github/workflows/pr-title.yml` using `amannn/action-semantic-pull-request@v6` (triggers incl. `synchronize`) to enforce Conventional-Commit PR titles; mark it a required status check (FR-013a, research §H1)
- [ ] T023 [P] [US3] Document the Conventional Commits PR-title requirement in `CONTRIBUTING.md` (FR-013a)
- [ ] T024 [US3] Set `changelog: { disable: true }` in `.goreleaser.yaml` so git-cliff is the single source (research §H3, FR-014)
- [ ] T025 [US3] Insert a changelog step into `.github/workflows/release.yml` **before** the tag step (T010): `git cliff --unreleased --tag <next> --prepend CHANGELOG.md`, commit `chore(release): <next>` to `main` (via `TAP_GITHUB_TOKEN`/App identity that can bypass protection); update the goreleaser step (T012) to pass `--release-notes <(git cliff --latest)` (FR-012/013/014, research §A/H3)
- [ ] T026 [US3] Verify US3: `git cliff --unreleased --tag vX.Y.Z` produces a grouped section, and `git cliff --latest` equals the top `CHANGELOG.md` section; confirm an empty change set warns rather than emitting a broken section (quickstart S4, FR-012/014, edge case)

**Checkpoint**: Releases now carry an automated, single-sourced changelog + matching notes.

---

## Phase 6: User Story 4 - Verifiable, tamper-evident artifacts (Priority: P4)

**Goal**: Every binary archive and image is checksummed, keyless-signed, SBOM'd, and carries
build provenance verifiable with public material only.

**Independent Test**: Verify a release artifact's checksum, cosign signature, and provenance;
a tampered artifact fails verification (quickstart S6; contracts/verification.md; SC-004).

- [ ] T027 [US4] Add `signs:` to `.goreleaser.yaml` — cosign keyless over `checksums.txt` (`artifacts: checksum`, `sign-blob --bundle=${signature} ${artifact} --yes`) ⚠️ default is `none` (FR-022, research §G1)
- [ ] T028 [US4] Add `docker_signs:` to `.goreleaser.yaml` — cosign keyless signing of images/manifests by `${artifact}@${digest}` (FR-022, research §G1)
- [ ] T029 [US4] Add `sboms:` to `.goreleaser.yaml` — syft per-archive SPDX-JSON SBOMs (`artifacts: archive`) (FR-023, research §G2)
- [ ] T030 [US4] Add `id-token: write` + `attestations: write` to the `permissions:` block in `.github/workflows/release.yml` (cosign keyless OIDC + attestation, research §G/J)
- [ ] T031 [US4] Add `actions/attest-build-provenance@v4` steps to `.github/workflows/release.yml` after the goreleaser step: attest `dist/checksums.txt` via `subject-checksums`, and attest the image by digest with `push-to-registry: true` (SHA-pinned) (FR-024, research §G3)
- [ ] T032 [US4] Verify US4: run the checksum + `cosign verify-blob` + `gh attestation verify` flow from `contracts/verification.md`; run the negative tamper test (flip a byte → verification fails) (quickstart S6, SC-004)

**Checkpoint**: All artifacts are independently verifiable; supply-chain hardening complete.

---

## Phase 7: User Story 5 - Frictionless installation (Priority: P5)

**Goal**: Consumers install via Homebrew, Scoop, or a one-line checksum-verifying script;
stable channels stay current automatically and skip prereleases.

**Independent Test**: On each platform, install via at least one channel and confirm the
released version runs; prereleases don't update stable channels (quickstart S5; FR-018/019/020).

- [ ] T033 [US5] Add `homebrew_casks:` to `.goreleaser.yaml` targeting `rune-task-runner/homebrew-tap` with `repository.token: "{{ .Env.TAP_GITHUB_TOKEN }}"` and `skip_upload: auto` ⚠️ NOT `brews:` (deprecated) (FR-018, FR-020, research §E)
- [ ] T034 [US5] Add `scoops:` to `.goreleaser.yaml` targeting `rune-task-runner/scoop-bucket` with `repository.token: "{{ .Env.TAP_GITHUB_TOKEN }}"` and `skip_upload: auto` (FR-018, FR-020, research §E)
- [ ] T035 [US5] Add `TAP_GITHUB_TOKEN` to the env of the goreleaser step in `.github/workflows/release.yml` (research §F)
- [ ] T036 [P] [US5] Create `scripts/install.sh` (POSIX sh): detect OS+arch, map to the archive name, download from the GitHub Release, **verify SHA-256 against `checksums.txt`**, extract, install `rune` to `INSTALL_DIR` (default `/usr/local/bin`) (FR-019, research §I, contracts/artifacts.md)
- [ ] T037 [P] [US5] Lint `scripts/install.sh` with shellcheck and add a smoke-test invocation (e.g. `INSTALL_DIR=/tmp/runebin sh scripts/install.sh` against a published/snapshot release) (FR-019, quickstart S5)
- [ ] T038 [P] [US5] Update `docs/installation.md` with Homebrew (`brew install rune-task-runner/tap/rune`), Scoop, and install-script instructions (FR-030)
- [ ] T039 [US5] Verify US5: confirm `--snapshot` skips tap/bucket upload for a prerelease tag; document brew/scoop install + run the install-script smoke test (quickstart S5, FR-018/020)

**Checkpoint**: All five stories deliver independently; full distribution surface is live.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, runbook, and end-to-end validation spanning all stories.

- [ ] T040 [P] Create `docs/releasing.md` — maintainer runbook: how to cut a release (bump/prerelease inputs), preconditions, and re-run/failure recovery (FR-028, FR-030, contracts/release-workflow.md)
- [ ] T041 [P] Update `docs/docker.md` with GHCR multi-arch pull + image signature/provenance verification (FR-030, contracts/verification.md)
- [ ] T042 [P] Update `README` install section to point at the new channels (Homebrew/Scoop/script/Docker) (FR-030)
- [ ] T043 [P] Add optional helper tasks to `Runefile` (e.g. `changelog`, `verify`) alongside the existing `release-dryrun` task (research §K)
- [ ] T044 Ensure the `docs-verify` harness (`test/docs`) passes for all new/updated docs — links resolve and examples validate (Constitution gate #6)
- [ ] T045 End-to-end validation: cut a real **prerelease** (`prerelease: true`) and walk all of quickstart.md (S2–S8) — confirm `latest`/tap/bucket are NOT updated, then run `goreleaser` idempotency/re-run check (SC-008, SC-009, quickstart S8)
- [ ] T046 [P] Record the deferred follow-ups in `docs/releasing.md`: `dockers_v2` migration (classic dockers removed in GoReleaser v3), optional `v0`/`v0.4` moving image tags, deb/rpm packages, and Apple notarization / Windows Authenticode (research §D, spec Out of Scope)

---

## Dependencies & Execution Order

### Phase dependencies

- **Setup (Phase 1)**: no dependencies; all five tasks are `[P]`.
- **Foundational (Phase 2)**: depends on Setup; **blocks all user stories**. Internal order:
  T006 → T007 (both edit `.goreleaser.yaml`); T008 → T009 → T010 (all edit `release.yml`, ordered);
  T011 is independent (`ci.yml`) and may run any time after T006.
- **User Stories (Phase 3–7)**: each depends only on Foundational and is independently testable.
  Recommended priority order P1 → P2 → P3 → P4 → P5.
- **Polish (Phase 8)**: after the desired stories are complete (T044/T045 require all in-scope stories).

### Cross-story file serialization (important)

- `.goreleaser.yaml` is edited by T006/T007 (Found.), T013 (US1), T015/T016 (US2), T024 (US3),
  T027/T028/T029 (US4), T033/T034 (US5) — these **serialize** (same file).
- `.github/workflows/release.yml` is edited by T008/T009/T010 (Found.), T012 (US1), T017/T018
  (US2), T025 (US3), T030/T031 (US4), T035 (US5) — these **serialize** (same file).
- US3's T025 inserts the changelog step **before** the tag step (T010) and amends the goreleaser
  step (T012), so US3's workflow edit depends on T010 + T012 existing.

### Within each user story

- Config/workflow edits (shared files) precede that story's **verify** task.
- Each story's verify task is the gate to call the story "done."

---

## Parallel Opportunities

- **Phase 1 (Setup)**: T001–T005 all run in parallel (distinct repos/settings).
- **US3**: T020 (`cliff.toml`), T021 (`CHANGELOG.md`), T022 (`pr-title.yml`), T023 (`CONTRIBUTING.md`)
  are `[P]` (distinct new files); T024/T025 (shared files) follow.
- **US5**: T036 (`scripts/install.sh`), T037 (shellcheck), T038 (`docs/installation.md`) are `[P]`
  vs. the `.goreleaser.yaml`/`release.yml` edits (T033–T035).
- **Polish**: T040, T041, T042, T043, T046 are `[P]` (distinct docs/files).
- Across teams: once Foundational is done, US1–US5 can be staffed in parallel **provided** the
  shared-file edits to `.goreleaser.yaml` / `release.yml` are coordinated (rebased/sequenced).

### Parallel example: US3 setup files

```bash
Task: "Create cliff.toml (Conventional-Commit → Keep a Changelog map)"
Task: "Create curated first CHANGELOG.md entry"
Task: "Create .github/workflows/pr-title.yml (PR-title enforcement)"
Task: "Document Conventional Commits in CONTRIBUTING.md"
```

---

## Implementation Strategy

### MVP first (User Story 1 only)

1. Phase 1 Setup (only T003–T005 strictly needed for US1; T001/T002 are US5 prerequisites).
2. Phase 2 Foundational (T006–T011) — **blocks everything**.
3. Phase 3 US1 (T012–T014).
4. **STOP & VALIDATE**: a real run publishes binaries + checksums; `rune --version` matches.

### Incremental delivery

US1 (binaries) → US2 (images) → US3 (changelog) → US4 (signing/SBOM/provenance) → US5
(Homebrew/Scoop/install script). Each is a shippable increment that doesn't break the prior ones.

---

## Notes

- `[P]` = different files, no incomplete dependency. The shared-file note above governs what is
  *not* parallel.
- ⚠️ Two correctness traps surfaced in research: `signs.artifacts` defaults to **`none`** (T027),
  and **`brews:` is deprecated → use `homebrew_casks:`** (T033). `goreleaser check` (T006/T011)
  catches both.
- Re-verify before pasting verbatim: git-cliff `--bumped-version`/rc `[bump]` keys and exact
  `homebrew_casks:` fields vary by version (`git cliff --help`, `goreleaser check`).
- Tests run inside Docker per project policy; the dry-run gate is the primary verification layer.
- Commit after each task or logical group; stop at any checkpoint to validate a story independently.
