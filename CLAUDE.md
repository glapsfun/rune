<!-- SPECKIT START -->
Active feature plan: `specs/008-styled-cli-output/plan.md` (Styled CLI Output &
Friendlier Help — add a restrained semantic palette to non-interactive output
(`--list`, status/echo/cache, diagnostics) and redesign `--help`, without changing
any byte scripts/CI see except the deliberately-rewritten help. New leaf
`internal/style` package owns the palette + Lip Gloss `Theme` (plain when color
off); a new `--color=auto|always|never` flag drives a per-stream color decision
(`ColorStdout` for `--list`/`--help`, `ColorStderr` for status/diagnostics)
resolved in `cmd/rune/root.go`; diagnostic colors re-homed into the shared theme
with plain output byte-identical). For additional context about technologies,
project structure, shell commands, and constraints, read the current plan and its
`research.md`.
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
