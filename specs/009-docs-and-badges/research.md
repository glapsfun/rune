# Phase 0 Research: Modern, Example-Rich Documentation & README Status Badges

Consolidated findings for the two open areas the spec deferred to planning: (A) README status
badges (providers, exact URLs, integrity) and (B) the documentation-structure framework for
reorganizing in place. All badge URLs and framework/syntax claims were verified against live
sources (shields.io, pkg.go.dev, goreportcard.com, diataxis.fr, GitHub docs).

---

## A. README status badges

### A1. Badge set, providers, and exact URLs

- **Decision**: Seven badges, canonical targets kept strictly separate — repo-scoped
  (`glapsfun/rune`): CI, release/tag, license, Go version; module-scoped
  (`rune-task-runner/rune`): Go Report Card, Go Reference. Plus a static Docs badge. Exact
  URLs and link targets are pinned in [`contracts/badges.md`](./contracts/badges.md).
- **Rationale**: These are the sets the user selected during clarification. The repo/module
  split is real (git remote vs. Go import path) and shields fetches repo files from the repo
  host while Go tooling resolves the module path — crossing them breaks badges.
- **Alternatives considered**: —

### A2. CI badge — shields.io vs. native GitHub SVG

- **Decision**: shields.io `github/actions/workflow/status/glapsfun/rune/ci.yml?branch=main`.
- **Rationale**: One consistent style across the whole row; `?branch=main` keeps
  non-default-branch failures from reddening the default view; supports style/cache tuning.
  shields takes the workflow **filename** (`ci.yml`), not the display name.
- **Alternatives considered**: Native `actions/workflows/ci.yml/badge.svg` — zero-config and
  independent of shields' uptime, but taller/fixed style that clashes in a mixed row. Rejected
  for row consistency.

### A3. Release badge — `v/tag` vs. `v/release`

- **Decision**: `github/v/tag/glapsfun/rune?sort=semver`, linking to `/tags`.
- **Rationale**: A plain git tag `v0.1.0` exists but there may be no published GitHub *Release*
  object; the tag variant always renders a version. Revisit to `github/v/release` once real
  GitHub Releases are cut.
- **Alternatives considered**: `github/v/release` — hides prereleases unless
  `?include_prereleases`, and shows `no releases` when no Release object exists. Deferred.

### A4. License badge — static vs. dynamic

- **Decision**: Static `badge/License-MIT-yellow.svg`, linking to `/blob/main/LICENSE`.
- **Rationale**: Always reads "MIT" regardless of GitHub's license auto-detection, which can
  show `unknown` for nonstandard `LICENSE` files.
- **Alternatives considered**: Dynamic `github/license/glapsfun/rune` — rejected for
  reliability.

### A5. Style and layout

- **Decision**: One shields style for the whole row (`flat`/`flat-square`, not
  `for-the-badge`); centered `<p align="center">` HTML block of `<a><img alt=…></a>` pairs
  under the H1; every `<img>` carries `alt` text.
- **Rationale**: GitHub markdown does not center plain lines; a 7-badge `for-the-badge` row is
  too wide; `alt` text is the graceful-degradation path (FR-005) when a provider is down.
  shields badges are legible in both light/dark themes by design.
- **Alternatives considered**: Left-aligned single markdown line (simpler, also common) — kept
  as fallback if the centered block proves fussy.

### A6. Badge integrity as a test, not live fetch

- **Decision**: `test/docs/badges_test.go` asserts URL **shape and targeting** (canonical
  repo/module, no placeholder tokens, alt text present, image wrapped in the correct link,
  real `ci.yml` referenced) — it does **not** perform live HTTP.
