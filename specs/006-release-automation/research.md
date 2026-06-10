# Phase 0 Research: Release Automation

**Feature**: 006-release-automation · **Date**: 2026-06-10

This document resolves the technical unknowns for an automated, tag-driven release
pipeline for Rune (a pure-Go CLI). The stack is anchored by what the repo already uses —
**GoReleaser v2** (`.goreleaser.yaml`, the `release-dryrun` task) and **GitHub Actions**
(`ci.yml`) — plus a distroless `Dockerfile`. Current tool context verified during
research: **GoReleaser v2.16 (May 2026)**.

Each decision is stated as **Decision / Rationale / Alternatives / Source**. Concrete
config belongs in the `contracts/` artifacts and `tasks.md`; snippets here are
illustrative of the chosen approach.

---

## A. End-to-end release flow (the spine everything hangs off)

A single **`workflow_dispatch`** run on `main` does the whole thing:

1. **Guard** — refuse unless `github.ref == refs/heads/main`, the working tree is clean,
   and CI is green for `HEAD` (query combined check-runs via `gh`). A protected `release`
   GitHub Environment with required reviewers gates *who* can run it.
2. **Compute version** — read the latest `v*` tag, apply the maintainer's `bump`
   (`major`/`minor`/`patch`) input; if `prerelease` is checked, append `-rc.N`. Refuse if
   the resulting tag already exists.
3. **Changelog** — `git cliff --unreleased --tag <next> --prepend CHANGELOG.md`, then
   commit `chore(release): <next>` to `main` (docs-only commit).
4. **Tag** — annotated tag `<next>` on that commit; push it.
5. **Release** — `goreleaser release --clean --release-notes <(git cliff --latest)`:
   builds 6 binaries → archives + `checksums.txt` → SBOMs → cosign-signs checksums →
   builds & pushes the multi-arch GHCR image (+ `latest` if not prerelease) → signs images →
   updates the Homebrew cask + Scoop manifest (skipped for prereleases) → creates the
   GitHub Release (`prerelease: auto`).
6. **Provenance** — `actions/attest-build-provenance` over `checksums.txt`
   (`subject-checksums`) and the image digest.

**Decision**: One dispatch-triggered workflow that tags **and** publishes in the same run.
**Rationale**: A tag pushed using the default `GITHUB_TOKEN` does **not** trigger a second
`on: push: tags` workflow (documented GitHub behavior) — so the common "dispatch pushes
tag → tag-triggered workflow builds" split silently never fires without an extra PAT. One
run also gives a single audit trail and atomic re-runs.
**Alternatives**: Two-workflow split (rejected: token-push won't trigger; needs a PAT just
to fire stage 2). Fully auto release-on-merge (rejected per spec — maintainer picks bump).
**Source**: https://docs.github.com/en/actions/using-workflows/triggering-a-workflow

> **Operational note (branch protection)**: Step 3 commits `CHANGELOG.md` to `main`. If
> `main` is protected, the release identity (a GitHub App install token or a fine-grained
> PAT, see §F) must be allowed to bypass the push restriction, or the changelog commit must
> be routed through an auto-merged PR. Recommended: a GitHub App with bypass. The
> changelog commit is docs-only, so tagging it (rather than the exact CI-tested commit) does
> not change shipped behavior.

---

## B. Versioning & tag computation

**Decision**: Compute the next `vMAJOR.MINOR.PATCH` in-line in the workflow from
`git tag --list 'v*' --sort=-v:refname | head -n1` + the `bump` input; prerelease appends
`-rc.N` (N = count of existing `-rc.*` for that version + 1). Reject an existing tag.
**Rationale**: ~15 lines of shell, no extra dependency, full control. The maintainer owns
the version; tooling only mechanizes it.
**Pre-1.0 (0.x) caveat**: Under SemVer §4, `0.y.z` allows anything to change; by widespread
convention a **breaking change bumps MINOR**, not MAJOR, until 1.0. Because the maintainer
picks the bump, tooling should be **advisory only** — surface the bump git-cliff *would*
suggest from commits and warn on an under-bump (e.g. a `feat!` present but `patch` chosen),
never block or auto-set.
**Alternatives**: `ietf-tools/semver-action` / `cbrgm/semver-bump-action` (fine, but a
dependency for trivial logic); auto-inference (rejected per spec).
**Source**: https://semver.org/ §4 · https://www.conventionalcommits.org/en/v1.0.0/

---

## C. Cross-platform binaries (already largely configured)

