# Contract: Artifact Verification (consumer interface)

Anyone can verify a release with **public material only** — no pre-shared secret (FR-022, SC-004).

## 1. Checksum — FR-021

```sh
# from the dir containing the downloaded archive + checksums.txt
sha256sum --check checksums.txt --ignore-missing
```

Untampered → `OK`; tampered → checksum mismatch (SC-004).

## 2. Signature (cosign keyless) — FR-022

```sh
cosign verify-blob \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp 'https://github.com/rune-task-runner/rune/.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  checksums.txt
```

Verifies the checksums file was signed by Rune's release workflow identity. The install
script performs the checksum step automatically; signature verification is documented for
consumers who want it.

## 3. Container image signature — FR-022

```sh
cosign verify \
  --certificate-identity-regexp 'https://github.com/rune-task-runner/rune/.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  ghcr.io/rune-task-runner/rune:<version>
```

## 4. Build provenance — FR-024

```sh
gh attestation verify checksums.txt --repo rune-task-runner/rune
gh attestation verify oci://ghcr.io/rune-task-runner/rune:<version> --repo rune-task-runner/rune
```

Confirms *how/where* the artifact was built (SLSA-style provenance), complementing the
signature's *authenticity* claim.

## 5. SBOM — FR-023

```sh
# inspect dependencies shipped in an archive
cat rune_<version>_<os>_<arch>.sbom.spdx.json | jq '.packages[].name'
```

## Negative test (proves verification works) — SC-004

Modify any byte of an archive (or `checksums.txt`) and re-run steps 1–2: the checksum check
and `cosign verify-blob` MUST fail.
