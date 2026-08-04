# Data Model: VS Code Marketplace Extension for the Rune LSP

No runtime data or storage is involved; the "entities" here are publishing
artifacts and their relationships, plus the definitive inventory of touched
surfaces.

## Entities

### Extension Package (`.vsix`)

The installable artifact produced by `vsce package`.

| Field | Source | Rule |
|-------|--------|------|
| `name` | manifest `name: "runefile"` | Fixed; with publisher forms the extension ID `rune-task-runner.runefile`. (`"rune"` was rejected at first publish — Marketplace names are globally unique and taken.) |
| `version` | Release tag minus `v` (stamped, D4) | Plain `major.minor.patch`; never committed; unique per registry (duplicates rejected → idempotent retries). |
| Contents | `editors/vscode/` minus `.vscodeignore` | MUST include `extension.js`, grammar, language configuration, README, CHANGELOG, LICENSE, icon, and production `node_modules` (`vscode-languageclient`). |

**Relationships**: produced once per stable Release; uploaded as a GitHub
release asset; the same bytes are pushed to both Registry Listings.

**State transitions**: `packaged` → `attached-to-release` → `published
(marketplace)` + `published (open-vsx)`. Each transition is independently
retryable; `published` is terminal per version (registries are append-only).

### Marketplace Listing (per registry)

| Field | Source | Rule (contract L-refs) |
|-------|--------|------------------------|
| Display name / description | manifest | Identifies the official Rune extension (L1). |
| Publisher / namespace | `rune-task-runner` | Project-controlled on both registries (L2, D2). |
| Icon | `icon.png`, 128×128 | Required (L3). |
| README (listing body) | `editors/vscode/README.md` | Binary prerequisite stated up front with install link (L4). |
| License | `LICENSE` (MIT copy) | Present in package (L5). |
| Repository / bugs / homepage | manifest URLs | MUST point at `github.com/glapsfun/rune` (L6 — fixes existing wrong URL). |
| Changelog tab | `CHANGELOG.md` | Points to repo changelog (L7). |
| Version listed | == Rune release tag | Traceability both directions (FR-006). |

### Publisher Identity

| Attribute | Marketplace | Open VSX |
|-----------|-------------|----------|
| Namespace | `rune-task-runner` | `rune-task-runner` |
| Created via | manage portal (one-time runbook) | `ovsx create-namespace` (one-time runbook) |
| Credential | Azure DevOps PAT, Marketplace→Manage scope | Eclipse access token |
| Stored as | `VSCE_PAT` secret, `release` environment | `OVSX_PAT` secret, `release` environment |
| Rotation | ≤ 1 year expiry; runbook section | runbook section |

### Release (existing entity, extended)

A stable release (`workflow_dispatch` with `prerelease=false`) now
additionally triggers the `publish-extension` job. Prereleases (`-rc.N`)
never publish an extension. The core `release` job's steps are frozen — it
only gains a job-level `outputs:` mapping exposing the computed tag, and
the new job consumes that output plus the pushed tag and the created
GitHub release.

### Preflight Check (extension runtime behavior)

| State | Trigger | Outcome |
|-------|---------|---------|
| Binary found, `lsp` works | Runefile opened | LanguageClient starts (unchanged behavior). |
| Binary missing | spawn fails (ENOENT) | Error notification: "install Rune" + buttons → docs / settings; client NOT started. |
| Binary too old (no `lsp`) | probe exits non-zero / unknown command | Error notification: "upgrade Rune" + same buttons; client NOT started. |
| No Runefile in workspace | — | Extension never activates (`onLanguage:runefile` only). |

## Surface Inventory (files touched)

| # | Surface | Change |
|---|---------|--------|
| S1 | `editors/vscode/package.json` | Repository URL fix, icon/bugs/homepage/keywords/license fields, `0.0.0` placeholder version, `ovsx` + pinned `@vscode/vsce` dev-deps, publish scripts. |
| S2 | `editors/vscode/package-lock.json` | NEW — reproducible `npm ci`. |
| S3 | `editors/vscode/extension.js` | Preflight (D5): resolve `rune.path`, probe, actionable error notification, start client only on success. |
| S4 | `editors/vscode/README.md` | Rewritten as Marketplace listing page; build-from-source demoted to contributor note. |
| S5 | `editors/vscode/CHANGELOG.md` | NEW — points at repo changelog. |
| S6 | `editors/vscode/LICENSE` | NEW — MIT copy. |
| S7 | `editors/vscode/icon.png` | NEW — 128×128 listing icon. |
| S8 | `editors/vscode/.vscodeignore` | NEW — lean `.vsix`. |
| S9 | `.github/workflows/ci.yml` | NEW `extension` smoke job (D7). |
| S10 | `.github/workflows/release.yml` | NEW `publish-extension` job (D6, D9); `release` job byte-identical. |
| S11 | `docs/releasing.md` | One-time publisher setup runbook, per-release behavior, recovery + token rotation (D2, D8). |
| S12 | `editors/README.md` | VS Code section: Marketplace/Open VSX install first. |
| S13 | Root `README.md` | Editor-support bullet links to the Marketplace listing. |
