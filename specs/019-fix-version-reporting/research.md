# Research: Accurate Version Reporting for Go-Toolchain Installs

All Technical Context unknowns resolved. Ground truth was verified
empirically on this repo (build + `go version -m`) rather than taken from
documentation alone.

## D1 — Source of truth for the fallback version: `runtime/debug.ReadBuildInfo()`

**Decision**: When the ldflags-stamped `main.version` is still the `dev`
placeholder, resolve the reported version from
`debug.ReadBuildInfo().Main.Version`.

**Rationale**: The Go toolchain embeds the main module's version in every
module-mode binary. For `go install github.com/rune-task-runner/rune/cmd/rune@vX.Y.Z`
(or `@latest`, which resolves to the newest tag) `Main.Version` is exactly
`vX.Y.Z` — no network, no files, no new dependencies, identical on all
three OSes. This is the standard idiom for exactly this bug (used by
golangci-lint, gopls, etc.).

**Alternatives considered**:
- *Stamp via `go install` instructions with `-ldflags` in docs* — rejected:
  `go install pkg@version` cannot take ldflags for the module's variables
  reliably, and users copy the plain command.
- *Fetch latest release at runtime* — rejected: network dependency,
  violates "boringly portable", and answers a different question.
- *Serve a version file next to the binary* — rejected: not how Go module
  installs work.

## D2 — Precedence: ldflags first, build info second, `dev` last

**Decision**: Resolution order: (1) ldflags `main.version` if ≠ `"dev"`;
(2) module version from build info when trustworthy (D3); (3) literal
`"dev"`. The `runetest`-only `RUNE_TEST_VERSION` override stays the
outermost wrapper, unchanged: `installedVersion(resolveVersion(version))`.

**Rationale**: Release artifacts (GoReleaser archives, Homebrew, Docker —
all stamp `-X main.version`) short-circuit at step 1, guaranteeing
byte-identical output (FR-002, SC-003) without touching the release
pipeline. Integration tests keep their env hook semantics (it already wins
over whatever the binary was stamped with, in `runetest` builds only).

## D3 — Trust boundary: only module-cache builds use the embedded version (FR-003)

**Decision**: The build-info fallback is used **only when the binary was
built from the module cache**, detected as: `Main.Version` is non-empty and
not `"(devel)"`, **and** the build has no `vcs.*` build settings. Checkout
builds (which carry `vcs.revision`/`vcs.modified` settings) keep reporting
`dev`.

**Rationale**: Verified empirically on this repo (Go ≥1.24 behavior): a
`go build` from a source checkout stamps `Main.Version` too — e.g.
`v0.4.4-0.20260812092550-432dbb8de56b+dirty` — and, at a clean tagged
commit, would stamp exactly `v0.4.3`, indistinguishable from a release by
version string alone. FR-003 requires local builds to never claim a release
version, and User Story 3's acceptance test demands checkout builds report
a development build. The presence of `vcs.*` settings is the precise,
toolchain-provided marker of "built from a checkout"; module-zip builds
(`go install mod@ver`) have none. This rule is also robust to
`GOFLAGS=-buildvcs=false` checkout builds: those stamp `(devel)` and fall
through to `dev` anyway.

**Alternatives considered**:
- *Report the stamped pseudo-version/`+dirty` for checkout builds* —
  rejected: a clean build at a tag claims the release version verbatim,
  violating FR-003; also changes contributor-facing output for no
  spec-driven reason.
- *Only special-case `"(devel)"`* — rejected: misses the clean-at-tag case
  above.

## D4 — Display normalization: trim the leading `v`

**Decision**: `strings.TrimPrefix(Main.Version, "v")` before reporting, so
`v0.4.3` renders as `0.4.3`.

**Rationale**: Today's output (`rune version 0.4.3 (commit 13d059a)`) has
no `v`; all install methods must present consistently (spec assumption).
Pseudo-versions from `go install mod@<untagged-commit>` render as
`0.4.4-0.20260812…-432dbb8de56b` — self-evidently not a clean release
(spec edge case), and `internal/semver` parses the prerelease form, so the
`minimum_version` gate compares it correctly (newer-than-required passes,
FR-004; dev-prerelease-of-X does not satisfy X, matching existing semver
tests).

## D5 — Commit field: unchanged, no VCS fabrication

**Decision**: `main.commit` keeps its ldflags/`"none"` behavior everywhere.
Module-cache builds have no `vcs.revision`, so a go-install binary reports
`rune version 0.4.3 (commit none)` — the existing graceful degradation
(FR-007).

**Rationale**: Release artifacts already stamp the short commit; inventing
a commit for module builds is impossible, and populating it for checkout
builds is out of scope (those stay `dev` per D3). Considered deriving a
short commit from pseudo-version suffixes — rejected: cosmetic, and blurs
the "release artifacts unchanged" guarantee.

## D6 — Gate integration: zero changes to `internal/`

**Decision**: No changes to `internal/cli/version_gate.go`,
`internal/config` or `internal/semver`. The fix is entirely "hand the true
version string to `cli.Options.Version`".

**Rationale**: `MinimumRequirement.Satisfied` already implements the full
policy: parseable versions are compared, unparseable ones (`dev`) are
dev-builds that bypass the gate. Once go-install binaries report `0.4.3`
instead of `dev`, FR-004/FR-006 (`rune version --check`, `--json`) follow
with no further code. Preserves the locked package layout (Principle IV).

## D7 — Test strategy: pure function + existing harness (Principle VI)

**Decision**: Factor the fallback as a pure function
`versionFromBuildInfo(version string, bi *debug.BuildInfo, ok bool) string`
(exact signature at implementation) so unit tests construct `BuildInfo`
values covering: stamped release (ldflags wins), module version, `(devel)`,
empty, pseudo-version, `vcs.*` present (checkout), `+dirty`, missing build
info. Binary-level behavior is already covered by
`test/integration/minimum_version_test.go` via `RUNE_TEST_VERSION`; those
tests re-run unchanged as the regression net (SC-004). A true end-to-end
`go install mod@tag` test is **not** added to CI.

**Rationale**: `debug.ReadBuildInfo()` cannot be faked at the process
level, so the seam is the function boundary — same pattern as the existing
`installedVersion` hook. An e2e go-install test would need network access
to the module proxy and a published tag inside the Docker test harness —
brittle and disproportionate; the quickstart documents it as a manual
release-verification step instead (V2).
