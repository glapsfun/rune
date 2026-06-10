# Quickstart: Validating Release Automation

**Feature**: 006-release-automation · **Date**: 2026-06-10

Runnable scenarios that prove the feature works end-to-end. Implementation details live in
`tasks.md`; this is a validation/run guide. Tests run **inside Docker** per project policy.

## Prerequisites

- `goreleaser` (`~> v2`, ≥ v2.16), `cosign`, `syft`, `git-cliff`, `docker` (buildx), `gh`.
- Repo checked out with full history + tags (`git fetch --tags`).
- For a *real* release: `TAP_GITHUB_TOKEN` secret + the protected `release` Environment set up.

---

## Scenario 1 — Config is valid (fast gate)

```sh
goreleaser check
```

**Expected**: `config is valid`. Catches deprecated keys (e.g. `brews:` → `homebrew_casks:`).
Wire this into `ci.yml` (Constitution gate #7).

## Scenario 2 — Full dry-run, no publish (FR-027, gate #7)

```sh
goreleaser release --snapshot --clean
```

**Expected**: `dist/` contains all 6 archives, `checksums.txt`, per-archive SBOMs, and local
multi-arch image layers tagged `…-amd64`/`…-arm64` — and **nothing is pushed**. This is the
existing `release-dryrun` Runefile task.

Verify the matrix:

```sh
ls dist/*.tar.gz dist/*.zip | wc -l        # expect 6 archives (5 tar.gz + 1 zip... )
ls dist/*.sbom.spdx.json                   # one SBOM per archive
```

## Scenario 3 — Version string matches (FR-006, SC-006)

```sh
./dist/rune_*_$(go env GOOS)_$(go env GOARCH)*/rune --version
```

**Expected**: `rune version <snapshot-version> (commit <shortsha>)`.

## Scenario 4 — Changelog generation (FR-012/014)

```sh
git cliff --unreleased --tag v9.9.9 | head -40         # preview a section
diff <(git cliff --latest) <(sed -n '/## \[/,/## \[/p' CHANGELOG.md | head -n -1)
```

**Expected**: a Keep a Changelog section grouped Added/Changed/Fixed; the `--latest` render
equals the top `CHANGELOG.md` section (single source — they feed release notes identically).

## Scenario 5 — Install script (FR-019)

```sh
shellcheck scripts/install.sh
# against a published release (or a local file:// served dir):
INSTALL_DIR=/tmp/runebin sh scripts/install.sh
/tmp/runebin/rune --version
```

**Expected**: detects OS/arch, downloads the right archive, **verifies the checksum**,
installs `rune`, and the installed binary reports the version. A tampered archive aborts the
install.

## Scenario 6 — Verify a real release artifact (FR-021/022/024, SC-004)

See [contracts/verification.md](./contracts/verification.md). Smoke version:

```sh
sha256sum --check checksums.txt --ignore-missing                       # checksum
cosign verify-blob --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp 'https://github.com/glapsfun/rune/.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' checksums.txt
gh attestation verify checksums.txt --repo glapsfun/rune       # provenance
```

**Negative test**: flip a byte in an archive → checksum + signature verification MUST fail.

## Scenario 7 — Multi-arch image (FR-010, US2)

```sh
docker buildx imagetools inspect ghcr.io/glapsfun/rune:<version>
docker run --rm --platform linux/arm64 ghcr.io/glapsfun/rune:<version> --version
docker run --rm --platform linux/amd64 ghcr.io/glapsfun/rune:<version> --version
```

**Expected**: the manifest lists `linux/amd64` + `linux/arm64`; each platform runs and reports
the version.

## Scenario 8 — Cut a real release (maintainer, FR-001/002/003)

Actions → **Release** → Run workflow → pick `bump` (+ `prerelease` if an rc).

**Expected**: tag `vX.Y.Z` created; GitHub Release with all assets; GHCR image (+ `latest` if
stable); Homebrew cask + Scoop manifest updated (skipped for prereleases); `CHANGELOG.md`
updated on `main`. Re-running after a mid-way failure converges without duplicate tags (SC-008).

---

## Acceptance trace

| Scenario | Covers |
|----------|--------|
| 1, 2 | FR-027; Constitution gate #7 |
| 2, 3 | FR-006/007/009/021/023; SC-002/006 |
| 4 | FR-012/013/014/015 |
| 5 | FR-019; SC-003 |
| 6 | FR-021/022/024; SC-004 |
| 7 | FR-010/017; US2 |
| 8 | FR-001/002/003/005/016/018/020; SC-001/008/009 |
