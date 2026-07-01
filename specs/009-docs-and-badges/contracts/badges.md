# Contract: README Status Badges

Defines the exact badge set, image URLs, link targets, and integrity rules for the README
badge row (FR-001..FR-006a). The `test/docs/badges_test.go` harness (Phase 2) asserts these.

## Canonical targets (do not cross)

- **Repo host** (repo-scoped badges — CI, release/tag, go-version, license): `glapsfun/rune`
- **Module path** (module-scoped badges — Go Report Card, Go Reference): `rune-task-runner/rune`

## Badge set

| # | Badge | Image URL | Link target |
|---|-------|-----------|-------------|
| 1 | CI / build | `https://img.shields.io/github/actions/workflow/status/glapsfun/rune/ci.yml?branch=main` | `https://github.com/glapsfun/rune/actions/workflows/ci.yml` |
| 2 | Release (tag) | `https://img.shields.io/github/v/tag/glapsfun/rune?sort=semver` | `https://github.com/glapsfun/rune/tags` |
| 3 | License (MIT) | `https://img.shields.io/badge/License-MIT-yellow.svg` | `https://github.com/glapsfun/rune/blob/main/LICENSE` |
| 4 | Go version | `https://img.shields.io/github/go-mod/go-version/glapsfun/rune` | `https://go.dev/` (or unlinked) |
| 5 | Go Report Card | `https://goreportcard.com/badge/github.com/rune-task-runner/rune` | `https://goreportcard.com/report/github.com/rune-task-runner/rune` |
| 6 | Go Reference | `https://pkg.go.dev/badge/github.com/rune-task-runner/rune.svg` | `https://pkg.go.dev/github.com/rune-task-runner/rune` |
| 7 | Docs | `https://img.shields.io/badge/docs-README-blue` | `https://github.com/glapsfun/rune/blob/main/docs/README.md` |

## Decisions (from Phase 0 research)

- **CI badge → shields.io form, not native** — one consistent style across the row; native
  GitHub `badge.svg` is taller and clashes. Pin `?branch=main` so non-default-branch failures
  don't redden the default view.
- **Release → git-tag variant, not `v/release`** — a plain tag `v0.1.0` exists but there may
  be no published GitHub *Release* object; the `v/tag?sort=semver` form always shows something.
  Revisit to `github/v/release` once real GitHub Releases are cut.
- **License → static `badge/License-MIT-yellow` , not dynamic `github/license`** — static is
  reliable regardless of GitHub's license auto-detection.
- **Style** — one style for the whole row: `flat` (default) or `flat-square`. Do **not** use
  `for-the-badge` for a 7-badge row (too wide). Mixing providers is fine as long as the shields
  ones share a style.
- **Layout** — centered `<p align="center">` HTML block of `<a><img alt=…></a>` pairs under the
  H1 (GitHub markdown does not center plain lines). Every `<img>` MUST carry `alt` text so the
  badge degrades to a readable label + link when the provider is unavailable (FR-005).

## Integrity rules (asserted by `badges_test.go`)

1. **Canonical targets**: repo-scoped badge URLs contain `glapsfun/rune`; module-scoped badge
   URLs contain `rune-task-runner/rune`. No badge crosses the two.
2. **No placeholders**: no `USER/REPO`, `owner/repo`, `example`, or `TODO` tokens in any badge
   URL.
3. **Alt text present**: every badge image in the README has non-empty alt text (FR-005).
4. **Link + image pairing**: each badge image is wrapped in a link to its authoritative source
   per the table above (FR-003).
5. **Workflow filename**: the CI badge references the real workflow file `ci.yml` that exists
   under `.github/workflows/`.
6. **Static only for external network**: the harness validates URL *shape/targeting*, not live
   HTTP (CI has no network; matches the existing `links_test.go` policy of not fetching
   external `http(s)` links).

## Graceful-degradation notes (documented, not asserted live)

- CI badge before first run → `no status`; release/tag badge before any tag → `no releases` /
  `invalid` (tag variant avoids this since `v0.1.0` exists).
- Go Report Card grade computes on first visit to the report URL; Go Reference populates once
  `proxy.golang.org` has indexed the module. Both require the module path to be importable —
  verify the report + reference pages load once after publishing.
- **Module-path fallback (blocking check)**: Go Report Card and Go Reference require
  `github.com/rune-task-runner/rune` to be `go get`-able. If the vanity import is not
  configured (the repo is `glapsfun/rune`), these badges 404 permanently. Before shipping,
  confirm both provider pages load; if not, drop them or substitute repo-scoped badges.
