# Feature Specification: Release Automation

**Feature Branch**: `006-release-automation`

**Created**: 2026-06-10

**Status**: Draft

**Input**: User description: "i want to create relis proces to destrbute app — artefact must be to compile to arch (linux = x86, arm), (macos = x86, arm) also win = x86, arm and docker image for x86 and arm. relise must create tag with fix version template, version example v0.0.1. also automation add change to change log and overall apply best practice for release that type of app"

## Clarifications

### Session 2026-06-10

- Q: Commits are free-form today; how should changelog notes be categorized into Added/Changed/Fixed? → A: Adopt Conventional Commits enforced on squash-merge PR titles; the changelog is auto-grouped from those titles.
- Q: Where should the Homebrew formula and Scoop manifest be hosted? → A: Dedicated public tap & bucket repos the project owns (e.g. `homebrew-tap`, `scoop-bucket`), auto-updated by each stable release via a push token.

### Session 2026-06-11

- Q: FR-002 said the next version is computed from "the most recent release tag," which contradicts the shipped behavior; how should the version lifecycle read? → A: Compute from the most recent **stable** tag (ignoring pre-releases). The same change kind then drives the whole lifecycle — with pre-release selected, repeated runs iterate `…-rc.1`, `…-rc.2`, … of the same target; without pre-release, the run promotes that target to its stable version.

## User Scenarios & Testing *(mandatory)*

This feature serves two audiences:

- **Maintainers** who cut releases and need the process to be one safe, repeatable action.
- **Consumers** (end users, CI systems, and AI agents) who download, verify, install, and run a released version on their platform.

### User Story 1 - Cut a versioned, cross-platform binary release (Priority: P1)

A maintainer decides the project is ready to ship. They initiate a release and choose the kind of change (major, minor, or patch). The system computes the next version following the `vMAJOR.MINOR.PATCH` template (e.g. `v0.0.1`), creates and publishes the version tag, builds a self-contained binary for every supported operating system and architecture, packages each with license and readme, generates a checksums file, and publishes them all as a single immutable release that consumers can download.

**Why this priority**: This is the core of the request and the minimum viable release process. Without it there is no reliable way to distribute the app. Every other story builds on the artifacts and tag produced here.

**Independent Test**: From a clean repository state on the release branch, initiate a patch release and confirm that (a) a new `vX.Y.Z` tag exists, (b) downloadable binaries exist for all six OS/architecture combinations, (c) a checksums file covers every artifact, and (d) a released binary reports the exact released version when asked for its version.

**Acceptance Scenarios**:

1. **Given** the latest release is `v0.3.2` and the working state is clean on the release branch, **When** the maintainer initiates a "minor" release, **Then** a `v0.4.0` tag is created and published and a release containing all platform binaries is produced.
2. **Given** no prior release tag exists, **When** the maintainer initiates the first release, **Then** the version baseline is established and a complete release is produced.
3. **Given** a release has been produced, **When** a consumer downloads the artifact for their platform and runs it, **Then** the program reports the exact released version.
4. **Given** a release has been produced, **When** a consumer recomputes the checksum of any downloaded artifact, **Then** it matches the published checksums file.

---

### User Story 2 - Publish multi-architecture container images (Priority: P2)

As part of the same release, the system builds and publishes a container image that runs on both x86-64 and arm64 hosts under one image reference. The image is tagged with the exact version and with a moving "latest" pointer, so consumers can pull either a pinned version or the newest stable release without caring which CPU architecture they are on.

**Why this priority**: Container distribution was explicitly requested and is a primary consumption path for CI and agent environments. It depends on the version/tag from P1 but is otherwise independent.

**Independent Test**: After a release, pull the versioned image reference on an x86-64 host and on an arm64 host; confirm the correct architecture is served automatically and the container reports the released version.

**Acceptance Scenarios**:

1. **Given** a release `v0.4.0` has been cut, **When** a consumer pulls the versioned image on an arm64 machine, **Then** the arm64 variant is served and runs.
2. **Given** the same release, **When** a consumer pulls the versioned image on an x86-64 machine, **Then** the x86-64 variant is served and runs.
3. **Given** a stable release is published, **When** a consumer pulls the "latest" image reference, **Then** they receive the most recent stable version.
4. **Given** a pre-release version, **When** it is published, **Then** the "latest" pointer is **not** moved to it.

---

### User Story 3 - Automated changelog (Priority: P3)

Each release records what changed. The system derives the change list from the commit and pull-request history since the previous release, writes it into a committed `CHANGELOG.md` (grouped into Added / Changed / Fixed / etc. with the version and date), and publishes the same notes on the release page. Maintainers do not hand-assemble release notes.

