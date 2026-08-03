# Contract: Extension Publishing Pipeline

Binding rules for the `publish-extension` release job and the CI smoke job.
Verified by CI runs, release dry behavior, and quickstart scenarios V4–V6.

- **P1 — Additive only.** The existing `release` job's **steps** in
  `.github/workflows/release.yml` remain byte-identical. The only permitted
  change to that job is adding a job-level `outputs:` mapping that exposes
  the computed tag (`steps.ver.outputs.tag`) to downstream jobs. The new
  `publish-extension` job declares `needs: release` and reads the tag via
  `needs.release.outputs.tag`; its failure MUST NOT fail, roll back, or
  block any core release artifact (binaries, archives, images, tag,
  changelog commit, attestations).

- **P2 — Stable-only gate.** `publish-extension` runs iff the dispatch
  input `prerelease` is false. No extension version is ever published for a
  `-rc.N` tag.

- **P3 — Version stamping.** The job checks out the tag created by
  `release`, stamps `editors/vscode/package.json` with the tag minus the
  leading `v` (`npm version --no-git-tag-version`), and commits nothing.
  The published extension version MUST equal the release tag minus `v`.

- **P4 — Artifact order.** The job MUST (a) `npm ci`, (b) `vsce package`,
  (c) upload the `.vsix` to the GitHub release for that tag, (d) publish
  the **same file** to the Marketplace via `vsce publish --packagePath`,
  (e) publish the **same file** to Open VSX via `ovsx publish`. Attachment
  (c) precedes (d)/(e) so a registry outage still leaves the artifact
  recoverable.

- **P5 — Secrets.** Registry credentials enter only as environment
  variables from the `release` environment secrets (`VSCE_PAT`,
  `OVSX_PAT`). They MUST NOT appear in the repository, in workflow-file
  literals, or in logs.

- **P6 — Idempotent recovery.** Re-running the failed job for the same tag
  MUST be safe and MUST be able to end green: both publish steps pass
  `--skip-duplicate`, so an already-published side is a logged no-op while
  the other side publishes. Accidental re-publication of a different
  artifact under the same version stays impossible (registries are
  append-only per version, and the release job refuses existing tags).
  Recovery never requires a new Rune release (documented in
  `docs/releasing.md`).

- **P7 — Partial-failure visibility.** Marketplace and Open VSX publishes
  are separate steps so a one-registry failure is identifiable from the run
  UI. Open VSX failure MUST NOT retroactively affect an already-successful
  Marketplace publish (and vice versa).

- **P8 — CI smoke gate.** `ci.yml` gains an `extension` job that runs on
  every push/PR: Node 22 → `npm ci` → stamp a throwaway version (plain
  `major.minor.patch`, e.g. `0.0.1` — vsce rejects prerelease-suffixed
  versions) → `npx vsce package` → assert the `.vsix` contains `extension.js`, the
  TextMate grammar, `language-configuration.json`, README, CHANGELOG,
  LICENSE, icon, and `node_modules/vscode-languageclient`. A manifest or
  packaging regression MUST fail CI, not release day.