- **Rationale**: CI has no network and the existing `links_test.go` already declines to fetch
  external `http(s)` links (flaky). Shape/targeting checks catch the realistic failure modes
  (typo'd owner, crossed repo/module, missing link) deterministically.
- **Alternatives considered**: Live badge fetch (rejected: network flakiness, rate limits).

---

## B. Documentation structure (reorganize in place)

### B1. Framework — Diátaxis, mapped to the clarified sections

- **Decision**: Organize by Diátaxis modes. Map: **tutorial** = `getting-started.md` (the one
  true guaranteed linear path); **how-to** = the folded `docs/guides/*` → `docs/how-to/*`;
  **reference** = `cli.md`, `runefile.md`, `GRAMMAR.md`, `mcp.md` (kept austere); **use-cases**
  = project-shaped **how-tos** built from the existing `examples/*` (Python, Node, MCP);
  **user-guide** = a curated ordered *tour* that links into the other modes rather than
  duplicating them.
- **Rationale**: The existing `docs/guides/*` are already how-to-shaped ("one capability per
  page, syntax + example + pitfalls"), so folding them into `how-to/` is a rename, not a
  rewrite. Use-cases serve a reader who arrives *with their own project/goal* → how-to, not
  tutorial (the decisive Diátaxis test). Keeping `user-guide/` as connective tissue prevents
  the classic rot of two pages explaining the same thing.
- **Alternatives considered**: Treating `user-guide/` as one long tutorial (rejected: it's a
  tour, not a guaranteed-success lesson); treating use-cases as tutorials (rejected: readers
  bring their own project).

### B2. Explanation content

- **Decision**: Keep the clarified three-folder structure (`how-to/`, `user-guide/`,
  `use-cases/`); house **explanation**-mode content (the "why opt-in caching", "Rune vs
  make/just", the MCP security model) in `overview.md` and the `user-guide/` narrative rather
  than introducing a fourth `explanation/` folder in this feature.
- **Rationale**: Respects the Q1 clarification (which named exactly those three new sections)
  while still giving explanation content a clear home. A dedicated `explanation/` bucket is a
  reasonable *future* refinement, noted but out of scope here to avoid empty scaffolding.
- **Alternatives considered**: Adding `explanation/` now (deferred — scope creep beyond the
  clarified structure; Diátaxis itself warns against building empty four-folder scaffolding).

### B3. Goal-oriented index

- **Decision**: `docs/README.md` becomes the intent-first router — an `| I want to… | Go to |`
  table routing by reader goal (first task → tutorial; cache/params/parallel → how-to; Python/
  Node/agent → use-cases; flag/grammar lookup → reference), plus a compact "by document type"
  section. GitHub auto-renders it when the folder is opened.
- **Rationale**: Tables scan fast and render cleanly on GitHub; extends the pattern already in
  `README.md` (docs table) and `guides/README.md`. Satisfies FR-013 (≤2 clicks to any page).
- **Alternatives considered**: Long bulleted prose (rejected: slower to scan).

### B4. "Modern/fancy" on GitHub without a site generator

- **Decision**: Use, sparingly, (1) **GitHub Alerts** — `> [!NOTE|TIP|IMPORTANT|WARNING|
  CAUTION]` blockquotes (one or two per page, never nested/stacked) for the "pitfalls" call-
  outs the guides already write in prose; (2) **`<details>`** collapsible blocks (blank line
  after `<summary>`) for long expected output / advanced options; (3) a short manual anchor
  TOC only on long reference pages (GitHub auto-generates a TOC via the header icon otherwise);
  (4) a standardized **"Next steps" footer** on every page; (5) **command + expected-output**
  block pairs (fenced `sh`, then fenced `text`).
- **Rationale**: These are the only theme-safe, asset-free devices that materially raise polish
  on github.com; each maps directly to an FR (FR-015 output, FR-016 callouts, FR-014
  structure/footers). No image/logo assets (per Q3).
- **Alternatives considered**: Site generator / hosted theme (rejected by Q1); custom HTML/CSS
  (GitHub sanitizes most of it).

### B5. Reorganizing without breaking links (GitHub has no server-side redirects)

- **Decision**: `git mv` each moved page (history follows); update **all** internal links in
  the same change; leave a one-line **redirect stub** at each old path
  (`> [!NOTE]` pointing to the new location) for external inbound links, kept for a
  deprecation window; do it **incrementally**, not big-bang. All links stay **relative**.
- **Rationale**: GitHub serves raw files with no redirect layer, so a move 404s every existing
  link (including `README.md`'s docs table and cross-links in `overview.md`/`guides/README.md`).
  Stubs are the closest thing to a redirect; relative links keep the tree portable in-editor
  and on github.com. Enforced by the existing `links_test.go` (no broken internal links).
- **Alternatives considered**: Move without stubs (rejected: breaks external links); a
  third-party link checker (rejected: the repo already has `links_test.go`, no new dep).

### B6. Prior art to model

- **Decision**: Model the split on **ripgrep** (narrative `GUIDE.md` kept separate from
  reference), **just** (peer command runner; heavy anchor TOC + command/output pairs), and
  **chezmoi** (explicit `user-guide/` folder split from reference) — all markdown-in-repo,
  no hosted site.
- **Rationale**: Direct precedents for exactly the in-repo, GitHub-rendered handbook shape and
  the `user-guide/` naming.
- **Alternatives considered**: —

---

## Verification approach (reuses existing infrastructure)

- **Decision**: Extend `test/docs` (CI gate `docs-verify`, run via `rune docs-check`) — add
  `badges_test.go`; grow `codeblocks_test.go`'s `selfContainedPages` allowlist as reorganized
  pages come into compliance; rely on `links_test.go` (already scans `docs/**/*.md` +
  `README.md`) to cover the reorg and stubs; keep `examples_test.go` two-tier verification for
  the use-case-backing examples.
- **Rationale**: Constitution Principle VI already makes docs tested fixtures and gate #6
  already means "examples run + links resolve"; the Q2 clarification (enforce ongoing) is
  satisfied by extension, not new infrastructure.
- **Alternatives considered**: A separate docs pipeline (rejected: duplicates the existing
  harness).

---

## Resolved unknowns

All Technical-Context items are resolved; no `NEEDS CLARIFICATION` remain. Deferred, out of
scope, and non-blocking: a dedicated `explanation/` folder (future), and switching the release
badge to `v/release` once real GitHub Releases are cut.
