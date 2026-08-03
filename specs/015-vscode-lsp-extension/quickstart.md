# Quickstart: Validating the Marketplace Extension

Runnable scenarios proving the feature end-to-end. Contract references:
[publishing-pipeline.md](contracts/publishing-pipeline.md) (P1–P8),
[marketplace-listing.md](contracts/marketplace-listing.md) (L1–L9).

## Prerequisites

- Node 22+ and npm (packaging happens outside the Go/Docker harness — the
  Docker-only rule covers `go test`, not `vsce`). For a hermetic run use
  `docker run --rm -v "$PWD/editors/vscode:/w" -w /w node:20 <cmd>`.
- A built `rune` binary for the runtime scenarios: `go run ./cmd/rune` or a
  release install.
- VS Code ≥ 1.75 for the interactive scenarios.

## V1 — Package builds and contains the listing surface (P8, L3–L7)

```sh
cd editors/vscode
npm ci
npm version --no-git-tag-version 0.0.1
npx vsce package
unzip -l rune-0.0.1.vsix
```

(The throwaway version must be plain `major.minor.patch` — vsce rejects
prerelease suffixes like `0.0.1-smoke`.)

**Expected**: packaging succeeds with no manifest errors; the archive lists
`extension/extension.js`, `extension/syntaxes/runefile.tmLanguage.json`,
`extension/language-configuration.json`, `extension/readme.md`,
`extension/changelog.md`, `extension/LICENSE.txt`, `extension/icon.png`, and
`extension/node_modules/vscode-languageclient/…` — note vsce normalizes the
doc filenames (`readme.md`, `changelog.md`, `LICENSE.txt`). (This is exactly
what the CI `extension` job asserts, case-insensitively.)

## V2 — Side-load install gives working language features (L8)

```sh
code --install-extension editors/vscode/rune-0.0.1.vsix
```

Open a `Runefile` containing a known error (e.g. a dependency on an
undefined task).

**Expected**: syntax highlighting applies; a diagnostic with a stable
`RUNE####` code appears; completion/hover/outline/format all work — with
zero interaction with the repo beyond installing the `.vsix` (SC-002
analogue; the Marketplace path is V6).

## V3 — Missing/old binary shows actionable guidance (L9, SC-005)

1. In VS Code settings, set `rune.path` to a nonexistent path (e.g.
   `/tmp/no-such-rune`). Reload the window and open a Runefile.
   **Expected**: one error notification explaining the `rune` binary is
   required, with working "Install instructions" and "Open Settings"
   buttons; no repeated "server crashed" toasts; no server process spawned.
2. Point `rune.path` at a stub that fails on `lsp` (simulates a
   pre-LSP binary):
   `printf '#!/bin/sh\necho "unknown command" >&2\nexit 1\n' > /tmp/old-rune && chmod +x /tmp/old-rune`
   **Expected**: the notification asks to **upgrade** Rune, same buttons.
3. Restore `rune.path`; open a workspace with no Runefile.
   **Expected**: the extension does not activate (check the Running
   Extensions view).

## V4 — CI smoke job gates packaging regressions (P8)

Break the manifest deliberately on a scratch branch (e.g. point `icon` at a
missing file), push.

**Expected**: the `extension` CI job fails; reverting turns it green.

## V5 — Release integration, dry (P1–P5)

On a scratch run (or by inspection of `.github/workflows/release.yml`):

- `publish-extension` has `needs: release` and `if: ${{ !inputs.prerelease }}`
  — an rc dispatch shows the job **skipped** (P2).
- The `release` job's steps are unchanged relative to `main` before this
  feature (P1) — `git diff` shows only the new `publish-extension` job plus
  a job-level `outputs:` mapping on `release` exposing the computed tag.
- Secrets appear only as `secrets.VSCE_PAT` / `secrets.OVSX_PAT`
  environment references (P5).

## V6 — Post-release verification (first real publish)

After the first stable release with this feature (tag `vX.Y.Z`):

1. `.vsix` asset `rune-X.Y.Z.vsix` is attached to the GitHub release (P4).
2. Marketplace: search "Rune" in VS Code → the `rune-task-runner.rune`
   extension appears with icon, description, license, and repository link
   pointing at `github.com/glapsfun/rune` (L1–L6); listed version is
   `X.Y.Z` (P3, FR-006). Install it on a machine with `rune` on PATH, open
   a Runefile → diagnostics appear within 5 minutes of starting the search
   (SC-001, SC-002).
3. Open VSX: `https://open-vsx.org/extension/rune-task-runner/rune` shows
   the same version (P7); install in VSCodium and repeat the smoke check
   (US3).
4. Re-run test: re-run the `publish-extension` job for the same tag →
   both registry steps log a duplicate-version skip (`--skip-duplicate`),
   exit 0, and the run ends green with no side effects (P6).
