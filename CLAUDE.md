<!-- SPECKIT START -->
Active feature plan: `specs/007-interactive-tui/plan.md` (Interactive Task Picker —
replace the `--choose` fzf/numbered picker with a modern, styled in-terminal task
picker built on Bubble Tea + Bubbles + Lip Gloss; opt-in via `--choose`, tears down
and hands off to the existing `execute()` path on selection, non-interactive/CI
paths unchanged; new pure `internal/tui` package wired from `internal/cli/choose.go`).
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
