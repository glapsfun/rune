# Feature Specification: VS Code Marketplace Extension for the Rune LSP

**Feature Branch**: `015-vscode-lsp-extension`

**Created**: 2026-07-29

**Status**: Draft

**Input**: User description: "DEPLOY LSP VSCODE TO markect extantion — over all i want install lsp for rune form marcetplase extantions"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Install Rune language support from the Marketplace (Priority: P1)

A developer who uses Rune opens VS Code, searches for "Rune" in the
Extensions view, and installs the official Rune extension with one click.
When they open a `Runefile`, they immediately get syntax highlighting,
diagnostics, completion, go-to-definition, hover, document symbols, and
formatting — powered by the `rune` executable already installed on their
machine. No cloning the repository, no building a package by hand, no
side-loading a file.

**Why this priority**: This is the entire point of the feature. Today the
extension exists in the repository but can only be used by building and
side-loading it manually, which almost no end user will do. Marketplace
availability is what turns the existing LSP work (feature 011) into
something users can actually adopt.

**Independent Test**: On a clean machine with VS Code and the `rune` binary
installed, search the Marketplace for "Rune", install the extension, open a
Runefile with a known error, and confirm a diagnostic appears — without ever
touching the Rune source repository.

**Acceptance Scenarios**:

1. **Given** a user with VS Code and the `rune` executable on their PATH,
   **When** they search the VS Code Marketplace for "Rune" and install the
   official extension, **Then** the extension installs successfully and its
   listing clearly identifies it as the official Rune task runner extension.
2. **Given** the extension is installed, **When** the user opens a file
   named `Runefile`, `.runefile`, or `*.rune`, **Then** syntax highlighting
   is applied and the language server starts automatically, providing
   diagnostics and other language features.
3. **Given** the extension is installed but the `rune` executable is not
   found, **When** the user opens a Runefile, **Then** the extension shows a
   clear, actionable message explaining that the `rune` binary is required
   and how to install it or point the extension at it (via the existing
   `rune.path` setting), instead of failing silently.

---

### User Story 2 - Extension stays current with each Rune release (Priority: P2)

A Rune maintainer cuts a new release. As part of the established release
process, an updated version of the VS Code extension is published to the
Marketplace automatically, so users receive the update through VS Code's
normal extension auto-update mechanism without any manual packaging or
upload steps by the maintainer.

**Why this priority**: Without automated publishing, the Marketplace listing
rots after the first upload. Rune already has a fully automated release
pipeline (feature 006); the extension must ride it, not add a manual
chore that will be forgotten.

**Independent Test**: Perform a release (or a release dry run) and verify
that a new extension version is produced and published (or, in dry-run mode,
packaged and validated) with a version number consistent with the release,
with no manual intervention beyond the normal release trigger.

**Acceptance Scenarios**:

1. **Given** a new Rune release is triggered through the normal release
   process, **When** the release completes, **Then** a matching new version
   of the extension is published to the Marketplace without manual steps.
2. **Given** the extension publish step fails (e.g., credential or service
   error), **When** the release runs, **Then** the failure is clearly
   visible to maintainers and the recovery path is documented, and the
   failure does not corrupt or roll back the rest of the release.
3. **Given** a published extension version, **When** a user inspects it,
   **Then** its version is traceable to the Rune release that produced it.

---

### User Story 3 - Install on VS Code-compatible editors (Priority: P3)

A developer using a VS Code-compatible editor that cannot access the
Microsoft Marketplace (e.g., VSCodium or other open-source builds, which use
the Open VSX registry instead) can also install the Rune extension from
their editor's default extension registry.

**Why this priority**: A meaningful share of developers use non-Microsoft
builds that are legally barred from the Microsoft Marketplace. Publishing to
the open registry is low incremental effort once Marketplace publishing
exists, but it is not required for the core "install from marketplace"
outcome.

**Independent Test**: In VSCodium (or via the Open VSX website), search for
"Rune", install the extension, and confirm the same language features work.

**Acceptance Scenarios**:

1. **Given** a user of a VS Code-compatible editor backed by the Open VSX
   registry, **When** they search for "Rune" and install the extension,
   **Then** they get the same extension version and functionality as
   Marketplace users.

---

### Edge Cases

- What happens when the `rune` executable is missing, or is an older version
  that does not support the `lsp` command? The extension must surface a
  clear, actionable message (install/upgrade guidance, `rune.path` setting)
  rather than a silent failure or a cryptic crash loop.
- What happens when another extension already claims the same name or a
  similar identifier on the Marketplace? The official listing must be
  discoverable and distinguishable (publisher identity, description,
  repository link).
- What happens if publishing succeeds on one registry but fails on the other
  (Marketplace vs. Open VSX)? Maintainers must be able to see the partial
  failure and re-publish the missing side without cutting a new release.
- What happens when a user has the extension installed but works in a
  workspace with no Runefile? The extension must stay dormant (no server
  process, no errors) until a Rune file is opened.
