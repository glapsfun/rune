# Feature Specification: Best-Practices Refactor — Structure, Docs, CI, and Docker

**Feature Branch**: `002-best-practices-refactor`

**Created**: 2026-06-08

**Status**: Draft

**Input**: User description: "i want to do refactoring and improve project with best practisc. structura project as basest practis. creat docs how to use. add ci. improve docker."

## Overview

This is a **project-hardening / non-functional** feature. It improves how Rune is
maintained, documented, verified, and distributed — without changing what Rune *does*.
Runefile semantics, CLI contracts, and observable behavior remain backward-compatible
(see the constitution's tool backward-compatibility promise). The work spans five themes
the user requested: best-practice **refactoring**, idiomatic **project structure**,
usage **documentation**, a strengthened **CI** quality bar, and improved **Docker**
distribution.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Adopt Rune from documentation (Priority: P1)

A developer who has never seen Rune lands on the repository, reads a clear README, learns
what Rune is and why they'd use it, installs it on their platform, writes a tiny Runefile,
and successfully runs their first task — all without reading source code or asking anyone.

**Why this priority**: The project currently has **no README and no usage documentation**
(only `docs/GRAMMAR.md`). Without onboarding docs, the tool is effectively unadoptable by
anyone outside the original author. This unlocks every other value the project offers, so
it is the highest-value standalone slice.

**Independent Test**: Hand the documentation to someone unfamiliar with the project and
have them install Rune and run a documented quickstart task on a clean machine. Success =
they reach a running task using only the docs.

**Acceptance Scenarios**:

1. **Given** a newcomer on the repository landing page, **When** they read the README,
   **Then** they can state in one sentence what Rune is, who it's for, and how it differs
   from `make`/`just`, and find a link to install and quickstart instructions.
2. **Given** the installation docs, **When** the user follows the instructions for their
   platform (Linux, macOS, or Windows), **Then** they obtain a working `rune` binary and
   can confirm it with a version/help command.
3. **Given** the quickstart, **When** the user copies the example Runefile and runs the
   documented command, **Then** the task executes and produces the documented output.
4. **Given** the usage/reference docs, **When** the user looks up a CLI command or a
   Runefile language construct (dependencies, parameters, caching, executors, dotenv, MCP),
   **Then** they find at least one runnable example for it.

---

### User Story 2 - Trust every change via constitutional CI gates (Priority: P1)

A contributor opens a pull request. Automated checks run the project's **full**
constitutional quality bar and give fast, specific pass/fail feedback, so unsafe or
non-conforming changes never reach `main`.

**Why this priority**: Rune is a tool other teams' CI pipelines depend on; silent
regressions are expensive. The existing pipeline is missing constitutional gates —
race detection, `gofumpt`/`goimports`, generated-artifact (golden) consistency, and a
build check — so it does not yet enforce the quality the constitution requires. Closing
those gaps is foundational and independently valuable.

**Independent Test**: Open one PR that deliberately violates each gate (bad formatting, a
lint error, a failing test, a data race, a stale golden file) and one clean PR. Success =
each violating PR is blocked with a message naming the failed gate; the clean PR passes.

**Acceptance Scenarios**:

1. **Given** any push or pull request, **When** CI runs, **Then** it verifies formatting
   (`gofmt` + `gofumpt` + `goimports`), `go vet`, and `golangci-lint`, and fails if any is
   unclean.
2. **Given** the test stage, **When** it runs, **Then** the full suite executes on Linux,
   macOS, and Windows **with the race detector enabled** and must pass with zero races.
3. **Given** the fuzz targets, **When** CI runs, **Then** they build and smoke-run.
4. **Given** generated artifacts (AST/token/formatter golden files, and any generated
   docs), **When** CI runs, **Then** drift between source and committed artifacts fails the
   build with a "regenerate X" message — golden files are never hand-edited to pass.
5. **Given** the supported target platforms, **When** CI runs, **Then** the `rune` binary
   builds successfully for each.
6. **Given** any failing gate, **When** a contributor views the result, **Then** merging is
   blocked and the failure clearly identifies which gate failed and how to fix it.

---

### User Story 3 - Run Rune with zero local install via Docker (Priority: P2)

A user who doesn't want to install Go (or anything) pulls an official, minimal Rune
container image, mounts their project, and runs their tasks. Maintainers keep using an
improved, documented Docker-based test harness.

**Why this priority**: Distribution convenience and reproducibility. Today Docker exists
only as a *test harness* — there is no runtime image users can pull to run Rune. This is
high value but depends on nothing else, so it can ship independently after docs/CI.

**Independent Test**: Build/pull the production image on a machine with no Go toolchain,
run `rune` against a mounted Runefile, and confirm it executes the task identically to a
native run. Separately, run the test suite via the documented Docker harness.

**Acceptance Scenarios**:

1. **Given** a machine with no Go/toolchain installed, **When** the user runs the official
   Rune image against a mounted working directory, **Then** the documented task runs and
   produces the same result as a native `rune` invocation (default `sh` executor parity).
2. **Given** the production image, **When** it is built, **Then** it is a minimal,
   multi-stage image containing the static `rune` binary and no build toolchain.
3. **Given** a task that needs an external runtime (`python`/`node`/agent) **and** the
   minimal image that lacks it, **When** the user runs it, **Then** they get a clear,
   documented error explaining the runtime is not present (not a confusing crash).
4. **Given** the test harness, **When** a maintainer runs the documented Docker command,
   **Then** the full suite runs inside the container (never on the host) as project policy
   requires.

---

### User Story 4 - Maintain a best-practice repository structure and code (Priority: P2)

A contributor (human or AI) finds a repository whose layout matches the documented
idiomatic Go structure, whose code conforms to the project's Go engineering discipline,
and which contains the standard hygiene files an open project is expected to have.

**Why this priority**: Maintainability and contributor onboarding. The layout is already
close to the constitution's; this story closes remaining gaps (missing `LICENSE`,
contributor guide, any lint/discipline violations surfaced by the stricter gates) and
performs a behavior-preserving cleanup pass.

