# Tasks: VS Code Marketplace Extension for the Rune LSP

**Input**: Design documents from `/specs/015-vscode-lsp-extension/`

**Prerequisites**: plan.md, spec.md, research.md (D1–D10), data-model.md (S1–S13), contracts/ (publishing-pipeline.md P1–P8, marketplace-listing.md L1–L9), quickstart.md (V1–V6)

**Tests**: The CI `extension` smoke job (P8) is this feature's test layer and is written **first** within US1 (test-first: it stays red until the listing assets land). No Go tests are touched.

**Organization**: Tasks are grouped by user story. US1 delivers a Marketplace-grade, installable extension package; US2 wires publishing into the release pipeline; US3 adds Open VSX.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)

## Path Conventions

Monorepo subtree per plan.md: extension in `editors/vscode/`, pipelines in `.github/workflows/`, docs in `docs/` and README files. No `src/`/`tests/` — no Go code changes in this feature.

---

## Phase 1: Setup (Packaging Infrastructure)

**Purpose**: Make the extension reproducibly packageable before any listing or pipeline work.

- [x] T001 Prepare `editors/vscode/package.json` for tooling: set the `version` field to the `0.0.0` placeholder (stamped from the tag at publish time, never committed — research D4), pin `@vscode/vsce` and add `ovsx` as devDependencies, keep the `package` script and add `publish:vsce` (`vsce publish --packagePath`) and `publish:ovsx` (`ovsx publish`) scripts; then run `npm install` in `editors/vscode/` to generate and commit `editors/vscode/package-lock.json` for reproducible `npm ci` (S1, S2, research D3)
- [x] T002 [P] Create `editors/vscode/.vscodeignore` so the `.vsix` stays lean: exclude `package-lock.json`, `.vscodeignore` itself, and any dev/test artifacts; MUST NOT exclude `extension.js`, `syntaxes/`, `language-configuration.json`, `README.md`, `CHANGELOG.md`, `LICENSE`, `icon.png`, or production `node_modules` (S8, D3)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: None beyond Setup — this feature touches a single self-contained subtree plus two workflow files; there is no shared foundation layer. User story phases can begin immediately after Phase 1.

**Checkpoint**: `npm ci && npx vsce package` succeeds locally in `editors/vscode/`.

---

## Phase 3: User Story 1 — Install Rune language support from the Marketplace (Priority: P1) 🎯 MVP

**Goal**: The extension package is Marketplace-grade: correct identity metadata, icon, license, listing README, actionable missing-binary UX — everything needed for a trustworthy one-click install (FR-001–FR-004, L1–L9).

**Independent Test**: Quickstart V1–V3 — package the `.vsix`, verify its contents, side-load it into VS Code, confirm language features on a Runefile, and confirm the missing/old-binary notification (no repo interaction beyond installing the `.vsix`; the Marketplace-search flow itself becomes verifiable after the first US2 publish, quickstart V6).

### Tests for User Story 1 (test-first — written before the assets exist) ⚠️

- [x] T003 [US1] Add the `extension` smoke job to `.github/workflows/ci.yml`: Node 22 → `npm ci` in `editors/vscode/` → stamp a throwaway version (`npm version --no-git-tag-version 0.0.1` — MUST be plain `major.minor.patch`; vsce rejects prerelease suffixes like `0.0.1-ci`) → `npx vsce package` → assert (case-insensitively — vsce normalizes doc filenames to `readme.md`/`changelog.md`/`LICENSE.txt`) the `.vsix` contains `extension/extension.js`, `extension/syntaxes/runefile.tmLanguage.json`, `extension/language-configuration.json`, `extension/readme.md`, `extension/changelog.md`, `extension/LICENSE`, `extension/icon.png`, and `extension/node_modules/vscode-languageclient/` (contract P8; S9). NOTE: red→green cycle — write the job and its assertions first and verify the failure LOCALLY (run the same commands in `editors/vscode/`); land T003 together with T004–T007 in the same push so pushed commits keep CI green (constitution Quality Gates).

### Implementation for User Story 1