**Why this priority**: Explicitly requested and central to a trustworthy release, but it depends on a release/tag existing to scope "since last release."

**Independent Test**: Cut a release after a few merged changes and confirm a new dated version section appears in the committed `CHANGELOG.md`, the same notes appear on the release page, and entries link back to the relevant tag/commits — with no manual editing required.

**Acceptance Scenarios**:

1. **Given** several changes merged since `v0.3.2`, **When** `v0.4.0` is released, **Then** `CHANGELOG.md` gains a `[0.4.0]` section dated with the release date and grouped by change type.
2. **Given** a release is published, **When** a consumer views the release page, **Then** it shows the same change notes as the `CHANGELOG.md` section.
3. **Given** a changelog section is generated, **When** a reader opens it, **Then** the version links to its tag and to a comparison against the previous version.

---

### User Story 4 - Verifiable, tamper-evident artifacts (Priority: P4)

Every release artifact — binary archives and container images — can be verified by anyone, with no pre-shared secret. The system publishes checksums, cryptographic signatures, a software bill of materials (SBOM), and build provenance that ties each artifact to the exact source commit and build. Consumers (including automated agents) can confirm an artifact is authentic and unmodified before trusting it.

**Why this priority**: Supply-chain integrity is a best practice the maintainer selected and is increasingly expected of a tool that both humans and AI agents execute. It hardens the artifacts from P1/P2 but is not required for a first usable release.

**Independent Test**: Download a release artifact and its signature/SBOM/provenance; verify the signature and provenance using only publicly available information and confirm verification succeeds for an untampered artifact and fails for a modified one.

**Acceptance Scenarios**:

1. **Given** a published release, **When** a consumer verifies an artifact's signature using only public material, **Then** verification succeeds.
2. **Given** a published release, **When** a consumer tampers with an artifact and re-verifies, **Then** verification fails.
3. **Given** a published release, **When** a consumer inspects an artifact, **Then** an SBOM and build provenance linking it to the source commit are available.
4. **Given** a published container image, **When** a consumer verifies it, **Then** its signature, SBOM, and provenance are present and valid.

---

### User Story 5 - Frictionless installation (Priority: P5)

Beyond raw downloads, consumers can install the released version through the path that suits them: a Homebrew formula (macOS/Linux), a Scoop manifest (Windows), or a one-line install script that detects their operating system and architecture, downloads the correct artifact, verifies its checksum, and places the binary on their path. These installation channels are kept current automatically as part of each stable release.

**Why this priority**: Convenience installers maximize adoption and are a selected best practice, but they layer on top of the artifacts and verification from earlier stories.

**Independent Test**: On each supported platform, install the released version via at least one documented channel and confirm the installed program reports the released version.

**Acceptance Scenarios**:

1. **Given** a stable release, **When** a macOS user installs via the package manager channel, **Then** the released version is installed and runnable.
2. **Given** a stable release, **When** a Windows user installs via the package manager channel, **Then** the released version is installed and runnable.
3. **Given** a stable release, **When** any user runs the one-line install script, **Then** the correct artifact for their OS/architecture is downloaded, checksum-verified, and installed.
4. **Given** a pre-release version, **When** it is published, **Then** the stable installation channels are **not** updated to point at it.

---

### Edge Cases

- **First release**: No prior tag exists — the version baseline must be established cleanly and the changelog must cover history from the project start.
- **Re-running after partial failure**: If a release fails partway (e.g., binaries published but images not), re-running must converge to a complete release without manual cleanup or duplicate tags.
- **Version collision**: If the computed/target version tag already exists, the release must refuse rather than overwrite.
- **Unclean or wrong source**: A release initiated from a non-release branch or a dirty working tree must be refused.
- **Quality gates failing**: A release must not be publishable from a commit that does not pass the project's quality gates.
- **Empty change set**: If there are no user-facing changes since the last release, the changelog handling must produce a sensible result (and the maintainer is warned) rather than an empty or broken section.
- **One platform fails to build**: A failure to produce any required artifact must fail the release atomically (no partial publish) and report which target failed.
- **Pre-release vs stable**: Pre-release versions must be clearly marked and must not move the "latest" image pointer or update stable install channels.
- **Missing or expired publishing credentials**: Absent registry/signing/publishing credentials must produce a clear, early failure, not a silently incomplete release.
- **Local dry run**: A maintainer must be able to produce all artifacts locally without publishing, to validate a release before cutting it for real.

## Requirements *(mandatory)*

### Functional Requirements

