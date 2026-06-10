# Implementation Plan: Release Automation

**Branch**: `006-release-automation` | **Date**: 2026-06-10 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/006-release-automation/spec.md`

## Summary

Stand up an automated, tag-driven release pipeline for Rune (a pure-Go CLI). A maintainer
runs one GitHub Actions workflow and picks a bump (major/minor/patch, optionally a
prerelease); the automation computes the next `vX.Y.Z` tag, prepends a Conventional-Commit
changelog to `CHANGELOG.md`, creates the tag, and runs GoReleaser to build six binaries
(Linux/macOS/Windows × amd64/arm64) plus a multi-arch GHCR image, publish a GitHub Release
with checksums, sign artifacts and images keyless (cosign) with SBOMs and build provenance,
and update a Homebrew cask + Scoop manifest in dedicated tap/bucket repos. A curl|sh install
script and refreshed install/verify docs round out distribution. The approach is anchored by
the repo's existing GoReleaser v2 config + distroless Dockerfile; the work is almost entirely
declarative (YAML/TOML/CI) plus one POSIX install script — **no engine/Go changes**. See
[research.md](./research.md) for the verified, current (GoReleaser v2.16) technical decisions.

## Technical Context

**Language/Version**: Go 1.25 (unchanged; release embeds version via existing
`-X main.version`/`-X main.commit` ldflags — `cmd/rune/main.go` already declares these and
wires the `version` subcommand). Feature deliverables are YAML (GoReleaser, GitHub Actions),
TOML (git-cliff), and POSIX `sh` (install script).

**Primary Dependencies**: GoReleaser `~> v2` (≥ v2.16 for `homebrew_casks`); cosign
(keyless/Sigstore); syft (SBOM); git-cliff (changelog); Docker buildx + QEMU; GitHub Actions —
`actions/checkout@v4`, `actions/setup-go@v5`, `docker/setup-qemu-action@v3`,
`docker/setup-buildx-action@v3`, `docker/login-action@v3`, `goreleaser/goreleaser-action@v7`,
`actions/attest-build-provenance@v4`, `amannn/action-semantic-pull-request@v6` (all SHA-pinned).

**Storage**: N/A (no datastore). State lives in git tags, `CHANGELOG.md`, GitHub Releases,
and the GHCR registry.

**Testing**: `goreleaser check` (config validity) + `goreleaser release --snapshot --clean`
(full dry-run, no publish — Constitution gate #7) run in Docker per project policy; shellcheck
+ a smoke test for `scripts/install.sh`; existing `docs-verify` harness for updated docs.

**Target Platform**: Release artifacts target Linux/macOS/Windows on amd64+arm64 and a Linux
multi-arch (amd64/arm64) container; the pipeline runs on a GitHub-hosted `ubuntu-latest`
runner (arm64 images built via QEMU emulation).

**Project Type**: Single project — release/distribution infrastructure for an existing CLI.
No source-tree restructuring.

**Performance Goals**: Maintainer effort = one action (SC-001); end-to-end pipeline target
well under ~15 min wall-clock. Not latency-sensitive.

**Constraints**: `CGO_ENABLED=0` static binaries (Constitution V); **keyless** signing — no
long-lived private keys (Constitution VII); secrets only from CI env (`GITHUB_TOKEN`,
`TAP_GITHUB_TOKEN`); releases only from a clean `main` after green CI; idempotent / re-runnable;
prereleases must not move `latest` or update stable Homebrew/Scoop channels.

**Scale/Scope**: 6 binary targets + 2 container arches + 4 distribution channels (Releases,
GHCR, Homebrew, Scoop) + supply-chain artifacts (checksums, signatures, SBOMs, provenance);
low, manual release cadence.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Evaluated against `.specify/memory/constitution.md` v1.0.0 (principles I–VIII + gates).

| Principle | Verdict | Notes |
|-----------|---------|-------|
| I. Command Runner, Not a Build System | ✅ PASS | No change to task-execution/caching semantics. |
| II. Errors Are a Feature | ✅ PASS | Pipeline fails fast with clear diagnostics (tag-exists, dirty-tree, CI-not-green, `goreleaser check`); FR-028 mandates clear, re-runnable failures. |
| III. Minimal, Total DSL | ✅ PASS | No DSL/expression-language change. |
| IV. Hand-Written Front End, Idiomatic Go | ✅ PASS | No parser/engine/`internal/` change; package layout untouched. |
| V. Boringly Portable | ✅ PASS (core enabler) | Ships the static, CGO-free single binary on all three OSes — this feature *is* the portable-distribution mechanism. Keeps `CGO_ENABLED=0`. |
| VI. Test-First, Multi-Layer Verification | ✅ PASS | Adds `goreleaser check` + `--snapshot` dry-run as a CI gate (operationalizes gate #7); install script linted + smoke-tested; docs changes go through `docs-verify`. |
| VII. AI-Native, Secure by Default | ✅ PASS | Keyless signing (no stored keys); secrets only from CI env, never a Runefile; distroless nonroot image; provenance/attestation strengthen the agent-facing trust surface. |
| VIII. Go Engineering Discipline | ✅ PASS | Minimal/no new Go; lint stays clean. New shell/YAML held to shellcheck/`goreleaser check`. |

**Engineering Constraints**: Docker-only testing preserved (dry-run runs in the test harness);
locked package layout untouched; no DSL backward-compat impact; surface/docs ship together
(install + releasing docs in-scope). Adopting Conventional Commits is a contributor-process
change documented in `CONTRIBUTING.md`, not a code-compat change.

**Quality Gates**: Directly advances gate #7 (`release-dryrun`) by wiring it into `ci.yml`.

**Result**: **PASS — no violations.** Complexity Tracking is empty.

## Project Structure

### Documentation (this feature)

```text
specs/006-release-automation/
├── plan.md              # This file
├── research.md          # Phase 0 — verified release best practices (GoReleaser v2.16)
├── data-model.md        # Phase 1 — release entities, version state machine, artifact schema
├── quickstart.md        # Phase 1 — dry-run, cut-a-release, and verify validation guide
├── contracts/           # Phase 1 — release interface contracts
│   ├── release-workflow.md   # workflow_dispatch inputs + run contract
│   ├── artifacts.md          # published artifact naming/layout contract
│   ├── version.md            # version string + tag/semver contract
│   ├── changelog.md          # CHANGELOG.md / release-notes contract
│   └── verification.md       # checksum/signature/provenance/SBOM verify contract
└── tasks.md             # Phase 2 — created by /speckit-tasks (NOT here)
```

### Source Code (repository root)

This feature touches release/distribution configuration, not the Go engine. Affected paths:

```text
.goreleaser.yaml                 # EXTEND: release{prerelease:auto}, dockers + docker_manifests,
                                 #   homebrew_casks, scoops, signs (cosign keyless), docker_signs,
                                 #   sboms (syft); changelog -> {disable: true}