- What happens on the very first publish, where the publisher account and
  extension namespace must exist before automation can work? One-time
  account/namespace setup steps must be documented for maintainers.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The official Rune VS Code extension MUST be published to the
  Visual Studio Code Marketplace under a publisher identity controlled by
  the Rune project, so end users can find and install it by searching for
  "Rune".
- **FR-002**: The published extension MUST provide the full existing feature
  set of the in-repo extension client: Runefile language registration and
  syntax highlighting, plus LSP-backed diagnostics, completion,
  go-to-definition, hover, document symbols, and formatting via the user's
  installed `rune` executable.
- **FR-003**: The extension MUST NOT bundle the `rune` binary; it MUST use
  the `rune` executable from the user's system (respecting the existing
  `rune.path` setting) and MUST present a clear, actionable notification
  when the executable cannot be found or cannot serve as a language server.
- **FR-004**: The Marketplace listing MUST include the metadata users need
  to trust and evaluate the extension: display name, description, icon,
  license, repository link, and installation prerequisites (the `rune`
  binary requirement) in the listing README.
- **FR-005**: Publishing a new extension version MUST be integrated into the
  existing automated release process, so each Rune release that changes the
  extension (or on every release — see Assumptions) publishes an updated
  extension version without manual packaging or upload.
- **FR-006**: Extension versions MUST be traceable to the Rune release that
  produced them, and the versioning scheme MUST satisfy Marketplace version
  ordering rules so updates are offered to users automatically.
- **FR-007**: Publishing credentials/tokens MUST be stored and used only via
  the project's existing secret-management practices for release automation,
  never committed to the repository.
- **FR-008**: A failed publish MUST be observable by maintainers and MUST be
  recoverable by re-running the publish step for the same version, without
  requiring a new Rune release; the recovery procedure MUST be documented.
- **FR-009**: The extension SHOULD also be published to the Open VSX
  registry with the same version and content, so users of VS
  Code-compatible editors that cannot use the Microsoft Marketplace can
  install it (User Story 3).
- **FR-010**: Maintainer documentation MUST cover the one-time setup
  (publisher account/namespace creation, token provisioning) and the
  ongoing per-release behavior, and user documentation MUST be updated to
  present Marketplace installation as the primary installation path for VS
  Code (replacing the current build-it-yourself instructions).

### Key Entities

- **Extension package**: The installable artifact containing the language
  client, grammar, and metadata; versioned, published to one or more
  registries; depends at runtime on the user's `rune` executable.
- **Marketplace listing**: The public page (name, publisher, description,
  icon, README, license, repository link) through which users discover and
  install the extension; exists once per registry (VS Code Marketplace,
  Open VSX).
- **Publisher identity**: The project-controlled account/namespace under
  which the extension is published; owns the listing and the publish
  credentials.
- **Release**: The existing Rune release event that triggers packaging and
  publication of a new extension version and determines its version number.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user on a machine with VS Code and the `rune` binary
  installed can go from "searching the Marketplace" to "working diagnostics
  in an open Runefile" in under 5 minutes with no steps outside VS Code.
- **SC-002**: Installing the extension requires zero interaction with the
  Rune source repository (no clone, no build, no side-load).
- **SC-003**: After a Rune release completes, the updated extension version
  is available on the Marketplace within the same release cycle (no
  separate manual publish step performed by a maintainer).
- **SC-004**: 100% of extension versions published after this feature ships
  are traceable to a specific Rune release.
- **SC-005**: When the `rune` executable is absent, 100% of users see an
  actionable guidance message rather than a silent failure (verified by
  test).
- **SC-006**: The documented VS Code installation path in user-facing docs
  is the Marketplace install; manual build instructions remain only as a
  development/contributor workflow.

## Assumptions

- The existing in-repo extension client (`editors/vscode/`, from feature
  011) is functionally complete for v1 of the listing; this feature is about
  packaging, publishing, listing quality, and release integration — not new
  language features.
- The extension will NOT bundle platform-specific `rune` binaries in v1; it
  requires a separately installed `rune` executable, consistent with how the
  extension works today. Bundling or auto-downloading the binary is out of
  scope and could be a future feature.
- The publisher identity will be owned by the Rune project organization
  (the extension manifest already declares a `rune-task-runner` publisher
  name); securing that namespace on both registries is part of the one-time
  setup.
- Extension versions will follow the Rune release version, and every Rune
  release publishes an extension version (even if only the embedded server
  changed), keeping the two version streams aligned and simple to reason
  about.
- Publishing to Open VSX (User Story 3) is in scope as a SHOULD; if
  namespace acquisition there is blocked, the feature can still ship with
  Marketplace-only publishing.
- The existing release automation and secret-management infrastructure
  (feature 006) is the vehicle for publish automation; no new release
  system is introduced.
- Pre-release Rune versions (release candidates) do not need to publish
  Marketplace pre-release extension versions in v1; only stable releases
  publish.