#### Versioning & tagging

- **FR-001**: Release versions MUST follow semantic versioning with a leading `v` (`vMAJOR.MINOR.PATCH`, e.g. `v0.0.1`).
- **FR-002**: A maintainer MUST be able to initiate a release by selecting the change kind (major, minor, or patch) and whether it is a pre-release; the system MUST compute the next version from the most recent **stable** release tag (ignoring pre-release tags) plus the chosen change kind.
- **FR-003**: The system MUST create and publish the version tag automatically as part of the release; the maintainer MUST NOT have to hand-create the tag.
- **FR-004**: The system MUST refuse to (re)release a version whose tag already exists, and MUST NOT produce duplicate tags.
- **FR-005**: The system MUST support pre-release versions (e.g. `v1.2.0-rc.1`) and MUST mark them as pre-releases. Because the target is derived from the most recent **stable** tag, the same change kind drives the whole lifecycle: with pre-release selected, repeated runs MUST iterate `…-rc.1`, `…-rc.2`, … of the same target; with pre-release not selected, the run MUST promote that target to its stable version (`vMAJOR.MINOR.PATCH`).
- **FR-006**: Each released artifact MUST report the exact released version when asked (version embedded at build time, including the source revision).

#### Build artifacts

- **FR-007**: Each release MUST produce a self-contained binary for all six targets: Linux (x86-64, arm64), macOS (x86-64, arm64), and Windows (x86-64, arm64).
- **FR-008**: Binaries MUST be self-contained with no external runtime dependency, consistent with the project's portability commitment.
- **FR-009**: Each binary MUST be packaged in an OS-appropriate archive (compressed tarball for Unix-like systems, zip for Windows) that includes the license and readme.
- **FR-010**: Each release MUST publish a single multi-architecture container image (Linux x86-64 and arm64) addressable by one reference, tagged with the exact version and — for stable releases — a moving "latest" tag.
- **FR-011**: All published artifacts MUST be identifiable with the release version.

#### Changelog

- **FR-012**: Each release MUST add a dated, version-headed section to a committed `CHANGELOG.md` that follows the "Keep a Changelog" convention and groups entries by change type (Added / Changed / Fixed / etc.).
- **FR-013**: Changelog entries MUST be derived automatically from Conventional-Commit-formatted squash-merge pull-request titles since the previous release, without manual assembly; the commit type (`feat`, `fix`, etc.) determines the change-type group.
- **FR-013a**: The project MUST adopt the Conventional Commits convention for squash-merge pull-request titles and MUST enforce it with an automated check on pull requests, so that changelog categorization is reliable.
- **FR-014**: The release page notes MUST present the same change information as the corresponding `CHANGELOG.md` section.
- **FR-015**: Each changelog version MUST link to its tag and to a comparison against the previous version.

#### Distribution & publishing

- **FR-016**: Binaries, archives, and the checksums file MUST be attached to a published release for the tag.
- **FR-017**: Container images MUST be pushed to a container registry from which consumers can pull by version or by "latest".
- **FR-018**: A package-manager install path MUST be maintained for macOS/Linux (Homebrew) and Windows (Scoop) via dedicated tap and bucket repositories the project owns; each stable release MUST automatically update the formula and manifest in those repositories (the maintainer does not hand-edit them).
- **FR-019**: A one-line install script MUST be provided that detects the consumer's OS and architecture, downloads the correct artifact, verifies its checksum, and installs the binary onto the path.
- **FR-020**: Stable installation channels (latest image tag, package-manager manifests, install-script default) MUST NOT be updated to point at a pre-release.

#### Integrity & supply chain

- **FR-021**: A checksums file (SHA-256) covering every release artifact MUST be published.
- **FR-022**: Release artifacts (binaries/archives and container images) MUST be cryptographically signed such that consumers can verify them using only publicly available material (no pre-shared secret).
- **FR-023**: A software bill of materials (SBOM) MUST be produced and attached for the binaries and the container image.
- **FR-024**: Build provenance/attestation MUST be generated that links each artifact to the exact source commit and build event, and MUST be verifiable by consumers.

#### Process safety & operability

- **FR-025**: A release MUST only proceed from a clean state on the designated release branch; releases from a dirty tree or other branches MUST be refused.
- **FR-026**: A release MUST NOT be publishable from a commit that has not passed the project's quality gates (lint, cross-OS tests, build, documentation checks, and release dry-run).
- **FR-027**: The system MUST provide a non-publishing dry-run that produces all artifacts locally for verification before a real release.
- **FR-028**: A failed or partial release MUST be reported clearly and MUST be safe to re-run to completion (idempotent), without manual cleanup.
- **FR-029**: Only authorized maintainers MUST be able to publish a release.
- **FR-030**: The repository MUST document how to cut a release and how consumers install and verify artifacts across all channels.