- [x] T004 [P] [US1] Create `editors/vscode/icon.png` — 128×128 PNG listing icon, generated with a throwaway script (e.g. Python Pillow or ImageMagick in the scratchpad: the ᚱ rune glyph, light on a dark rounded-square background; simple and legible at 42px tile size; no trademarked material). Commit ONLY the resulting `icon.png` — the generator script is not part of the repo. Reference it from the manifest in T007 (S7, L3, D10)
- [x] T005 [P] [US1] Copy the repository root `LICENSE` (MIT) to `editors/vscode/LICENSE` so vsce packages it and the listing shows a license (S6, L5)
- [x] T006 [P] [US1] Create `editors/vscode/CHANGELOG.md` — thin file directing readers to the repository `CHANGELOG.md` / GitHub releases; rendered on the listing's Changelog tab (S5, L7)
- [x] T007 [US1] Harden listing metadata in `editors/vscode/package.json`: fix `repository.url` from `https://github.com/rune-task-runner/rune` to `https://github.com/glapsfun/rune`, add `icon: "icon.png"`, `bugs` and `homepage` URLs (both under `github.com/glapsfun/rune`), and `keywords` (`rune`, `runefile`, `task runner`, `lsp`, `language server`); keep existing `name`/`displayName`/`publisher`/`categories`/`license` (S1, L1–L6, D10; sequential after T001 — same file)
- [x] T008 [US1] Add the missing/old-binary preflight to `editors/vscode/extension.js`: before constructing the `LanguageClient`, resolve the executable from `rune.path` (default `rune`) and probe it by spawning `<executable> lsp --help` with a 5-second timeout; on spawn error (ENOENT — binary missing) show one `window.showErrorMessage` ("Rune binary not found…") with buttons **Install instructions** (opens the docs installation page URL) and **Open Settings** (opens `rune.path`); on non-zero exit or timeout (binary present but no `lsp` subcommand — pre-LSP install) show the "please upgrade Rune" variant with the same buttons; in both cases do NOT start the client (no crash-restart loop); on success (exit 0) start the client exactly as today (S3, FR-003, L9, D5)
- [x] T009 [US1] Rewrite `editors/vscode/README.md` as the Marketplace listing page: state the `rune` binary prerequisite FIRST with a link to the canonical install docs, then the feature list (diagnostics with `RUNE####` codes, completion, go-to-definition, hover, outline, formatting), the settings table (`rune.path`, `rune.trace.server`), and demote the build-from-source instructions to a short contributor note at the bottom (S4, L4, SC-006)
- [x] T010 [P] [US1] Update `editors/README.md` VS Code section: installing from the VS Code Marketplace is the primary path (extension ID `rune-task-runner.rune`), side-loading the release-asset `.vsix` is the offline fallback, building from source is the contributor path (S12, SC-006)
- [x] T011 [P] [US1] Update the root `README.md` editor-support bullet to mention the extension is installable from the VS Code Marketplace and link the listing / `editors/README.md` (S13)
- [ ] T012 [US1] Validate quickstart V1–V3: V1 package + `unzip -l` content check matches contract P8; V2 side-load into VS Code and confirm diagnostics/completion/hover/outline/format on a Runefile with a known error; V3 missing-binary notification (bogus `rune.path`), old-binary "upgrade" notification (stub script), and no activation in a Runefile-less workspace; confirm the T003 CI job is now GREEN

**Checkpoint**: A listing-grade `.vsix` builds and installs cleanly; missing-binary UX verified; CI gate green. Only the actual registry upload (US2) remains for the public install flow.

---

## Phase 4: User Story 2 — Extension stays current with each Rune release (Priority: P2)

**Goal**: Every stable release automatically publishes a matching extension version to the Marketplace; failures are visible and recoverable without re-releasing (FR-005–FR-008, P1–P7).

**Independent Test**: Quickstart V5 — inspect/dry-inspect `release.yml`: the core `release` job is byte-identical, `publish-extension` is stable-gated and secrets appear only as environment references; recovery procedure documented. Full live proof is the first stable release (quickstart V6, post-merge).

### Implementation for User Story 2

- [x] T013 [US2] Add the `publish-extension` job to `.github/workflows/release.yml`. The existing `release` job's STEPS stay byte-identical; the only touch to it is a job-level `outputs:` mapping exposing the computed tag (`tag: ${{ steps.ver.outputs.tag }}`) per amended contract P1. New job: `needs: release`, `if: ${{ !inputs.prerelease }}`, `environment: release` (NOTE: with required reviewers this means a SECOND approval after the core release — accepted trade-off, documented in T014); steps: checkout `needs.release.outputs.tag` → setup Node 22 → `npm ci` in `editors/vscode/` → stamp the version with `npm version --no-git-tag-version "${TAG#v}"` (P3, D4) → `npx vsce package` → upload `rune-<ver>.vsix` to the GitHub release via `gh release upload` with `GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}` (P4c, D9) → `npx vsce publish --packagePath` with `VSCE_PAT` env from secrets (P4d, P5) — Marketplace publish and (later, T016) Open VSX publish as separate steps for partial-failure visibility (P7)
- [x] T014 [US2] Update `docs/releasing.md` with the extension-publishing section: one-time Marketplace publisher setup runbook (create `rune-task-runner` publisher at marketplace.visualstudio.com/manage, mint an Azure DevOps PAT with Marketplace→Manage scope for all accessible orgs, store as `VSCE_PAT` secret on the protected `release` environment — D2), per-release behavior (stable-only, version == tag minus `v`, and the second required-reviewer approval on the `publish-extension` job — the protected `release` environment gates each job separately; approving it is expected and NOT a manual publish step), and failure recovery (re-run failed jobs is idempotent thanks to duplicate-version rejection; manual fallback: download the release-asset `.vsix` and `vsce publish --packagePath` locally; PAT expiry/rotation steps — D8, FR-008, S11)
- [x] T015 [US2] Validate quickstart V4 + V5: V4 — break the manifest on a scratch branch and confirm the `extension` CI job goes red, revert to green; V5 — `git diff main -- .github/workflows/release.yml` shows only the added `publish-extension` job plus the job-level `outputs:` mapping on `release` (P1), the `!inputs.prerelease` gate and `needs: release` are present (P2), and `VSCE_PAT` appears only as a `secrets.*` reference (P5)

