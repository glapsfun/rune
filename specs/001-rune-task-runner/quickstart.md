# Quickstart & Validation Guide: Rune

**Feature**: 001-rune-task-runner | **Date**: 2026-06-08

End-to-end validation scenarios that prove the feature works. Each scenario maps to a user
story / success criterion. Implementation details live in `data-model.md` and `contracts/`;
this guide is about **running and observing**.

## Prerequisites

- Go 1.24+ (verified toolchain: go1.26.2).
- For multi-language scenarios: `python3` and `node` on PATH.
- For the AI-agent scenario: an installed, authenticated agent CLI (`claude`, `codex`, or
  `copilot`).

## Build

```bash
CGO_ENABLED=0 go build -o ./dist/rune ./cmd/rune
export PATH="$PWD/dist:$PATH"
rune --version          # expect: version string, exit 0
```

## Scenario 1 — Author & run tasks (US1, P1)

Create `Runefile`:

```rune
set default := "greet"

# Say hello.
greet name="world":
    @echo "hello, {{name}}"

build: greet
    @echo "building..."
```

Validate:

| Command | Expected |
|---------|----------|
| `rune` | runs default `greet` → `hello, world`, exit 0 |
| `rune greet Ada` | `hello, Ada` |
| `rune build` | `greet` runs first (once), then `building...`; exit 0 |
| `rune --list` | shows `greet` (with its doc) and `build`; runs nothing |

Proves: discovery, default task, params+defaults, dependencies run-once, listing (FR-001/2/3/4/5/22).

## Scenario 2 — Errors caught before running (US2, P2)

```rune
a: b            # b does not exist
c: c            # self-cycle
greet:
    @echo {{undefined_var}}
```

| Command | Expected |
|---------|----------|
| `rune a` | error "unknown task: b" with `Runefile:line:col` + caret; **exit 3**; nothing runs |
| `rune c` | error "dependency cycle: c → c"; exit 3 |
| `rune greet` | error "undefined variable: undefined_var" with span; exit 3 |

Proves: static validation, located diagnostics, no side effects on error (FR-012/13/14, SC-002).

## Scenario 3 — Multi-language bodies (US3, P3)

```rune
build_dir := "dist"

analyze (python):
    print("coverage from {{build_dir}}")

bundle (node):
    console.log("bundling " + "{{build_dir}}")
```

| Command | Expected |
|---------|----------|
| `rune analyze` | runs under python3 → `coverage from dist` |
| `rune bundle` | runs under node → `bundling dist` |
| (rename to a missing interpreter) | actionable "interpreter not found" error; exit 1 |

Proves: executor selection, interpolation into bodies, missing-interpreter handling (FR-016/17, US3).

## Scenario 4 — Shared with AI agents (US4, P4)

```rune
# Show recent git log.
logs:
    @git log --oneline -5

[confirm("Really clean?")]
clean:
    rm -rf dist
```

Server (local stdio):

```bash
rune mcp        # starts MCP server on stdio
```

From an MCP client (or `rune serve --mcp --http --addr 127.0.0.1:7777 --token-file ./tok`):

| Check | Expected |
|-------|----------|
| list tools | `logs` present with its doc + empty input schema; `clean` present with `destructiveHint: true` |
| call `logs` | returns git log output + exitCode 0 (same engine as CLI) |
| call `clean` without approval | refused / requires approval |
| HTTP call without token | rejected (SC-010) |
| any tool description | contains no secret values (SC-007) |

Agent task (drives an installed agent CLI):

```rune
set agent_cmd := ["claude", "-p"]

triage (agent):
    Summarize the last 5 commits. You may call the `logs` task. Do not modify files.
```

| Command | Expected |
|---------|----------|
| `rune triage` | drives the agent CLI; agent may call `logs`; final text becomes task output; exit 0 |
| (agent CLI absent/unauthenticated) | actionable error naming the tool; exit 1 |

Proves: task→tool exposure, destructive gating, transport/auth, agent executor (FR-025…31, Q1–Q3).

## Scenario 5 — CI/CD ergonomics (US5, P5)

```rune
[cache(inputs = ["go.mod", "**/*.go"], outputs = ["dist/rune"])]
build-cached:
    CGO_ENABLED=0 go build -o dist/rune ./cmd/rune

[parallel]
checks: lint test
lint:
    @echo lint
test:
    @echo test
```

| Command | Expected |
|---------|----------|
| `rune build-cached` (1st) | `running: build-cached`; builds; writes `.rune/cache/…` |
| `rune build-cached` (2nd, no changes) | `cached: build-cached`; skipped; < 10% of 1st run time (SC-006) |
| touch a `.go` file, re-run | `running:` again |
| `rune checks` | `lint` and `test` run concurrently |
| `rune --dry-run build-cached` | prints plan + would-be cache decision; runs nothing |
| `rune --dump --format json` | emits structured parse of the Runefile |

Proves: opt-in caching with visible decisions, parallel fan-out, dry-run, machine output
(FR-015/19/20/22/23, SC-006/9).

## Test suite expectations (Principle VI)

```bash
go test ./...                 # unit + golden + integration (compiled-binary) tests pass
go test -run=xxx -fuzz=FuzzLexer -fuzztime=10s ./internal/lexer
go test -run=xxx -fuzz=FuzzParser -fuzztime=10s ./internal/parser
```

CI runs the above on Linux, macOS, and Windows; the compatibility corpus re-parses known
Runefiles to guard against silent grammar drift.
