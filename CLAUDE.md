<!-- SPECKIT START -->
Active feature plan: `specs/013-secret-masking/plan.md` (Secret Masking &
Sanitization — values of sensitive environment variables are masked as `***`
in every output surface: task stdout/stderr, echoed command lines, Rune's own
status/log/error lines, and MCP tool results, so credentials never reach a
terminal transcript or an agent's chat history. Secrets are identified by
*name* over the effective environment (built-in case-insensitive patterns:
TOKEN, SECRET, PASSWORD, PASSWD, APIKEY, API_KEY, PRIVATE_KEY, ACCESS_KEY,
CREDENTIAL, AUTH) plus two new list settings — `set secrets := [...]` to
declare, `set unmasked := [...]` to exempt — resolved via the existing
`evalList` in `internal/config/settings.go` and registered in
`internal/language/builtin.go`. No content scanning, no parser/lexer changes.
A new leaf package `internal/mask` (stdlib-only; justified in the plan's
Complexity Tracking) provides the value `Set` (min length 4, multi-line values
split per line) and a concurrent streaming `io.WriteCloser` with bounded carry
for chunk-spanning values. It wraps `Options.Stdout/Stderr` once at engine
construction — the single choke point covering CLI, shell echo, agent
write-back, and the MCP adapter's buffers (`internal/cli/mcp.go`) — and is
skipped entirely when the set is empty, so secret-free Runefiles stay
byte-identical (golden corpus untouched). Masking is at emission time (an
interrupted task never leaks an unmasked window), verbatim occurrences only
(base64/URL-encoded transforms documented out of scope), always on with no
agent-facing off switch. Read the plan, `research.md` (decisions D1–D8),
`data-model.md`, `quickstart.md`, and `contracts/secret-masking.md` (which
extends the base MCP secrets contract) for details.
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
