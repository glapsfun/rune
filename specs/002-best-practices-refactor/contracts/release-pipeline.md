# Contract — Release Pipeline

The "one tag → complete release" interface. Realizes FR-025/026/027, SC-006/007. Implemented in
`.github/workflows/release.yml` + the existing `.goreleaser.yaml`.

## Trigger

`push` of a tag matching `v*` (e.g. `v0.1.0`). Manual `workflow_dispatch` allowed for re-runs.

`permissions: { contents: write, packages: write }` — needed to create the GitHub Release and
push to GHCR. Auth uses the workflow's `GITHUB_TOKEN` only; **no additional secrets** (Principle
VII — no secret material in config/logs).

## Produced artifacts (the contract)

| Artifact | Produced by | Detail |
|----------|-------------|--------|
| Cross-platform binaries | goreleaser | linux/macos/windows × amd64/arm64, `-trimpath`, `-s -w`, version+commit ldflags |
| Archives | goreleaser | `tar.gz` (zip on Windows), **each including `LICENSE` + `README`** (FR-027) |
| `checksums.txt` | goreleaser | sha256 over all archives |
| GitHub Release + changelog | goreleaser | GitHub-sourced changelog (`changelog.use: github`) |
| Container image | buildx | `ghcr.io/rune-task-runner/rune:<version>` + `:latest`, multi-arch (see `docker-image.md`) |

## Sequence

1. Checkout (full history for changelog) + `setup-go` (`go-version-file: go.mod`).
2. `goreleaser release --clean` → binaries, archives, checksums, GitHub Release.
3. `docker/login-action` to GHCR (`GITHUB_TOKEN`) → `docker buildx build --push` the production
   image, tags `<version>` + `latest`, platforms `linux/amd64,linux/arm64`.

## Pre-publish safety (SC-006)

The CI **release-dryrun** job (`goreleaser release --snapshot --clean`, see `ci-gates.md`) runs
on every PR/push and **fails before any tag** if the config is invalid or a referenced file
(`LICENSE`/`README`) is missing. This guarantees every release-config reference resolves before a
real release is ever attempted (the spec edge case: "release config references a missing file").

## Dependencies / preconditions

- `LICENSE` (S1) and `README.md` (A1) MUST exist (archive refs) — ordering constraint in
  `data-model.md`.
- The production `Dockerfile` (D1) MUST build (gated by CI) before its publish step.
- `.goreleaser.yaml`: verify `archives.files` includes `LICENSE*` + `README*` (already present);
  no breaking changes to `builds` (GOOS/GOARCH already match the contract).

## Non-goals

- Package-manager taps (Homebrew/Scoop/apt) — out of scope (spec).
- Signing/SBOM/attestation — not required this iteration (could be a later additive change).
- Auto-version bumping — release is driven by a human-pushed `v*` tag.