cliff.toml                       # NEW: Conventional-Commit -> Keep a Changelog group mapping
CHANGELOG.md                     # NEW: curated first managed entry; git-cliff prepends after
scripts/
└── install.sh                   # NEW: OS/arch-detecting, checksum-verifying curl|sh installer
.github/workflows/
├── ci.yml                       # EXTEND: add release-dryrun job (goreleaser check + snapshot)
├── release.yml                  # NEW: dispatch -> version -> changelog -> tag -> goreleaser -> attest
└── pr-title.yml                 # NEW: Conventional-Commit PR-title enforcement (required check)
Dockerfile                       # UNCHANGED structurally; consumed by per-arch docker builds
docs/
├── installation.md              # EXTEND: Homebrew/Scoop/install-script + verification
├── docker.md                    # EXTEND: GHCR multi-arch pull + image verification
└── releasing.md                 # NEW: maintainer release runbook
CONTRIBUTING.md                  # EXTEND: Conventional Commits PR-title requirement
Runefile                         # OPTIONAL: changelog/verify helper tasks (release-dryrun exists)
```

**Structure Decision**: Single-project, infrastructure-only change. No `src/`/`internal/`
restructuring; the "implementation" is declarative config + CI workflows + one shell script +
docs, all at the repo root and under `.github/`, `docs/`, and a new `scripts/`. This keeps
the locked engine package layout (Constitution IV) untouched.

## Complexity Tracking

> No constitution violations — table intentionally empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|--------------------------------------|
| — | — | — |
