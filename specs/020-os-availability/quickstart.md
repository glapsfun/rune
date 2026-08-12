# Quickstart Validation: OS Availability Enforcement

**Feature**: 020-os-availability

End-to-end checks proving the feature works. Contracts referenced:
[availability.md](contracts/availability.md),
[mcp-surface.md](contracts/mcp-surface.md),
[dump-schema.md](contracts/dump-schema.md).

## Prerequisites

- Docker running (tests execute inside the harness — never on the host)
- `go` toolchain for building the host binary

```sh
go build -trimpath -o dist/rune ./cmd/rune   # or: rune build
```

## Fixture

Save as `Runefile` in a scratch directory (examples below assume a
linux/macOS host; on Windows the roles invert):

```rune
# Cross-platform dispatcher.
setup: setup-nix setup-win
    @echo "setup done"

# Unix-only half.
[unix]
setup-nix:
    @echo "unix setup"

# Windows-only half.
[windows]
setup-win:
    @echo "windows setup"
```

## Scenario 1 — Listing (regression, SC-005)

```sh
dist/rune --list
```

**Expected**: `setup` and `setup-nix` listed; `setup-win` absent
(unchanged behavior).

## Scenario 2 — Direct invocation hard-errors (FR-003, SC-002)

```sh
dist/rune setup-win; echo "exit=$?"
```

**Expected**: no output from the task; stderr names the task, `windows`,
and the host OS; `exit=3`. Also verify the multi-root rule:

```sh
dist/rune setup setup-win; echo "exit=$?"
```

**Expected**: nothing executes (no "unix setup"/"setup done"), `exit=3`.

## Scenario 3 — Dependency dispatch (FR-004, SC-003)

```sh
dist/rune setup
```

**Expected**: prints `unix setup` then `setup done`; no error, exit 0,
no mention of `setup-win`.

## Scenario 4 — Dump verdict (FR-005)

```sh
dist/rune --dump --format json | grep -A1 '"name": "setup-win"'
```

**Expected**: `setup-win` present with `"available": false`; `setup` and
`setup-nix` carry `"available": true`.

## Scenario 5 — MCP tool list (FR-002, SC-001)

```sh
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"qs","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  | dist/rune mcp | grep -o '"name":"[a-z-]*"'
```

**Expected**: `setup` and `setup-nix` appear; `setup-win` does not.
(Exact JSON-RPC framing may need adjusting to the mcp-go version; the
automated MCP adapter test is authoritative.)

## Automated suites (authoritative gate)

```sh
docker-compose run --rm test go test ./...
docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...
docker-compose run --rm test go test ./test/docs/...   # docs claims stay true
golangci-lint run
```

**Expected**: all green, including the new `internal/ast` availability
table tests (SC-004), scheduler skip tests, and CLI/MCP/dump tests.
