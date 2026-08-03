<!-- SPECKIT START -->
Active feature plan: `specs/015-vscode-lsp-extension/plan.md` (VS Code
Marketplace Extension for the Rune LSP — publish the existing
`editors/vscode/` client (feature 011) to the VS Code Marketplace and Open
VSX under the `rune-task-runner` publisher, so users install Rune language
support from the Extensions view instead of building a `.vsix` by hand.
Packaging/automation only: zero Go changes, the `rune lsp` server is
untouched. Touched surfaces (S1–S13 in `data-model.md`): manifest hardening
in `editors/vscode/package.json` (fix the wrong `rune-task-runner/rune`
repository URL → `glapsfun/rune`, add icon/license/keywords/bugs/homepage,
`0.0.0` placeholder version stamped from the release tag at publish time and
never committed), new `package-lock.json`, LICENSE copy, `icon.png`,
`CHANGELOG.md`, `.vscodeignore`, a missing/old-binary preflight in
`extension.js` (actionable notification, no crash loop, client not started),
listing-grade README rewrite, a CI `extension` packaging smoke job, a new
`publish-extension` job in `release.yml` (`needs: release`, stable-only via
`!inputs.prerelease`, order: npm ci → stamp → vsce package → attach `.vsix`
to the GitHub release → `vsce publish` → `ovsx publish` the same file;
secrets `VSCE_PAT`/`OVSX_PAT` from the protected `release` environment), and
docs (releasing runbook with one-time publisher setup + recovery,
`editors/README.md`, root README). Hard constraints: the core `release`
job's steps stay byte-identical (it only gains an `outputs:` mapping
exposing the tag) and a publish failure never blocks it (P1); no bundled
`rune` binary (FR-003); extension version == tag minus `v` (FR-006);
duplicate-version rejection makes job re-runs idempotent recovery (P6). Read
the plan, `research.md` (decisions D1–D10), `data-model.md` (S1–S13),
`quickstart.md` (V1–V6), and `contracts/` (`publishing-pipeline.md` P1–P8,
`marketplace-listing.md` L1–L9) for details.
<!-- SPECKIT END -->

## Development workflow

Rune dogfoods itself: the repo-root `Runefile` defines the dev tasks. Run `rune --list`
(or `go run ./cmd/rune --list`) to see them — `fmt`, `lint`, `test`, `test-race`, `build`,
`docker`, `docs-check`, `release-dryrun`.

Tests run **inside Docker**, never on the host (per global policy and the lack of a compose
plugin — use standalone `docker-compose`):

```sh
docker-compose run --rm test go test ./...
docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...
```

See `CONTRIBUTING.md` for the full workflow and CI gate set.
