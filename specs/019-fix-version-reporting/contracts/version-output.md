# Contract: Version Output & Compatibility Gate

The externally observable contract for `rune --version`, `rune version`,
`rune version --check [--json]`, and the `minimum_version` gate, across
install methods. Any violation is a bug; C1–C3 are hard regression gates.

## Output format (all install methods)

- **C1** — `rune --version` prints exactly one line:
  `rune version <VERSION> (commit <COMMIT>)`. `<COMMIT>` is the stamped
  short commit or the literal `none`. No other shapes.
- **C2** — The first line of `rune version` is byte-identical to the
  `rune --version` line (FR-005); its second line
  (`runefile language <N>`) is unchanged by this feature.
- **C3** — **Release artifacts (GoReleaser archive / Homebrew / Docker):
  output is byte-identical to pre-fix** — `<VERSION>` from the release tag
  minus `v`, `<COMMIT>` the stamped short commit (FR-002, SC-003).

## Version value by install method

- **C4** — `go install …/cmd/rune@vX.Y.Z` (and `@latest`, resolving to the
  newest tag): `<VERSION>` = `X.Y.Z`, `<COMMIT>` = `none` (FR-001, FR-007).
- **C5** — `go install …/cmd/rune@<untagged-commit>`: `<VERSION>` = the Go
  pseudo-version minus the leading `v` (e.g. `0.4.4-0.20260812…-432dbb8de56b`)
  — visibly not a clean release.
- **C6** — Builds from a source checkout (`go build`, `go run`, any
  tag/dirty state) and builds without module info: `<VERSION>` = `dev`,
  `<COMMIT>` = `none` unless ldflags-stamped (FR-003). A binary built with
  `-tags runetest` may override `<VERSION>` via `RUNE_TEST_VERSION`
  (test-only; production builds never honor it).

## Compatibility gate (`minimum_version`)

- **C7** — The gate, `rune version --check`, and `--check --json` all
  consume the same `<VERSION>` defined above (FR-004, FR-006). For C4/C5
  values the gate compares semver: older-than-required aborts with the
  standard diagnostic (exit 3, `installed version: <VERSION>`, nothing
  executed); satisfying versions pass silently.
- **C8** — `dev` remains a development build: gate bypassed (dev=true),
  never blocked — behavior identical to today (spec Story 3 / assumption).

## Validation record (2026-08-12, T014)

- [x] C1 — verified: stamped, dev, and old-stamped binaries all print the
  single `rune version <V> (commit <C>)` line
- [x] C2 — verified: `diff <(rune --version) <(rune version | head -1)`
  byte-identical
- [x] C3 — verified: `-X main.version=9.9.9 -X main.commit=abc1234` build
  prints `rune version 9.9.9 (commit abc1234)`; installed Homebrew 0.4.3
  output unchanged; GoReleaser snapshot stamps `0.4.4-next` as before
- [x] C4/C5 — unit-verified (`TestVersionFromBuildInfo` module-cache
  cases); live `go install …@<fix-commit>` pseudo-version check pending the
  branch being pushed (quickstart V2 pre-release form), clean-version check
  post-release via `@latest`
- [x] C6 — verified: `go run` and `go build` from this checkout (which Go
  stamps with a pseudo-version + `vcs.*`) both report
  `rune version dev (commit none)`
- [x] C7 — verified: 0.4.3-stamped binary vs `minimum_version "9.0.0"` →
  caret diagnostic, `installed version: 0.4.3`, `nothing was executed`,
  exit 3; `version --check` and `--check --json` report the same
- [x] C8 — verified: dev binary runs the same Runefile (exit 0) and
  `--check --json` reports `"development": true, "compatible": true`
