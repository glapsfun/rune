# Contract: Release Workflow (maintainer interface)

**Surface**: a single GitHub Actions workflow (`.github/workflows/release.yml`) triggered by
**`workflow_dispatch`**. This is the one action a maintainer takes to cut a release (SC-001).

## Inputs

| Input | Type | Required | Values | Meaning |
|-------|------|----------|--------|---------|
| `bump` | choice | yes | `patch` \| `minor` \| `major` | Which SemVer component to increment from the latest tag (FR-002) |
| `prerelease` | boolean | no (default `false`) | — | When true, produce `<next>-rc.N` and mark as a GitHub pre-release (FR-005) |

## Preconditions (the workflow MUST refuse otherwise)

1. Triggered on `main` (`github.ref == refs/heads/main`) — else fail (FR-025).
2. Working tree clean (`git status --porcelain` empty) — else fail (FR-025).
3. CI green for `HEAD` (all check-runs `success`) — else fail (FR-026).
4. Caller authorized via the protected `release` Environment (required reviewers) (FR-029).
5. Computed tag does not already exist — else fail (FR-004).

## Guarantees (postconditions on success)

- A new annotated tag `vX.Y.Z[-rc.N]` exists, created by the automation (FR-003).
- `CHANGELOG.md` has a new dated section; the GitHub Release notes match it (FR-012, FR-014).
- A GitHub Release exists with: 6 binary archives, `checksums.txt`, signatures, SBOMs
  (FR-007, FR-009, FR-016, FR-021–023); marked pre-release iff `prerelease` (FR-005).
- A multi-arch image `ghcr.io/rune-task-runner/rune:<version>` is pushed (+ `:latest` iff
  stable), signed, with provenance (FR-010, FR-017, FR-020, FR-022, FR-024).
- Stable releases update the Homebrew cask + Scoop manifest; prereleases do not (FR-018, FR-020).

## Failure semantics

- Any required artifact failing to build fails the whole run — **no partial publish** of
  binaries (edge case).
- The run is **safe to re-run** with the same inputs once a tag exists / after fixing the
  cause; it converges to one complete release (FR-028, SC-008). GitHub Releases upsert; GHCR
  pushes are content-addressed; tap/bucket commits may need manual cleanup.

## Dry-run contract (no publish)

`goreleaser release --snapshot --clean` (and `goreleaser check`) MUST succeed locally / in CI
and produce every artifact type without pushing anything (FR-027; Constitution gate #7).

## Required secrets / permissions

- `GITHUB_TOKEN` (this repo) — releases, tag push, GHCR.
- `TAP_GITHUB_TOKEN` (GitHub App install token or fine-grained PAT) — push to
  `homebrew-tap` + `scoop-bucket` (research §F).
- Job permissions: `contents: write`, `packages: write`, `id-token: write`,
  `attestations: write` (research §J).
