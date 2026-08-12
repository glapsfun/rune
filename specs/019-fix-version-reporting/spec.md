# Feature Specification: Accurate Version Reporting for Go-Toolchain Installs

**Feature Branch**: `019-fix-version-reporting`

**Created**: 2026-08-12

**Status**: Draft

**Input**: User description: "bug — not bump version in rune --version"

## Bug Summary

Rune's documented install instructions (root README, Getting Started, and
Installation docs) all offer `go install github.com/rune-task-runner/rune/cmd/rune@latest`
as a first-class install method. A binary installed this way reports
`rune version dev (commit none)` from `rune --version` and `rune version`,
even though it was built from a tagged release (e.g. v0.4.3). The version
appears "not bumped": users who upgrade via `go install` see `dev` forever,
no matter which release they actually have.

Only the official release artifacts (GitHub release archives, Homebrew,
Docker) receive a stamped version today. As a side effect, every
Go-toolchain-installed binary is treated as a development build by the
Runefile `minimum_version` compatibility gate, which then never blocks it —
an outdated install silently runs Runefiles that declare they need a newer
Rune.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Installed Version Is Reported Truthfully (Priority: P1)

A user installs Rune with the documented Go toolchain command (at `@latest`
or a pinned `@vX.Y.Z`). Running `rune --version` or `rune version` shows the
actual release version they installed — not `dev`.

**Why this priority**: This is the reported bug. Version identity is the
first thing support, bug reports, and upgrade decisions rely on; today an
entire documented install path cannot answer "which Rune do I have?".

**Independent Test**: Install Rune from a tagged release using the
documented Go toolchain command on a machine without the repo checked out,
run `rune --version`, and confirm the output names that release's version.

**Acceptance Scenarios**:

1. **Given** Rune was installed via the documented Go toolchain command at a
   tagged release (e.g. v0.4.3), **When** the user runs `rune --version`,
   **Then** the output reports version 0.4.3 (not `dev`).
2. **Given** Rune was installed the same way at a pinned older release,
   **When** the user runs `rune version`, **Then** the first line reports
   that pinned version and remains byte-identical to `rune --version`.
3. **Given** an official release artifact (GitHub archive, Homebrew, Docker),
   **When** the user runs `rune --version`, **Then** the output is unchanged
   from today: the release version plus the short commit.

---

### User Story 2 - Compatibility Gate Works for Go-Toolchain Installs (Priority: P2)

A team's Runefile declares `minimum_version` to protect contributors from
running an outdated Rune. A contributor who installed an older release via
the Go toolchain is correctly warned/blocked, instead of being waved through
as a "development build".

**Why this priority**: Direct consequence of the same defect. The
`minimum_version` safety net (feature 010) is silently defeated for the
whole `go install` install base, which can cause confusing downstream
failures the gate exists to prevent.

**Independent Test**: With a Runefile declaring a `minimum_version` above an
installed-via-toolchain older release, run any task and confirm the standard
incompatibility diagnostic appears and nothing executes.

**Acceptance Scenarios**:

1. **Given** a Go-toolchain install of release 0.4.0 and a Runefile with
   `minimum_version` of 0.5.0, **When** the user runs a task, **Then** Rune
   aborts with the standard version-mismatch diagnostic (installed version
   shown as 0.4.0) and nothing is executed.
2. **Given** the same install and a Runefile with `minimum_version` of 0.3.0,
   **When** the user runs a task, **Then** the gate passes silently and the
   task runs.
3. **Given** the same mismatched setup, **When** the user runs
   `rune version --check` (and `--check --json`), **Then** the report names
   the real installed version and states the requirement is not satisfied.

---

### User Story 3 - Development Builds Stay Honestly Labeled (Priority: P3)

A contributor building Rune from a local source checkout still sees a
version that is clearly a development build — it must never masquerade as a
released version.

**Why this priority**: Guards against over-correcting the bug. A local build
claiming to be "0.4.3" would corrupt bug reports and wrongly subject dev
builds to the `minimum_version` gate, breaking the existing contributor
workflow.

**Independent Test**: Build Rune from a source checkout (not a tagged module
install), run `rune --version`, and confirm the output identifies a
development build.

**Acceptance Scenarios**:

1. **Given** a binary built from a local source checkout, **When** the user
   runs `rune --version`, **Then** the output identifies it as a development
   build rather than claiming a release version.
2. **Given** a development build, **When** a Runefile declares any
   `minimum_version`, **Then** the existing behavior is preserved: the build
   is not blocked (development builds bypass the gate by design).

---

### Edge Cases

- Go toolchain install at `@latest` vs. a pinned `@vX.Y.Z`: both must report
  the version actually resolved and installed.
- Go toolchain install at a commit or pseudo-version (not a tagged release):
  the reported version must reflect what was installed and must not be
  mistaken for a clean release; the compatibility gate must not block a
  build that is genuinely newer than required.
- Local `go build` / `go run` from a checkout, with or without local
  modifications: remains a development build (Story 3).
- Commit identity: official artifacts keep showing their short commit; where
  the commit is unknowable the output degrades gracefully exactly as today
  (`commit none`) rather than showing a misleading value.
- The first line of `rune version` must remain byte-identical to
  `rune --version` in every install scenario (existing documented contract).
- Release pipeline unchanged: this fix must not require any change to how
  releases are cut, tagged, or published.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A Rune binary installed via the Go toolchain from a tagged
  release MUST report that release's version in `rune --version` and
  `rune version`, instead of a development placeholder.
- **FR-002**: Official release artifacts MUST continue to report exactly
  what they report today — release version plus short commit — with no
  change to output format or content.
- **FR-003**: Binaries built from a local source checkout MUST remain
  identifiable as development builds and MUST NOT claim a release version.
- **FR-004**: The Runefile `minimum_version` compatibility gate MUST
  evaluate the true installed version for Go-toolchain installs: older
  releases are blocked (or warned, under the existing ignore flag) with the
  standard diagnostic, and satisfying releases pass.
- **FR-005**: The first output line of `rune version` MUST remain
  byte-identical to `rune --version` across all install methods.
- **FR-006**: `rune version --check` and its machine-readable output MUST
  report the same true installed version used by the gate.
- **FR-007**: When the source commit is not known for an install method, the
  output MUST degrade gracefully as it does today rather than fabricate or
  omit the field inconsistently.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of installs performed with the documented Go toolchain
  command at a tagged release report that exact release version from
  `rune --version`.
- **SC-002**: With an older toolchain-installed release and a Runefile
  requiring a newer one, the incompatibility diagnostic appears and nothing
  executes — verified for both task runs and `rune version --check`.
- **SC-003**: For official release artifacts, version output is
  byte-identical before and after the fix.
- **SC-004**: Zero regressions in the existing version, version-gate, and
  release verification test suites; local-checkout builds still identify as
  development builds.

## Assumptions

- The released version number is derived from the release tag (tag minus the
  `v` prefix), and the version a Go-toolchain install should report is the
  module version it was resolved from.
- Displayed version formatting may normalize the leading `v` so that all
  install methods present the version number consistently (today's output
  has no `v` prefix).
- The existing design decision that development builds bypass the
  `minimum_version` gate is intentional and stays as-is; this fix only stops
  real releases from being misclassified as development builds.
- Distribution channels themselves (Homebrew formula, Docker images, GitHub
  release automation) are out of scope and unchanged.
- No new user-facing flags or commands are introduced; the fix changes what
  existing commands report.