**Independent Test**: Run the full lint/discipline gate and a structure/hygiene audit on a
clean checkout. Success = zero violations, required hygiene files present, and the existing
behavioral test suite still passes unchanged (no golden regenerations needed).

**Acceptance Scenarios**:

1. **Given** the repository layout, **When** it is audited against the constitution's
   documented package layout (`cmd`, `internal/*`, public `mcpserver`), **Then** every
   package is in its prescribed place or the deviation is explicitly justified.
2. **Given** the codebase, **When** the full `golangci-lint`/discipline gate runs, **Then**
   it reports zero violations (Principle VIII).
3. **Given** a behavior-preserving refactor, **When** the existing behavioral, integration,
   golden, and corpus tests run, **Then** they pass **without** regenerating golden files to
   accommodate changed behavior.
4. **Given** the release configuration that references `LICENSE*` and `README*`, **When**
   the repository is inspected, **Then** those files exist so the references resolve.

---

### User Story 5 - Produce a complete release with one tag (Priority: P3)

A maintainer tags a version and automated tooling produces the full, reproducible release
artifact set — cross-platform binaries with checksums and the published container image —
with no manual assembly.

**Why this priority**: This extends the CI and Docker asks into distribution. A
release configuration already exists but is not wired to a trigger, and there is no
container publish. Valuable but the lowest priority and the most optional of the five.

**Independent Test**: Perform a release tag (or a dry-run) and confirm the expected
artifact set is produced and references resolve, without any manual file assembly.

**Acceptance Scenarios**:

1. **Given** a release tag, **When** release automation runs, **Then** it produces
   cross-platform binaries (Linux/macOS/Windows, amd64/arm64) plus a checksums file.
2. **Given** a release, **When** it completes, **Then** the production container image is
   published to the project's registry tagged with the release version.
3. **Given** the release archives, **When** they are inspected, **Then** each includes the
   `LICENSE` and `README`.
4. **Given** a release dry-run with a missing referenced file, **When** it runs, **Then** it
   fails before publishing anything.

---

### Edge Cases

- **Race detector vs. static binary conflict**: race detection requires a C toolchain
  (CGO enabled), while the shipped artifact is a CGO-free static binary. CI must run race
  tests in a CGO-enabled context **separately** from the CGO-disabled release build. If a
  platform genuinely cannot run `-race`, that exception MUST be explicit and logged, never
  silently dropped.
- **Generated-artifact drift**: a committed golden file, a `gofumpt`-formatted file, or a
  generated doc diverges from source — CI must fail with a clear "regenerate X" instruction
  rather than passing silently.
- **Minimal image missing runtimes**: a user runs a `python`/`node`/agent task in the
  minimal container that lacks those interpreters — the failure must be clear and
  documented, with guidance (use native install or a fuller image).