**Checkpoint**: A stable release dispatch would publish the extension end-to-end; an rc dispatch skips the job; the core release path is provably untouched.

---

## Phase 5: User Story 3 — Install on VS Code-compatible editors via Open VSX (Priority: P3)

**Goal**: The same `.vsix`, same version, published to Open VSX so VSCodium-class editors can install it (FR-009, P4e, P7).

**Independent Test**: Workflow inspection (ovsx step present, separate from the Marketplace step, consumes the same `--packagePath` artifact); after the first live publish, quickstart V6 step 3 (open-vsx.org listing + VSCodium install).

### Implementation for User Story 3

- [x] T016 [US3] Append the Open VSX publish step to the `publish-extension` job in `.github/workflows/release.yml`: `npx ovsx publish` the SAME `.vsix` file (P4e), as a separate step after the Marketplace publish with `OVSX_PAT` env from secrets, so a one-registry failure is identifiable and does not affect the other (P5, P7; sequential after T013 — same job block)
- [x] T017 [P] [US3] Extend the `docs/releasing.md` runbook with Open VSX one-time setup: Eclipse Foundation account, publisher agreement, `npx ovsx create-namespace rune-task-runner`, access-token creation, `OVSX_PAT` secret on the `release` environment, and the note that Open VSX publish failure is recovered the same way (re-run job / manual `ovsx publish` of the release-asset `.vsix`) (D2, D8, S11)
- [x] T018 [P] [US3] Mention Open VSX in the install docs: `editors/README.md` (VSCodium/compatible editors install from Open VSX, same extension ID) and a one-line note in `editors/vscode/README.md` (S12, S4)

**Checkpoint**: All three stories implemented; both registries served by one artifact.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T019 [P] Contract sweep: re-read `contracts/publishing-pipeline.md` (P1–P8) and `contracts/marketplace-listing.md` (L1–L9) against the final diff; fix any drift (e.g. a forgotten manifest field or step ordering in the workflow)
- [ ] T020 Run the full local quickstart pass (V1–V5) top to bottom on the finished branch; record V6 (post-first-release verification: Marketplace search, Open VSX listing, `.vsix` release asset, idempotent re-run) in the PR description as the post-release checklist
- [x] T021 Run the repo gates that this feature can affect: the docs-verify harness via the Docker test harness (`docker-compose run --rm test go test ./test/docs/...` — never on the host, per the Docker-only testing policy; new/changed docs links must resolve) and confirm the full CI matrix including the new `extension` job is green on the PR

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: no dependencies.
- **Foundational (Phase 2)**: empty — US phases start right after Setup.
- **US1 (Phase 3)**: depends on Setup (T001 lockfile enables `npm ci` in T003's CI job).
- **US2 (Phase 4)**: depends on Setup; independent of US1 code-wise, but a live publish only makes sense once US1's listing quality is in — implement in priority order.
- **US3 (Phase 5)**: T016 depends on T013 (extends the same job); T017 extends T014's runbook file.
- **Polish (Phase 6)**: after all desired stories.

### Task-Level Notes

- T001 → T007 (same `package.json`, sequential).
- T003 written before T004–T009 and expected red until they land (test-first within one PR).
- T013 → T016 (same workflow job). T014 → T017 (same doc file).
- All [P] tasks touch distinct files with no incomplete-task dependencies.

### Parallel Opportunities

- Phase 1: T002 parallel with T001.
- US1: T004, T005, T006 in parallel after T003; T010, T011 in parallel with T008/T009.
- US3: T017, T018 in parallel after T016.

## Parallel Example: User Story 1

```bash
# After T003 (CI job, red), launch the asset tasks together:
Task: "Create editors/vscode/icon.png (128×128 listing icon)"
Task: "Copy root LICENSE to editors/vscode/LICENSE"
Task: "Create editors/vscode/CHANGELOG.md pointing at the repo changelog"

# While T008 (preflight) is in progress:
Task: "Update editors/README.md — Marketplace install becomes primary"
Task: "Update root README.md editor-support bullet"
```

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 (Setup) → Phase 3 (US1).
2. **STOP and VALIDATE**: quickstart V1–V3 + green CI `extension` job.
3. This already yields a shippable `.vsix` (side-load/offline path) even before registry publishing exists.

### Incremental Delivery

1. US1 → listing-grade package, validated locally (MVP).
2. US2 → stable releases publish to the Marketplace automatically; first real release completes SC-001–SC-004 (verify via quickstart V6).
3. US3 → same artifact reaches Open VSX users.
4. Polish → contract sweep + gate runs.

### Notes

- The one-time publisher/namespace creation (runbook in T014/T017) is a **human maintainer action** outside CI — it must happen before the first stable release after merge, or the `publish-extension` job will fail on auth (recovery: runbook + re-run, no re-release needed).
- Commit after each task or logical group; the whole feature is one PR with the red→green CI story visible in its history.
- No Go code, no `docs/GRAMMAR.md`, no golden files are touched; Docker-only testing rule is unaffected (no `go test` involvement).
