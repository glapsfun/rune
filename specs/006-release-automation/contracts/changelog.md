# Contract: Changelog & Release Notes

Single source: **git-cliff** drives both the committed file and the release-notes body
(research §H). GoReleaser's own changelog is disabled (`changelog: { disable: true }`).

## Committed file — FR-012/015

- `CHANGELOG.md` in repo root, **Keep a Changelog** format.
- One section per release: `## [<version>] - YYYY-MM-DD`.
- Entries grouped by type, mapped from Conventional-Commit prefixes:
  `feat`→**Added**, `fix`→**Fixed**, `perf`/`refactor`→**Changed** (full map in `cliff.toml`).
- Each version links to its tag and a compare URL vs the previous version (FR-015).
- New section is **prepended** at release time: `git cliff --unreleased --tag <next> --prepend CHANGELOG.md`.

## Release notes — FR-014

- The GitHub Release body = `git cliff --latest` output → **byte-identical** to the new
  `CHANGELOG.md` section. Passed to GoReleaser via `--release-notes <(git cliff --latest)`.

## Source of entries — FR-013/013a

- Derived from **Conventional-Commit squash-merge PR titles** since the previous tag.
- Enforced by a required PR-title check (`amannn/action-semantic-pull-request`) +
  repo setting "Default to PR title for squash merge commits".

## Edge cases

- **First managed entry**: hand-curated for current state; git-cliff manages from the next
  tag forward (free-form history before adoption is not retro-parsed) — research §H4.
- **Empty change set**: warn the maintainer; do not emit an empty/broken section.
- **Prerelease**: `-rc.N` gets its own section, folded into the GA section at release (FR-005).