**Decision**: Keep the existing `builds:` matrix — `CGO_ENABLED=0`, `-trimpath`,
`-ldflags "-s -w -X main.version={{.Version}} -X main.commit={{.ShortCommit}}"`, goos
`linux/darwin/windows` × goarch `amd64/arm64` (6 targets). Archives stay tar.gz (zip on
Windows — config already uses the v2 `formats:` plural form). `checksums.txt` (sha256).
**Rationale**: Matches Constitution V (static, CGO-free, three OSes) and FR-007/008. The
version vars (`main.version`, `main.commit`) already exist in `cmd/rune/main.go` and feed
the `version` subcommand, so FR-006 is satisfied by the build — no Go changes needed.
**Alternatives**: 32-bit / other arches (out of scope per spec).
**Source**: existing `.goreleaser.yaml`; https://goreleaser.com/customization/build/

---

## D. Multi-arch Docker images

**Decision**: Add `dockers:` (one per goarch, `use: buildx`, reusing the existing
`Dockerfile` with `--platform` + `VERSION`/`COMMIT` build-args) + `docker_manifests:` that
combine `…-amd64` and `…-arm64` into `:{{.Version}}` and a `:latest` manifest. Gate `latest`
with `name_template`/`skip_push` so **prereleases never move `latest`**
(`skip_push: '{{ .Prerelease }}'`). Push to `ghcr.io/rune-task-runner/rune`.
**Rationale**: GoReleaser `dockers` reuses our hand-written distroless Dockerfile (ko builds
its own image and ignores a Dockerfile; raw buildx means hand-scripting tags/manifests).
**Alternatives**: `dockers_v2:` — the future-proof unified block (builds + manifest in one),
**recommended target** since classic `dockers`/`docker_manifests` are deprecation-tracked
(soft-deprecated v2.12, removed in v3). Start on classic to match the existing Dockerfile,
plan a `dockers_v2` migration. `ko` (rejected: ignores Dockerfile). Raw buildx in CI
(rejected: duplicates GoReleaser templating).
**Source**: https://goreleaser.com/customization/docker_manifest/ ·
https://goreleaser.com/customization/dockers_v2/

---

## E. Homebrew + Scoop (⚠️ biggest divergence from old tutorials)

**Decision**: Use **`homebrew_casks:`** (NOT `brews:`) for the Homebrew formula, and
**`scoops:`** for the Scoop manifest, each pointing at an external repo
(`rune-task-runner/homebrew-tap`, `rune-task-runner/scoop-bucket`) via `repository.token`.
Set `skip_upload: auto` on both so prereleases don't update stable channels (FR-020).
**Rationale**: `brews:` is **fully deprecated as of v2.16** (soft-deprecated v2.10).
Homebrew Formulas are designed to build from source; the old `brews:` block abused them for
prebuilt binaries. **Casks** are the supported mechanism for prebuilt binaries and add
shell-completion install + post-install hooks. Most online examples are now obsolete.
**Alternatives**: `brews:` (rejected: deprecated). Submitting to `homebrew-core` / a
community bucket (rejected for now per spec — external review, acceptance bar; revisit later).
**Source**: https://goreleaser.com/customization/homebrew_casks/ ·
https://goreleaser.com/customization/scoop/ · https://goreleaser.com/blog/goreleaser-v2.16/

---

## F. Cross-repo push auth (tap & bucket)

**Decision**: The default `GITHUB_TOKEN` is repo-scoped and **cannot** push to the tap/bucket
repos. Use a **GitHub App installation token** (preferred) or a **fine-grained PAT** scoped
to only those two repos with `contents: write`, stored as a secret (`TAP_GITHUB_TOKEN`), and
referenced from each publisher's `repository.token`.
**Rationale**: App tokens are short-lived, finely scoped, and not tied to a person (survive
staff changes); fine-grained PAT is the simpler fallback. GoReleaser explicitly warns against
broad classic PATs.
**Alternatives**: Classic PAT with `repo` scope (rejected: over-broad). Deploy keys
(rejected: awkward across two repos).
**Source**: https://goreleaser.com/scm/github/ · https://goreleaser.com/customization/homebrew_casks/

---

## G. Supply-chain hardening (signing · SBOM · provenance)

Three complementary layers:

**G1 — Signing (cosign keyless).** Sign `checksums.txt` (covers every artifact
transitively) via `signs:` and sign images/manifests via `docker_signs:`, both keyless.
- **Decision**: `signs: [{cmd: cosign, artifacts: checksum, args: [sign-blob, --bundle=…, ${artifact}, --yes]}]` and a `docker_signs` block signing `${artifact}@${digest}`.
- **Rationale**: Docs state signing the checksum file is generally sufficient. Keyless uses
  Sigstore/Fulcio ephemeral certs bound to the workflow OIDC identity — **no long-lived keys**
  (aligns with Constitution VII, secrets-from-env-only). ⚠️ `signs.artifacts` defaults to
  `none` — must be set explicitly or nothing is signed.
