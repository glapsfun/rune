# Research: VS Code Marketplace Extension for the Rune LSP

All Technical Context unknowns resolved. Each decision records what was
chosen, why, and what else was evaluated.

## D1 — Target registries: VS Code Marketplace + Open VSX

**Decision**: Publish to the Visual Studio Marketplace (primary, FR-001)
and the Open VSX registry (secondary, FR-009) with identical content and
version.

**Rationale**: The Microsoft Marketplace is contractually restricted to
Microsoft VS Code products, so VSCodium, Gitpod, Eclipse Theia, and other
compatible editors resolve extensions from Open VSX. Publishing both covers
effectively the whole VS Code-compatible install base for one extra CLI
invocation per release.

**Alternatives considered**: Marketplace-only (leaves VSCodium users on the
build-it-yourself path — rejected as cheap to avoid); Open VSX-only
(misses the overwhelming majority of users — rejected); GitHub-Releases
`.vsix` only (no discoverability, no auto-update — rejected as primary, but
the `.vsix` **is** attached to each GitHub release as a side-load fallback,
see D9).

## D2 — Publisher identity and one-time setup

**Decision**: Publisher/namespace `rune-task-runner` on both registries
(already declared in `editors/vscode/package.json`). One-time setup is a
documented maintainer runbook in `docs/releasing.md`, not automation:

- **Marketplace**: create the publisher at
  `marketplace.visualstudio.com/manage`, generate an Azure DevOps PAT with
  the **Marketplace → Manage** scope (org "All accessible organizations"),
  store it as the `VSCE_PAT` secret on the protected `release` environment.
- **Open VSX**: Eclipse Foundation account, sign the publisher agreement,
  `ovsx create-namespace rune-task-runner`, generate an access token, store
  as `OVSX_PAT` in the same environment.

**Rationale**: Namespace creation, agreement signing, and PAT minting are
inherently interactive, one-shot, and credential-bearing — a runbook is the
correct artifact (spec edge case "first publish"). Placing the secrets in
the `release` environment reuses feature 006's required-reviewer protection
(Constitution VII).

**Alternatives considered**: Personal publisher account (bus factor —
rejected); automating namespace creation in the workflow (impossible for
Marketplace, pointless one-time work for Open VSX — rejected).

## D3 — Packaging: `@vscode/vsce`, no bundler

**Decision**: Package with `@vscode/vsce` (pinned devDependency, invoked as
`npx vsce package`/`publish`) and publish to Open VSX with `ovsx publish`
reusing the same `.vsix`. Keep the client as plain CommonJS with its single
runtime dependency; **no esbuild/webpack bundling in v1**. Add a
`package-lock.json` (for reproducible `npm ci`) and a `.vscodeignore`
(exclude dev files so the `.vsix` stays small).

**Rationale**: `vsce package` already walks `npm list --production` and
includes runtime dependencies in the `.vsix`, so a bundler buys nothing for
a 40-line client with one dependency — it would only add a build step and a
new toolchain to a repo that prides itself on being boring. `ovsx publish
rune-<ver>.vsix` accepts the prebuilt artifact, guaranteeing byte-identical
content on both registries (FR-009).

**Alternatives considered**: esbuild single-file bundle (smaller `.vsix`,
but adds a build pipeline for negligible gain at this size — rejected for
v1, revisit if dependencies grow); publishing from source with `ovsx`
separately (risks content drift between registries — rejected).

## D4 — Versioning: stamp the release tag, don't commit it

**Decision**: The extension version is the Rune release tag without the
leading `v` (e.g. tag `v0.4.0` → extension `0.4.0`). `package.json` in the
repo keeps a `0.0.0` placeholder; the release job stamps the real version
with `npm version --no-git-tag-version "${TAG#v}"` before packaging and
never commits the result. Only stable tags publish (see D6), so the version
is always plain `major.minor.patch` — exactly what the Marketplace
requires.