- **Behavior-changing refactor**: a cleanup unintentionally changes an exit code, stdout,
  or a diagnostic span — behavioral/golden tests must catch it and block merge; any
  intentional behavior change must be opt-in per the backward-compatibility promise.
- **Toolchain version drift**: CI's Go version disagrees with `go.mod`/the constitution
  minimum — the inconsistency (currently CI `1.26` vs `go.mod` `1.25`) must be reconciled.
- **Stale documentation examples**: a CLI/DSL change makes a documented example wrong — a
  docs-example check (e.g., running the example) should surface the drift.
- **Cross-platform docs**: documented path examples must use forward-slash path-join so
  they hold on Windows (Principle V).

## Requirements *(mandatory)*

### Functional Requirements

**Documentation (US1)**

- **FR-001**: The project MUST provide a root `README` stating what Rune is, who it is for,
  how it compares to `make`/`just`, and linking to install, quickstart, and reference docs.
- **FR-002**: Documentation MUST provide installation instructions for each supported
  platform, covering at least: prebuilt binary, build-from-source, and container.
- **FR-003**: Documentation MUST include a quickstart that takes a new user from zero to a
  running task, with a copy-pasteable example Runefile and the exact command to run it.
- **FR-004**: Documentation MUST cover usage of every CLI command/flag and every Runefile
  language construct at a how-to level (dependencies, parameters, caching opt-in,
  executors, dotenv, formatting), linking to `docs/GRAMMAR.md` for the formal grammar.
- **FR-005**: Documentation MUST explain how to expose tasks to AI agents via MCP, including
  the secure-by-default behavior (read-only default, env-only secrets, destructive opt-in)
  required by the constitution.
- **FR-006**: Documentation MUST be kept consistent with the shipped CLI/DSL; a check or
  defined process MUST flag drift between docs examples and actual behavior.

**CI quality gates (US2)**

- **FR-007**: Every push and pull request MUST automatically run the full quality-gate set
  before a change can merge.
- **FR-008**: The gate set MUST include formatting verification using `gofmt`, `gofumpt`,
  and `goimports`, plus `go vet` and `golangci-lint`, each required to pass.
- **FR-009**: The test suite MUST run on Linux, macOS, and Windows.
- **FR-010**: The test suite MUST run with the race detector enabled and pass with zero
  detected races.
- **FR-011**: Fuzz targets for the lexer and parser MUST build and smoke-run in CI.
- **FR-012**: CI MUST verify that committed generated artifacts (golden files, formatter
  output, and any generated docs) match what regeneration would produce; any drift MUST
  fail the build.
- **FR-013**: CI MUST verify the `rune` binary builds for every supported target platform.
- **FR-014**: The CI Go toolchain version MUST be consistent with `go.mod` and the
  constitution's minimum; the existing version inconsistency MUST be resolved.
- **FR-015**: A failing gate MUST block merge and present a clear, actionable failure that
  names the failed gate.

**Docker distribution (US3)**

- **FR-016**: The project MUST provide an official production container image that runs the
  `rune` binary.
- **FR-017**: The production image MUST be produced by a multi-stage build yielding a
  minimal image based on the static binary, with no build toolchain in the final image.
- **FR-018**: Documentation MUST explain how to run Rune via the container, including
  mounting a working directory / Runefile and passing task arguments.
- **FR-019**: Containerized Rune MUST behave identically to native for the default `sh`
  executor; limitations for `python`/`node`/agent executors in the minimal image MUST be
  documented, and such tasks MUST fail with a clear message rather than crashing obscurely.
- **FR-020**: The Docker-based test harness MUST be retained, improved, and documented as
  the supported way to run the test suite (tests run in-container, never on the host).

**Structure & refactor (US4)**

- **FR-021**: The repository structure MUST conform to the constitution's documented
  idiomatic Go layout; any deviation MUST be corrected or explicitly justified.
- **FR-022**: The codebase MUST conform to the Go engineering discipline (Principle VIII);
  violations surfaced by the stricter gates MUST be fixed.
- **FR-023**: All refactoring MUST be behavior-preserving — no change to Runefile semantics,
  CLI contracts, exit codes, or diagnostic spans. Any behavior change MUST be opt-in and
  explicitly called out per the backward-compatibility promise.
- **FR-024**: The repository MUST include the standard hygiene files it is expected to have
  and that the release configuration references — at minimum a `LICENSE` and a contributor
  / development guide — so all such references resolve.

**Releases (US5)**

