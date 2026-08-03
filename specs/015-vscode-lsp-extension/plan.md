# Implementation Plan: VS Code Marketplace Extension for the Rune LSP

**Branch**: `015-vscode-lsp-extension` | **Date**: 2026-07-29 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/015-vscode-lsp-extension/spec.md`

## Summary

The VS Code extension client for `rune lsp` already exists at
`editors/vscode/` (feature 011) but is only installable by building a
`.vsix` by hand. This feature publishes it to the VS Code Marketplace and
the Open VSX registry, and wires publishing into the existing
maintainer-initiated release workflow (feature 006) so every **stable**
release ships a matching extension version automatically. The work is
packaging/automation only — no Go code changes and no new language
features: harden the extension manifest for a public listing (icon,
license, correct repository URL, marketplace README), add a friendly
missing-binary preflight to the client, add a `publish-extension` job to
`.github/workflows/release.yml` gated on stable tags, add a CI packaging
smoke job so a broken package can never reach release day, and flip the
docs so "install from the Marketplace" is the primary VS Code path.

## Technical Context

**Language/Version**: Extension client: plain JavaScript (CommonJS) targeting
VS Code `^1.75.0`, Node 22 for packaging. Pipeline: GitHub Actions YAML.
**No Go changes** — the `rune lsp` server (features 010/011) is untouched.

**Primary Dependencies**: `vscode-languageclient` `^9.x` (already present,
the only runtime dependency); `@vscode/vsce` (dev, packaging/publishing to
the Marketplace); `ovsx` (dev, publishing to Open VSX). No bundler in v1
(research D3).

**Storage**: N/A. Publish credentials live only in GitHub Actions secrets
scoped to the protected `release` environment (`VSCE_PAT`, `OVSX_PAT`) —
never in the repository (research D2, D8).

**Testing**: CI packaging smoke job (`npm ci && vsce package` inside the
Node toolchain, artifact inspected for required files); manifest lint via
`vsce ls`; missing-binary UX verified by quickstart scenario; release
integration verified via the existing release-dryrun gate philosophy —
the publish job's package step also runs in the dry-run path. Go suite
untouched (Docker-only rule unaffected).

**Target Platform**: VS Code ≥ 1.75 on all OSes (the client is pure JS and
spawns the user's `rune` binary); listings on the Visual Studio Marketplace
(publisher `rune-task-runner`) and Open VSX (same namespace).

**Project Type**: Editor-extension packaging + release-pipeline automation
(monorepo subdirectory `editors/vscode/` + `.github/workflows/`).

**Performance Goals**: N/A beyond the existing activation contract — the
extension activates only `onLanguage:runefile` and stays dormant otherwise
(spec edge case; already true, pinned by quickstart scenario).

**Constraints**: The extension MUST NOT bundle the `rune` binary (FR-003);
only stable releases publish (`inputs.prerelease == false`); extension
version = release tag without the `v` (FR-006); plain-JS client stays
dependency-light — one runtime dependency, no build/transpile step; exit
codes/behaviour of the release job for binaries/images are frozen — the
extension publish is strictly additive and MUST NOT fail the core release
(FR-008, research D6).

**Scale/Scope**: One extension, two registries, ~6 files touched in
`editors/vscode/`, one new CI job, one new release-workflow job, docs
updates (`editors/README.md`, `editors/vscode/README.md`, root `README.md`,
`docs/releasing.md`).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principle | Status | Notes |
|---|-----------|--------|-------|
| I | Command Runner, Not a Build System | ✅ N/A | No runner/caching semantics touched. |
| II | Errors Are a Feature | ✅ Pass | The missing-binary preflight extends this ethos to the extension: actionable message with install guidance instead of a silent crash loop (FR-003). Diagnostics themselves unchanged. |
| III | Minimal, Total DSL | ✅ N/A | No DSL surface change; no `docs/GRAMMAR.md` impact. |
| IV | Hand-Written Front End, Idiomatic Go | ✅ Pass | Zero Go changes; locked `internal/` layout untouched. Extension client stays a thin JS shim (all intelligence in `rune lsp`). |
| V | Boringly Portable | ✅ Pass | No binary bundling — the extension spawns the user's `rune` from PATH/`rune.path`, so the single-static-binary story is unchanged. Client JS is platform-neutral. |
| VI | Test-First, Multi-Layer Verification | ✅ Pass | New CI packaging smoke job gates every push; quickstart scenarios pin listing content and missing-binary UX; docs changes go through the existing docs-verify harness. Golden/fuzz layers unaffected. |
| VII | AI-Native, Secure by Default | ✅ Pass | Publish tokens are GitHub environment secrets behind the protected `release` environment with required reviewers; never in-repo, never printed (masked by Actions). MCP surface untouched. |
| VIII | Go Engineering Discipline | ✅ N/A | No Go code in this feature. |
| — | Docker-only testing | ✅ Pass | Go suite unchanged. The extension packaging smoke runs in CI's Node toolchain (it is not part of the Go test suite); local validation uses a disposable Node container per quickstart. |
| — | Locked package layout | ✅ Pass | Only `editors/vscode/`, workflows, and docs change. |
| — | Surface changes carry their docs | ✅ Pass | Install docs, releasing docs, and the one-time publisher-setup runbook ship in the same PR. |

**Post-design re-check (after Phase 1)**: no new violations introduced; the
design keeps the release job's existing steps byte-identical and adds only a
job-level `outputs:` mapping (exposing the computed tag) plus a downstream
job. No Complexity Tracking entries required.

## Project Structure

### Documentation (this feature)

```text
specs/015-vscode-lsp-extension/
├── plan.md              # This file
├── research.md          # Phase 0 output (decisions D1–D10)
├── data-model.md        # Phase 1 output (entities + file/surface inventory)
├── quickstart.md        # Phase 1 output (validation scenarios V1–V6)
├── contracts/
│   ├── publishing-pipeline.md   # Release-integration contract (P1–P8)
│   └── marketplace-listing.md   # Listing/manifest contract (L1–L9)
├── checklists/requirements.md
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root)