**Rationale**: Satisfies FR-006 traceability (Marketplace version ==
release tag, bidirectionally) with zero release-time commits, no version
churn in the repo, and no risk of the changelog commit and version-bump
commit racing. The Marketplace rejects re-publishing an existing version,
which becomes a free idempotency guard: a re-run cannot silently
double-publish.

**Alternatives considered**: Independent extension versioning (breaks
traceability, invites drift — rejected); committing the bump each release
(extra bot commit + branch-protection friction for no benefit — rejected);
publishing rc tags as Marketplace pre-releases (Marketplace pre-release
semantics push odd/even minor conventions onto Rune's versioning; deferred
per spec assumption — stable-only in v1).

## D5 — Missing/old-binary UX: preflight in `activate()`

**Decision**: Before constructing the `LanguageClient`, the extension
resolves the configured executable (`rune.path`, default `rune`) and runs a
short preflight (`rune lsp --help`-class check via a spawn with timeout).
On failure it shows one `window.showErrorMessage` with actionable buttons —
"Install instructions" (opens the docs installation page) and "Open
Settings" (jumps to `rune.path`) — and does **not** start the client (no
crash-restart loop). An outdated binary whose `lsp` subcommand is missing
(pre-feature-011 installs) hits the same path with a "please upgrade"
wording.

**Rationale**: FR-003/SC-005 require 100% of missing-binary users to see
actionable guidance. `vscode-languageclient`'s default behavior on a
missing command is a generic "server crashed 5 times" toast — precisely the
cryptic failure the spec forbids. A preflight is ~20 lines in the existing
thin client and keeps all language intelligence in the binary
(Constitution IV's thin-shim spirit).

**Alternatives considered**: Auto-download the binary (explicitly out of
scope per spec assumption; platform-matrix + supply-chain surface —
rejected for v1); relying on the client library's error handler (message
quality too poor — rejected).

## D6 — Release integration: additive `publish-extension` job

**Decision**: Add a second job to `.github/workflows/release.yml`:
`publish-extension`, `needs: release`, gated
`if: ${{ !inputs.prerelease }}`. Steps: checkout the released tag →
setup Node 22 → `npm ci` → stamp version (D4) → `npx vsce package` →
upload `.vsix` to the GitHub release (D9) → `npx vsce publish --packagePath`
with `VSCE_PAT` → `npx ovsx publish` the same `.vsix` with `OVSX_PAT`. The
existing `release` job's steps stay byte-identical — its only change is a
job-level `outputs:` mapping exposing the computed tag
(`steps.ver.outputs.tag`) so the publish job can check it out via
`needs.release.outputs.tag`; a publish failure therefore can never un-ship
binaries, images, or the tag (FR-008, spec US2/AS2). Reusing the protected
`release` environment means required reviewers approve the publish job
separately (a second approval per release) — an accepted trade-off for
keeping the tokens behind the environment protection; documented in the
runbook.

**Rationale**: The release workflow already knows stable vs. rc via the
`prerelease` input, and the tag exists by the time the job runs — no new
versioning logic. Job-level separation gives free observability (red job in
the run) and free recovery ("Re-run failed jobs" re-executes only the
publish; the Marketplace duplicate-version rejection makes a re-run after a
partial success safe on the failed registry, see D8).

**Alternatives considered**: Steps appended inside the `release` job
(publish failure would mark the whole release red and block the digest/
attestation steps — rejected); a separate tag-triggered workflow (loses the
`prerelease` input and the protected-environment context; must re-derive
stability from tag shape — workable but strictly more logic, rejected);
GoReleaser `publishers`/extension hooks (GoReleaser has no first-class vsce
support; shell-escaping secrets into its config is worse — rejected).

## D7 — CI packaging smoke job

**Decision**: Add an `extension` job to `.github/workflows/ci.yml`:
Node 22, `npm ci`, `npx vsce package` (with a stamped throwaway version),
then assert the `.vsix` contains the grammar, `extension.js`, LICENSE, icon,
and README. Runs on every push/PR like the other gates.

