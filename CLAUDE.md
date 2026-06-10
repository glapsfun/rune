<!-- SPECKIT START -->
Active feature plan: `specs/006-release-automation/plan.md` (Release Automation —
GoReleaser v2 + GitHub Actions tag-driven release: cross-platform binaries, multi-arch
GHCR images, cosign/SBOM/provenance, git-cliff changelog, Homebrew/Scoop, install script).
For additional context about technologies to be used, project structure, shell commands,
and other important information, read the current plan and its `research.md`.
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