```text
editors/
├── README.md                    # UPDATED: Marketplace install becomes the VS Code path
└── vscode/
    ├── package.json             # UPDATED: icon, license, repository URL fix
    │                            #   (glapsfun/rune), keywords, ovsx/vsce dev-deps,
    │                            #   placeholder version 0.0.0 (stamped at publish)
    ├── package-lock.json        # NEW: required for reproducible `npm ci` in CI/release
    ├── extension.js             # UPDATED: missing/old-binary preflight with
    │                            #   actionable notification (FR-003)
    ├── README.md                # UPDATED: becomes the Marketplace listing page
    ├── CHANGELOG.md             # NEW: shown on the listing's Changelog tab
    ├── LICENSE                  # NEW: copy of repo MIT license (vsce packages it)
    ├── icon.png                 # NEW: 128×128 listing icon
    ├── .vscodeignore            # NEW: keeps the .vsix lean
    ├── language-configuration.json  # unchanged
    └── syntaxes/runefile.tmLanguage.json  # unchanged

.github/workflows/
├── ci.yml                       # UPDATED: new `extension` smoke job
│                                #   (npm ci + vsce package + content assert)
└── release.yml                  # UPDATED: new `publish-extension` job after
                                 #   `release`, stable-only, publishes to both
                                 #   registries + attaches .vsix to the GitHub release

docs/
└── releasing.md                 # UPDATED: publisher one-time setup runbook,
                                 #   per-release behavior, failure recovery

README.md                        # UPDATED: editor-support blurb links to the listing
```

**Structure Decision**: Everything stays inside the existing
`editors/vscode/` subtree plus the two workflow files and docs — no new
top-level directories, no Go packages, honoring the locked package layout.
The extension version in `package.json` is a `0.0.0` placeholder stamped
from the git tag at publish time (never committed), so the repo carries no
version churn (research D4).

## Complexity Tracking

No constitution violations — table intentionally empty.
