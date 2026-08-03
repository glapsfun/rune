# Contract: Marketplace Listing & Extension Manifest

Binding rules for what users see and install. Verified by the CI smoke
job's content assertions and quickstart scenarios V1–V3.

- **L1 — Identity.** Extension ID is `rune-task-runner.rune`; display name
  "Rune"; description identifies it as the official language support for
  Runefiles by the Rune task runner project.

- **L2 — Publisher.** Published only under the project-controlled
  `rune-task-runner` publisher (Marketplace) / namespace (Open VSX).

- **L3 — Icon.** `icon.png` (128×128) referenced from the manifest `icon`
  field and included in the package.

- **L4 — Listing README.** `editors/vscode/README.md` is the listing body.
  It MUST state the `rune` binary prerequisite before any feature content,
  link to the canonical installation docs, document `rune.path` and
  `rune.trace.server`, and demote build-from-source to a contributor note.

- **L5 — License.** The package contains a `LICENSE` file (MIT, copied from
  the repo root) and the manifest declares `"license": "MIT"`.

- **L6 — Canonical URLs.** `repository`, `bugs`, and `homepage` MUST point
  at `https://github.com/glapsfun/rune` (fixing the current
  `rune-task-runner/rune` placeholder). No dead links on the listing.

- **L7 — Changelog tab.** The package contains a `CHANGELOG.md` that
  directs readers to the repository changelog/releases.

- **L8 — Functional parity.** The published extension provides exactly the
  in-repo client's features (FR-002): language registration for
  `Runefile`/`.runefile`/`*.rune`, TextMate grammar, and LSP features via
  the user's `rune` executable. No `rune` binary is bundled (FR-003).

- **L9 — Missing-binary UX.** When the configured executable is missing or
  lacks the `lsp` subcommand, the extension shows a single actionable error
  notification (install/upgrade guidance, buttons to the install docs and
  the `rune.path` setting) and does not start or crash-loop the client.
  When no Runefile-language document is opened, the extension does not
  activate at all.