### Key Entities *(include if feature involves data)*

- **Release**: An immutable, versioned publication tied to one tag; comprises a set of artifacts, a checksums file, signatures/SBOM/provenance, and a changelog section + release notes.
- **Version tag**: The `vMAJOR.MINOR.PATCH` (optionally `-prerelease`) identifier that names a release and ties it to a source commit.
- **Artifact**: A single distributable deliverable — a platform binary archive, a container image, a checksums file, a signature, an SBOM, or a provenance document.
- **Changelog**: The ordered, human-readable record of releases and their grouped changes, maintained in-repo and mirrored on release pages.
- **Distribution channel**: A path through which consumers obtain a release — release downloads, container registry, Homebrew, Scoop, or the install script.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A maintainer can cut a complete release (tag created through all artifacts published) from a single initiating action, with no manual editing of versions, tags, or notes.
- **SC-002**: 100% of supported platforms are covered every release — working downloads for all six OS/architecture binary combinations plus a container image that runs on both x86-64 and arm64.
- **SC-003**: A consumer on any supported platform can go from "nothing installed" to running the released version in under 2 minutes using at least one documented channel.
- **SC-004**: 100% of untampered release artifacts pass independent verification (checksum and signature) performed by a third party with no pre-shared secret, and tampered artifacts fail verification.
- **SC-005**: 100% of releases have an accurate changelog section and matching release notes generated without manual authoring.
- **SC-006**: Every released artifact reports the exact released version when queried; reported versions match the tag 100% of the time.
- **SC-007**: Zero releases are publishable from a commit that fails the project's quality gates.
- **SC-008**: Re-running a release after a partial failure converges to a complete, single release with no duplicate tags and no manual cleanup.
- **SC-009**: Pre-release versions never update the stable "latest" image tag or stable install channels (0 incidents).

## Assumptions

- **Architecture mapping**: "x86" means 64-bit (x86-64 / amd64) and "arm" means 64-bit (arm64). 32-bit targets are out of scope.
- **Hosting**: Binaries are published as repository releases and container images to the repository's container registry (the project is hosted on GitHub; the existing container image already references the GitHub repository). Specific registry naming is an implementation detail of planning.
- **Foundation already in place**: The app already builds as a single self-contained, CGO-free Go binary for the six targets, and a baseline cross-compilation config and a minimal distroless container build already exist. This feature wires those into an automated, tag-driven release pipeline and adds changelog generation, signing/SBOM/provenance, and the install channels. It also operationalizes the constitution's existing "release-dryrun" quality gate.
- **History convention**: The project adopts Conventional Commits, enforced on squash-merge pull-request titles via an automated PR check, so changelog entries can be categorized automatically. Existing free-form history before adoption is summarized into the first managed changelog section rather than retro-categorized.
- **Release source**: Releases are cut from the `main` branch after continuous-integration gates pass; the maintainer chooses the version bump (the process is maintainer-initiated, not auto-cut from commits).
- **Signing model**: Signing uses a keyless / transparency-log approach so no long-lived private signing key must be managed.
- **Container scope**: Container images are Linux multi-arch only (Windows/macOS container images are out of scope, as is standard for this tool class).
- **Changelog format**: `CHANGELOG.md` follows the "Keep a Changelog" convention.

## Dependencies

- A source repository with releases, a container registry, and continuous integration enabled.
- Publishing credentials available to the automation with permission to push tags, create releases, push container images, and sign artifacts.
- Two dedicated public repositories the project owns for the Homebrew tap and Scoop bucket, plus a token with push access to them so the release can update the formula/manifest automatically. (One-time setup; updates are automatic thereafter.)
- The project's existing quality gates (lint, cross-OS tests, build, documentation checks, release dry-run) as the precondition for any release.

## Out of Scope

- 32-bit architectures and non-amd64/arm64 targets.
- Linux distribution packages (`.deb` / `.rpm`) and package managers beyond Homebrew and Scoop (e.g., apt, dnf, nix, winget) — possible future work.
- Fully automatic version inference and release-on-merge (continuous release) — the maintainer chooses the bump.
- Platform code-signing/notarization that requires paid certificates (Apple Developer ID notarization, Windows Authenticode) — possible future work to remove OS "unidentified developer" warnings.
- GUI installers / native installer packages (`.msi`, `.pkg`, `.dmg`).