- **Source**: https://goreleaser.com/customization/sign/ · https://goreleaser.com/customization/docker_sign/

**G2 — SBOM (syft).** `sboms: [{artifacts: archive}]` → one SPDX-JSON SBOM per archive.
- **Rationale**: Ties each SBOM to a shipped artifact; SPDX-JSON is GoReleaser's default and
  broadly tooling-compatible.
- **Alternatives**: `artifacts: binary`; CycloneDX (valid — pick one org-wide).
- **Source**: https://goreleaser.com/customization/sbom/

**G3 — Build provenance (GitHub-native).** Use **`actions/attest-build-provenance`**;
attest `checksums.txt` once with `subject-checksums` (covers all binaries) and attest the
image by digest with `push-to-registry: true`.
- **Decision/Rationale (2025-2026 shift)**: GitHub-native attestation is now the default
  recommendation — it emits SLSA-style in-toto provenance, signs keyless via Sigstore, and is
  verifiable with `gh attestation verify`, with far less wiring than reusable-workflow
  generators. It is **complementary** to cosign: cosign asserts *authenticity* of the
  artifacts; provenance records *how/where they were built*.
- **Alternatives**: `slsa-framework/slsa-github-generator` — only path to true **SLSA Build
  L3** (isolated builder); heavier and harder to combine with GoReleaser. Choose only if L3 is
  a hard requirement. Bare `cosign attest` (lower-level, more manual).
- **Source**: https://github.com/actions/attest-build-provenance · https://slsa.dev

---

## H. Changelog: Conventional Commits → committed CHANGELOG.md + release notes

**H1 — PR-title enforcement.** Use **`amannn/action-semantic-pull-request@v6`** to validate
the PR *title* (which becomes the squash commit subject), and enable the repo setting
**"Default to PR title for squash merge commits."** Make the check a **required status** and
trigger on `synchronize` so it re-runs on every push.
- **Rationale**: With squash-merge, the validated title is the only commit on `main` — exactly
  what git-cliff parses. Per-commit linting is unnecessary noise here. (Satisfies FR-013a.)
- **Source**: https://github.com/amannn/action-semantic-pull-request

**H2 — Committed CHANGELOG.md (git-cliff is the single source).** GoReleaser only renders
release-notes text; it does **not** write a committed file. Use **git-cliff** with a
`cliff.toml` mapping `feat`→Added, `fix`→Fixed, `perf`/`refactor`→Changed, etc., and
`--prepend CHANGELOG.md --tag <next>` at release time.
- **Rationale**: git-cliff is stateless and takes the tag as an argument — it never *infers*
  the version, matching "maintainer picks the bump." (Satisfies FR-012/013.)
- **Alternatives**: **release-please** / **semantic-release** (rejected: both *own* versioning
  + tagging + release creation, conflicting with GoReleaser-on-tag). Mirroring grouping in
  GoReleaser `changelog.groups` (rejected: duplicates logic, drifts).
- **Source**: https://git-cliff.org/docs · https://github.com/googleapis/release-please

**H3 — One source, no double changelog.** Set `changelog: { disable: true }` in
`.goreleaser.yaml` and pass `--release-notes <(git cliff --latest)` so the committed section
and the GitHub Release body are byte-identical. (Satisfies FR-014.)
- **Source**: https://goreleaser.com/customization/changelog/ · https://goreleaser.com/customization/release/

**H4 — First entry / pre-convention history.** **Start fresh**: hand-write one curated Keep
a Changelog entry for the current state under the first managed version; configure git-cliff
to parse only commits *after* that baseline tag. Parsing the old free-form history would
yield garbage "Other" groupings.
- **Source**: https://keepachangelog.com/en/1.1.0/ · https://git-cliff.org/docs/usage/examples

**H5 — Prereleases.** `release.prerelease: auto` flags any `-rc/-beta/-alpha` tag as a GitHub
pre-release automatically; give the rc its own changelog section, folded into the GA section
at release. (Satisfies FR-005/020.)
- **Source**: https://goreleaser.com/customization/release/

---

## I. Install script

