# Releasing Rune

Rune releases are automated end-to-end by the **Release** GitHub Actions workflow
(`.github/workflows/release.yml`) driving [GoReleaser](https://goreleaser.com). A maintainer
picks a version bump; the workflow computes the tag, updates the changelog, and publishes
binaries, multi-arch images, signatures, SBOMs, provenance, and the Homebrew/Scoop packages.

## Cutting a release

1. Ensure `main` is green (CI passing) and contains everything you want to ship.
2. **Actions → Release → Run workflow**, on branch `main`:
   - **bump**: `patch`, `minor`, or `major`.
   - **prerelease**: check to cut a `-rc.N` (release candidate).
3. Approve the `release` environment prompt (only authorized maintainers can).

The workflow then:

- computes the next `vX.Y.Z` (or `vX.Y.Z-rc.N`) from the latest tag and your bump, refusing
  if that tag already exists;
- prepends a dated section to `CHANGELOG.md` (from Conventional-Commit PR titles) and commits
  it to `main`;
- creates and pushes the annotated tag;
- runs GoReleaser: 6 binary archives + `checksums.txt`, a multi-arch GHCR image
  (`linux/amd64` + `linux/arm64`), cosign signatures, SPDX SBOMs, and — for stable releases —
  the `latest` image tag plus updated Homebrew cask and Scoop manifest;
- attaches GitHub build-provenance attestations to the binaries and the image.

> **Versioning note (pre-1.0):** under SemVer, `0.y.z` treats a breaking change as a **minor**
> bump (not major). You choose the bump; the tooling does not auto-infer it.

## Verifying a release

Anyone can verify artifacts with public material only — no pre-shared secret.

```sh
# 1. Checksum (the install script does this automatically)
sha256sum --check checksums.txt --ignore-missing

# 2. Signature of the checksums file (covers every archive)
cosign verify-blob \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp 'https://github.com/glapsfun/rune/.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  checksums.txt

# 3. Build provenance
gh attestation verify checksums.txt --repo glapsfun/rune

# 4. Image signature + provenance
cosign verify ghcr.io/glapsfun/rune:<version> \
  --certificate-identity-regexp 'https://github.com/glapsfun/rune/.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com'
gh attestation verify oci://ghcr.io/glapsfun/rune:<version> --repo glapsfun/rune
```

Modifying any byte of an artifact makes steps 1–2 fail — that is the point.

## Local dry-run (no publish)

```sh
rune release-dryrun            # goreleaser release --snapshot --clean
goreleaser check               # validate .goreleaser.yaml only
```

A full local dry-run needs `goreleaser`, `cosign`, `syft`, and `docker` (buildx) installed.
CI runs a lean dry-run (`check` + a snapshot skipping sign/sbom/docker) on every push/PR.

## Recovering from a failed release

The workflow is safe to re-run:

- The tag-exists check makes version computation idempotent; `--clean` wipes `dist/` first.
- GitHub Releases are upserted and GHCR pushes are content-addressed, so re-pushing is safe.
- If the tag was already created but publishing failed **after** that, delete the tag and the
  (draft) release, then re-run — or re-run GoReleaser locally against the existing tag.
- **Homebrew tap / Scoop bucket** commits are the one spot that may need manual cleanup if a
  run failed midway through updating them; check those repos before re-running.

## One-time setup (already provisioned? skip)

1. Create public repos `glapsfun/homebrew-tap` and `glapsfun/scoop-bucket`.
2. Mint a token that can push to those two repos (GitHub App install token preferred, or a
   fine-grained PAT with `contents: write`); store it as the secret `TAP_GITHUB_TOKEN`.
3. If `main` is protected, allow the release identity to bypass the push restriction (the
   workflow commits `CHANGELOG.md` to `main`); store that identity's token as `RELEASE_TOKEN`,
   or rely on the default token if branch protection permits.
4. Create a protected `release` Environment with required reviewers (authorized maintainers).
5. Enable **Settings → General → "Default to PR title for squash merge commits"** so the
   Conventional-Commit PR title becomes the squash commit subject the changelog reads.
6. Mark the **Validate PR title** and **release-dryrun** checks as required in branch protection.

## Deferred follow-ups

- **SHA-pin** all third-party actions in the workflows (currently floating major tags); let
  Dependabot keep them current.
- Migrate `dockers:` + `docker_manifests:` to **`dockers_v2:`** before GoReleaser v3 removes
  the classic blocks.
- Optionally publish moving **`v0` / `v0.4`** image tags in addition to `:<version>`/`:latest`.
- **Linux packages** (`.deb` / `.rpm`) and additional package managers (apt, winget, nix).
- **Apple notarization** / **Windows Authenticode** code-signing to remove OS "unidentified
  developer" prompts (requires paid certificates). The Homebrew cask strips the macOS
  quarantine attribute as an interim measure.
