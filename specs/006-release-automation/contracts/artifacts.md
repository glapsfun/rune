# Contract: Published Artifacts (consumer interface)

The naming/layout consumers (install script, docs, automation) may rely on. Driven by
GoReleaser defaults; pinned here so changes are deliberate.

## Binary archives (GitHub Release assets) — FR-007/009

```
rune_<version>_<os>_<arch>.tar.gz      # linux, darwin
rune_<version>_<os>_<arch>.zip         # windows
```

- `<version>` = tag without the leading `v` (e.g. `0.4.0`, `0.4.0-rc.1`).
- `<os>` ∈ {`linux`, `darwin`, `windows`}; `<arch>` ∈ {`amd64`, `arm64`}.
- 6 archives per release; each contains the `rune` binary (+`.exe` on Windows), `LICENSE`,
  `README`.

## Checksums — FR-021

```
checksums.txt        # one "<sha256>  <archive-filename>" line per asset
```

## Signatures (cosign keyless) — FR-022

```
checksums.txt.sigstore.json     # bundle for the checksums file (covers all archives)
checksums.txt.pem               # certificate
```

Images are signed in-registry (verify by ref/digest, see verification.md).

## SBOMs — FR-023

```
rune_<version>_<os>_<arch>.sbom.spdx.json     # one per archive (SPDX-JSON)
```

## Provenance — FR-024

GitHub artifact attestation over `checksums.txt` and the image digest; verified with
`gh attestation verify` (no asset file the consumer manages directly).

## Container image — FR-010/017

```
ghcr.io/glapsfun/rune:<version>     # multi-arch manifest (linux/amd64 + linux/arm64)
ghcr.io/glapsfun/rune:latest        # stable releases only (FR-020)
```

Pulling the version/latest ref on amd64 or arm64 serves the matching variant automatically.
OCI labels carry `org.opencontainers.image.version` = `<version>` and `.revision` = commit.

## Version string — FR-006

```
rune version <version> (commit <shortsha>)
```

Reported by `rune --version` and `rune version`; matches the release tag exactly (SC-006).