**Decision**: Hand-write `scripts/install.sh` (POSIX `sh`) that detects OS+arch, maps to the
GoReleaser archive name, downloads it from the GitHub Release, verifies its SHA-256 against
`checksums.txt`, extracts, and installs `rune` to a bin dir (`/usr/local/bin` or
`$HOME/.local/bin`, `INSTALL_DIR` overridable). Document
`curl -sSL https://raw.githubusercontent.com/rune-task-runner/rune/main/scripts/install.sh | sh`.
**Rationale**: GoReleaser does not generate an installer; checksum verification in-script
satisfies FR-019. Lint with shellcheck; smoke-test in CI.
**Alternatives**: Attaching the script as a release asset via `release.extra_files` (can do
both). Third-party installer generators (unnecessary).
**Source**: FR-019; GoReleaser release assets docs.

---

## J. Gating, idempotency, hardening (operational readiness)

- **Gate on green main**: `if: github.ref == refs/heads/main` + a pre-flight step asserting
  all check-runs on `HEAD` succeeded (closes the race where a fresh dispatch beats an
  in-flight CI run) + a protected `release` Environment with required reviewers (FR-026/029).
  Source: https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment
- **Idempotency**: refuse to re-tag an existing version; `goreleaser release --clean` wipes
  `dist/` so re-runs start fresh; GitHub Releases are upserted and GHCR pushes are
  content-addressed (safe to re-push); tap/bucket commits may need manual cleanup on partial
  failure (FR-028).
- **Guards**: refuse a dirty tree (`git status --porcelain`) or non-`main` ref (FR-025).
- **Permissions** (job-scoped, least privilege): `contents: write`, `packages: write`,
  `id-token: write` (cosign keyless + attest OIDC), `attestations: write`. Repo default stays
  `contents: read`.
- **Pinning**: pin all third-party actions to commit SHAs (Dependabot updates); current
  majors are `goreleaser-action@v7`, `setup-go@v5`, `setup-qemu/buildx@v3`, `login-action@v3`,
  `attest-build-provenance@v4`, `action-semantic-pull-request@v6`; checkout needs
  `fetch-depth: 0` + `fetch-tags: true` (changelog + version need full history/tags).
  Source: https://goreleaser.com/ci/actions/

- **Quality gates**: this feature operationalizes Constitution gate #7. Add `goreleaser check`
  + `goreleaser release --snapshot --clean` as a CI job (currently only in the `release-dryrun`
  Runefile task, not in `ci.yml`).

---

## K. Changes to existing files (summary for planning)

| File | Change |
|------|--------|
| `.goreleaser.yaml` | Add `release: {prerelease: auto}`, `dockers:` + `docker_manifests:`, `homebrew_casks:`, `scoops:`, `signs:` (cosign keyless), `docker_signs:`, `sboms:` (syft); change `changelog:` to `{disable: true}`. Keep existing builds/archives/checksums/snapshot. |
| `.github/workflows/` | New `release.yml` (dispatch → version → changelog → tag → goreleaser → attest). New `pr-title.yml` (Conventional-Commit PR-title check). Add a `release-dryrun` job to `ci.yml` (`goreleaser check` + snapshot). |
| `cliff.toml` | New — Conventional-Commit → Keep a Changelog group mapping. |
| `CHANGELOG.md` | New — curated first managed entry; git-cliff prepends thereafter. |
| `scripts/install.sh` | New — OS/arch-detecting, checksum-verifying installer. |
| `Dockerfile` | No structural change; consumed by `dockers:` per-arch builds. |
| `docs/installation.md`, `docs/docker.md` | Add Homebrew/Scoop/install-script/verify instructions. |
| `docs/releasing.md` (or `RELEASING.md`) | New — maintainer runbook + verification. |
| `CONTRIBUTING.md` | Document the Conventional Commits PR-title requirement. |
| `Runefile` | Optionally add `changelog`/`sign-verify` helper tasks; `release-dryrun` already present. |

---

## L. Resolved unknowns

All Technical Context items are resolved — **no `NEEDS CLARIFICATION` remain**. The spec's
deferred items are answered here: prerelease initiating UX (§A/B — `prerelease` boolean
input), failure-notification (surfaced via the Actions run status + the `release` Environment;
no extra channel needed for v1), extra moving image tags (only `:{{.Version}}` and `:latest`
for v1; `v0`/`v0.4` deferred), and authorization (protected `release` Environment, §J).

> **Two items to re-verify against installed tool versions before pasting verbatim** (flagged
> by research): git-cliff's `--bumped-version` flag name and its rc-section `[bump]` keys vary
> across git-cliff releases — confirm with `git cliff --help`. And confirm `homebrew_casks:`
> field names against the exact GoReleaser version pinned in CI (`goreleaser check` will catch
> drift).