- **FR-025**: A release tag MUST automatically produce cross-platform binaries
  (Linux/macOS/Windows on amd64 and arm64) accompanied by a checksums file.
- **FR-026**: A release MUST publish the production container image to the project's
  container registry, tagged with the release version.
- **FR-027**: Release archives MUST include the `LICENSE` and `README`; a release with an
  unresolved referenced file MUST fail before publishing.

### Key Entities *(deliverable artifacts manipulated by this feature)*

- **Usage documentation set**: README + getting-started/quickstart + CLI & Runefile usage
  reference + MCP usage guide; complements the existing `docs/GRAMMAR.md`.
- **CI quality-gate pipeline**: the automated checks (format, vet, lint, cross-platform
  race tests, fuzz smoke, golden consistency, build) gating every change.
- **Container images**: the minimal production runtime image and the (improved) test-harness
  image.
- **Release bundle**: cross-platform binaries, checksums, published container image, and
  archive contents (LICENSE/README).
- **Repository hygiene files**: `LICENSE`, contributor/development guide, and any layout
  corrections.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A person unfamiliar with the project can install Rune and run their first
  task in under 10 minutes using only the documentation, on a clean machine.
- **SC-002**: 100% of pushes and pull requests trigger automated verification covering
  formatting, linting, cross-platform tests on all three supported OSes, race detection,
  fuzz smoke, generated-artifact consistency, and a build.
- **SC-003**: A change that violates any quality gate is blocked from merging 100% of the
  time, and the result names the specific gate that failed.
- **SC-004**: The full test suite passes with the race detector enabled on Linux, macOS,
  and Windows with zero detected data races.
- **SC-005**: A user can run Rune with no local language/toolchain installed via a single
  container command, and the published production image is under 30 MB.
- **SC-006**: Every artifact the release configuration references exists in the repository
  (zero unresolved references), verified by a successful release dry-run.
- **SC-007**: A maintainer can produce a complete cross-platform release (binaries +
  checksums + container image) by performing a single tag action, with no manual artifact
  assembly.
- **SC-008**: The repository passes the full lint/engineering-discipline gate with zero
  violations.
- **SC-009**: The behavior-preserving refactor changes no observable behavior — the entire
  existing behavioral, integration, golden, and corpus test suite passes unchanged, with no
  golden files regenerated to accommodate changed behavior.
- **SC-010**: Every CLI command and every Runefile language construct has at least one
  documented, runnable usage example.

## Assumptions

- **Non-functional scope**: This feature does not change what Rune does. Runefile
  semantics, CLI contracts, exit codes, and diagnostics remain backward-compatible; all
  refactoring is behavior-preserving.
- **CI platform**: The existing GitHub Actions pipeline (`.github/workflows/ci.yml`) is
  enhanced in place rather than replaced or migrated to another CI system.
- **Release tooling & registry**: Release automation builds on the existing GoReleaser
  configuration; the production container image is published to the project's GitHub
  Container Registry namespace.
- **Documentation form**: "Docs" means in-repository Markdown (README + `docs/` pages +
  the existing `docs/GRAMMAR.md`). A separately hosted documentation website is out of
  scope for this iteration.
- **Minimal image executor scope**: The minimal production image targets the default `sh`
  executor. `python`/`node`/agent executors that require external runtimes are out of scope
  for the minimal image and are documented as a known limitation.
- **Race detector execution**: Race-detector test runs use CGO-enabled toolchains in CI and
  are kept separate from the CGO-free static release build; the shipped single binary
  remains CGO-free and statically linked.
- **Supported platforms**: Linux, macOS, and Windows on amd64 and arm64, consistent with
  the existing release configuration and the constitution.
- **Go toolchain**: Go 1.25+ per `go.mod` (constitution minimum 1.24+); the current CI/
  go.mod version inconsistency is reconciled toward a single supported version.
- **Test policy**: The existing policy that the test suite runs inside Docker (never on the
  host) is retained.

## Out of Scope

- New Rune features, language constructs, executors, or CLI commands.
- Any change to Runefile semantics or observable runtime behavior (the backward-compat
  promise forbids non-opt-in changes).
- A hosted documentation website, versioned docs site, or docs search.
- Embedding language runtimes (python/node) into the minimal default image.
- Package-manager distribution channels (Homebrew, apt, Scoop, etc.) beyond what the
  existing release configuration already provides.
- Performance optimization work (governed separately by Principle VIII; requires
  benchmarks).