**Rationale**: Constitution VI (multi-layer verification) and the
release-dryrun philosophy: nothing may be discovered broken on release day.
`vsce package` validates the manifest (icon exists, repository URL shape,
engines range, README present), so this job is the "extension dry-run"
twin of gate 7. It needs no Docker (not part of the Go suite; the
Docker-only rule governs `go test`).

**Alternatives considered**: Validating only in the release job (violates
test-first; a manifest typo would surface mid-release — rejected); full
extension integration tests via `@vscode/test-electron` (heavy harness for
a 40-line client whose behavior is pinned by the LSP server's own Go tests
— rejected for v1, the preflight logic gets a targeted scenario in
quickstart instead).

## D8 — Failure observability & recovery

**Decision**: Recovery procedure (documented in `docs/releasing.md`):

1. A failed `publish-extension` job shows red on the release run; the core
   release is unaffected (D6).
2. First resort: **Re-run failed jobs** on the same run — idempotent because
   each registry rejects duplicate versions, so an already-published side
   no-ops with a clear error and the failed side publishes.
3. Manual fallback (registry outage, expired PAT): download the `.vsix`
   from the GitHub release assets and run `npx vsce publish --packagePath`
   / `npx ovsx publish` locally with a fresh token — never a new Rune
   release.
4. Token rotation steps for `VSCE_PAT` (Azure DevOps PATs expire ≤ 1 year)
   and `OVSX_PAT` live in the same runbook section.

**Rationale**: FR-008 requires re-publish-without-re-release; spec edge
case requires handling one-registry-succeeded-one-failed. The
duplicate-version rejection converts naive re-runs into safe idempotent
retries.

**Alternatives considered**: A dedicated re-publish `workflow_dispatch`
workflow (more YAML to maintain for what "Re-run failed jobs" + a local
fallback already cover — rejected; can be added later if re-runs prove
awkward).

## D9 — `.vsix` attached to the GitHub release

**Decision**: The `publish-extension` job uploads `rune-<version>.vsix` as
a GitHub release asset (via `gh release upload`) **before** pushing to
either registry.

**Rationale**: Gives air-gapped/enterprise users a side-load path, gives the
manual recovery path (D8) its artifact, and preserves an auditable copy of
exactly what was published. Upload-before-publish means even a
both-registries outage still leaves the artifact recoverable.

**Alternatives considered**: GoReleaser `extra_files` (the `.vsix` doesn't
exist when GoReleaser runs in the `release` job — rejected); not attaching
(loses the recovery artifact — rejected).

## D10 — Listing quality: manifest and content fixes

**Decision**: Bring `editors/vscode/` up to Marketplace listing standards
(contract `marketplace-listing.md`):

- **Repository URL fix**: manifest currently points at
  `github.com/rune-task-runner/rune`; the canonical repo is
  `github.com/glapsfun/rune` — the wrong URL would break every listing
  link and vsce's repo-relative resource resolution. (Found during audit.)
- **Icon**: new 128×128 PNG (`icon` manifest field) — required for a
  non-placeholder listing tile.
- **LICENSE**: copy the repo MIT license into the extension folder (vsce
  warns and the listing shows "no license" otherwise).
- **README.md**: rewritten as the listing page — what you get (with the
  feature list from 011), the `rune` binary prerequisite up front with an
  install-docs link, settings table; the build-from-source section moves to
  a contributor note at the bottom (SC-006).
- **CHANGELOG.md**: thin file pointing at the repo `CHANGELOG.md` releases
  section (the listing's Changelog tab renders it).
- **Metadata**: `keywords` (`rune`, `runefile`, `task runner`, `lsp`),
  existing categories kept, `qna: false`? — no: leave Q&A default,
  `bugs`/`homepage` URLs added.
- **`.vscodeignore`**: exclude nothing needed, but drop dev artifacts.

**Rationale**: FR-004 (trust/evaluate metadata) and the spec edge case on
distinguishability (official repo link + publisher identity). All items are
static files inside `editors/vscode/` — zero engine impact.

**Alternatives considered**: Marketplace badges row in the README (external
badge hosts must be on the Marketplace trusted-badge allowlist; keep v1
badge-free — rejected for now).
