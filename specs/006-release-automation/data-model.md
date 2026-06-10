# Phase 1 Data Model: Release Automation

**Feature**: 006-release-automation · **Date**: 2026-06-10

This feature has no application datastore. The "data model" is the set of **release-domain
entities** the pipeline produces and the **version/release state machine** that governs them.
Fields and validation rules trace to the spec's Functional Requirements (FR-xxx).

---

## Entity: Release

An immutable, versioned publication tied to one tag.

| Field | Type | Rules / Source |
|-------|------|----------------|
| `version` | semver string | `vMAJOR.MINOR.PATCH[-rc.N]`; unique; immutable (FR-001, FR-004) |
| `commit` | git SHA | The tagged commit (the docs-only changelog commit); embedded in artifacts (FR-006) |
| `is_prerelease` | bool | True iff version has a `-rc/-beta/-alpha` suffix; drives channel gating (FR-005, FR-020) |
| `created_at` | date | Release date; appears in the changelog section (FR-012) |
| `artifacts` | Artifact[] | All deliverables (FR-007–011, 021–024) |
| `changelog_section` | Changelog entry | The notes for this version (FR-012, FR-014) |
| `release_notes_url` | URL | GitHub Release page (FR-016) |

**Lifecycle**: created → published (immutable). A partial/failed publish is **re-runnable to
completion** without duplicate tags (FR-028). Releases are never edited in place except by a
safe re-run that converges to the same version.

---

## Entity: Version tag

| Field | Type | Rules / Source |
|-------|------|----------------|
| `name` | string | `v` + SemVer 2.0.0; e.g. `v0.0.1`, `v0.4.0-rc.1` (FR-001, FR-005) |
| `target_commit` | git SHA | Annotated tag; created by automation, not by hand (FR-003) |
| `bump_from` | enum | `major`\|`minor`\|`patch` chosen by maintainer (FR-002) |

**State machine (version computation)** — input = latest tag + `bump` + `prerelease` flag:

```text
latest = max(tags matching v*)  (default v0.0.0 if none)         [FR-002]
         │
         ├─ major → vX+1.0.0
         ├─ minor → vX.Y+1.0
         └─ patch → vX.Y.Z+1
         │
         ├─ prerelease=false → <next>                            [stable]
         └─ prerelease=true  → <next>-rc.{count(<next>-rc.*)+1}  [FR-005]
         │
         ▼
   if tag exists → REFUSE (no duplicate)                          [FR-004]
   advisory: warn if chosen bump < bump implied by commits        [research §B]
   pre-1.0 (0.x): breaking change ⇒ MINOR, not MAJOR              [research §B]
```

**Guards before any version is cut**: ref == `main`, clean tree, CI green for HEAD
(FR-025, FR-026); only authorized maintainers (protected `release` Environment, FR-029).

---

## Entity: Artifact

One distributable deliverable. All carry the release version (FR-011) and are covered by
checksums/signature/provenance (FR-021–024).

| Kind | Produced for | Notes |
|------|-------------|-------|
| Binary archive | 6 targets: {linux,darwin,windows} × {amd64,arm64} | `tar.gz` (zip on Windows), includes LICENSE+README (FR-007, FR-009) |
| `checksums.txt` | whole release | SHA-256 over every archive (FR-021) |
| Container image | linux/amd64 + linux/arm64 | Single multi-arch manifest `:{{version}}` (+ `:latest` if stable) on GHCR (FR-010) |
| Signature bundle | `checksums.txt` + images | cosign keyless; verifiable with public material only (FR-022) |
| SBOM | each archive | syft, SPDX-JSON (FR-023) |
| Provenance attestation | `checksums.txt` + image digest | GitHub `attest-build-provenance`, SLSA-style (FR-024) |

**Naming contract** is GoReleaser's defaults (`rune_<version>_<os>_<arch>.<ext>`); pinned in
[contracts/artifacts.md](./contracts/artifacts.md) so the install script and docs can rely on it.

**Atomicity**: if any required target fails to build, the release fails as a whole — no
partial publish of binaries (FR + edge case "one platform fails to build").

---

## Entity: Changelog

The accumulating, human-readable record (Keep a Changelog format), single-sourced by git-cliff.

| Field | Type | Rules / Source |
|-------|------|----------------|
| file | `CHANGELOG.md` | Committed; one dated `## [version] - YYYY-MM-DD` section per release (FR-012) |
| groups | enum sections | Added/Changed/Fixed/… mapped from Conventional-Commit types (FR-012, FR-013) |
| source | PR-title commits | Conventional-Commit squash-merge titles since previous tag (FR-013, FR-013a) |
| links | compare URLs | Each version links to its tag and a diff vs the previous version (FR-015) |
| release_notes | rendered text | `git cliff --latest` → identical to the GitHub Release body (FR-014) |

**First entry**: curated by hand for current state; git-cliff manages from the next tag
forward (research §H4). **Empty change set**: warn the maintainer; do not emit a broken/empty
section (edge case).

---

## Entity: Distribution channel

A path through which consumers obtain a release.

| Channel | Target | Stable-only? | Source FR |
|---------|--------|--------------|-----------|
| GitHub Release | binaries + checksums + sigs + SBOMs | no (prereleases marked) | FR-016 |
| GHCR registry | multi-arch image; `:latest` moves only for stable | `latest` is stable-only | FR-017, FR-020 |
| Homebrew tap | `rune-task-runner/homebrew-tap` cask | yes (`skip_upload: auto`) | FR-018, FR-020 |
| Scoop bucket | `rune-task-runner/scoop-bucket` manifest | yes (`skip_upload: auto`) | FR-018, FR-020 |
| Install script | `scripts/install.sh` (curl\|sh) | resolves latest stable | FR-019 |

**Invariant**: a prerelease updates **none** of the stable-only channels (FR-020, SC-009).

---

## Cross-entity invariants

- **Version coherence**: every artifact for a Release reports exactly `Release.version`
  (`rune --version`, image label, archive name) — 100% (SC-002, SC-006).
- **Verifiability**: every artifact is checksummed + signed + provenanced; untampered →
  verifies, tampered → fails (SC-004).
- **Idempotency**: re-running a failed release converges to one Release with no duplicate tag
  (SC-008).
