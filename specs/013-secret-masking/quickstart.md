# Quickstart: Validating Secret Masking & Sanitization

**Feature**: `013-secret-masking` | **Phase**: 1

Runnable scenarios proving the feature end-to-end. Details live in
[contracts/secret-masking.md](contracts/secret-masking.md) and
[data-model.md](data-model.md) — this guide only exercises them.

## Prerequisites

- Repo checked out on `013-secret-masking`; Docker + standalone
  `docker-compose` available (tests never run on the host).
- A build of the CLI: `go run ./cmd/rune` (or `rune build` via dogfooding).

## Scenario 1 — Agent-style env dump is masked (spec US1 / SC-002)

```sh
mkdir -p /tmp/rune-mask-demo
cat > /tmp/rune-mask-demo/Runefile <<'EOF'
leak:
    @echo "token is $API_TOKEN"
    @env | grep API_TOKEN
EOF

API_TOKEN=hunter2-super-secret go run ./cmd/rune --file /tmp/rune-mask-demo/Runefile leak
```

**Expected**: every occurrence prints as `token is ***` / `API_TOKEN=***`;
the string `hunter2-super-secret` appears nowhere. Exit 0.

MCP end-to-end (what an agent actually receives) — the stdio MCP server is the
`serve` subcommand (alias `mcp`) and needs the initialize handshake first:

```sh
API_TOKEN=hunter2-super-secret go run ./cmd/rune serve --file /tmp/rune-mask-demo/Runefile <<'EOF'
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"quickstart","version":"0"}}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"leak"}}
EOF
```

(The integration test's `runWithStdin` frames in
`test/integration/harness_test.go` are the authoritative reference for the
exact handshake.)

**Expected**: the JSON tool result contains `***` and never the raw value.

## Scenario 2 — Echoed command lines (spec US2)

```rune
deploy:
    curl -H "Authorization: Bearer {{env("API_TOKEN")}}" https://example.test
```

Run without `@` so the command echoes. **Expected**: the echoed line on stderr
shows `Bearer ***`.

## Scenario 3 — Declarations and exemptions (spec US3)

```rune
set secrets := ["DEPLOY_CFG"]
set unmasked := ["OAUTH_METHOD"]

show:
    @echo "$DEPLOY_CFG / $OAUTH_METHOD"
```

```sh
DEPLOY_CFG=s3://bucket/key?sig=abc OAUTH_METHOD=oauth2-pkce go run ./cmd/rune show
```

**Expected**: `*** / oauth2-pkce` — declared name masked despite its innocent
name; pattern-matching name (`AUTH`) exempted.

## Scenario 4 — Static validation (contract §1)

```sh
printf 'set secrets := [42]\nok:\n    @true\n' > /tmp/bad.rune
go run ./cmd/rune --file /tmp/bad.rune ok; echo "exit=$?"
```

**Expected**: positioned diagnostic (`file:line:col` + caret), `exit=3`,
nothing executed. Also: `set secert := ["X"]` flagged RUNE2008 by
`go run ./cmd/rune analyze`.

## Scenario 5 — No-secrets byte invariance (SC-003)

```sh
docker-compose run --rm test go test ./internal/cli/... ./test/corpus/... ./test/integration/...
```

**Expected**: all existing golden and styling tests pass **unmodified** —
Runefiles with empty mask sets take the unwrapped writer path.

## Scenario 6 — Chunk-spanning and multi-line values (FR-004)

Covered by unit tests; spot-check:

```sh
docker-compose run --rm test go test ./internal/mask/... -run 'Chunk|Multiline|Overlap|Concurrent' -v
```

**Expected**: secrets split across `Write` calls, PEM-style multi-line values,
nested values, and concurrent writers all mask correctly.

## Scenario 7 — Performance guardrail (SC-004)

```sh
docker-compose run --rm test go test ./internal/mask/... -bench 'Writer' -benchtime 10x
```

**Expected**: masked throughput within 10% of an unmasked baseline on a 10 MB
stream (benchmark compares both paths).

## Full gate set (before any commit)

```sh
docker-compose run --rm test go test ./...
docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...
go run ./cmd/rune lint      # golangci-lint via dogfood task
go run ./cmd/rune docs-check
```

**Expected**: all green, including `docs-verify` fixtures for the new
`docs/examples/secret-masking/` example and updated `docs/GRAMMAR.md`.
