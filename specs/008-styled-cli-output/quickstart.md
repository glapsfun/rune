# Quickstart: validating styled CLI output

All Go commands run **inside Docker** (repo policy):

```sh
docker-compose run --rm test go test ./...
docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...
```

Build a binary for manual checks:

```sh
docker-compose run --rm test go build -o /tmp/rune ./cmd/rune
```

## 1. Unit: theme degrades to plain (internal/style)

```sh
docker-compose run --rm test go test ./internal/style/...
```
Expect: with `enabled=false`, every role's `Render(s)` returns `s` unchanged
(no ANSI); with `enabled=true`, output contains SGR escapes but identical visible
text/width.

## 2. Golden: plain output unchanged when color is OFF (the guardrail)

```sh
docker-compose run --rm test go test ./internal/cli/... ./internal/diag/... ./test/integration/...
```
Expect: `--list`, status/echo/cache, and diagnostic output are byte-for-byte
identical to the committed goldens when piped, with `NO_COLOR=1`, and with
`--color=never`. The diagnostic golden `testdata/diag/render.golden` (color off)
is unchanged. Regenerate goldens only deliberately (never hand-edit):
```sh
docker-compose run --rm test go test ./internal/diag/... -run Golden -update
```

## 3. Manual: color on a TTY vs off

```sh
# Styled (real terminal):
/tmp/rune --list

# Plain — piped stdout is not a TTY:
/tmp/rune --list | cat            # no ANSI, columns aligned

# Plain — explicit off on a TTY:
/tmp/rune --color=never --list

# Forced on through a pipe (the one place escapes are expected):
/tmp/rune --color=always --list | cat | grep -c $'\x1b['   # > 0
```

## 4. Manual: `--color` validation

```sh
/tmp/rune --color=sometimes --list ; echo "exit=$?"
```
Expect: a clear error on stderr, non-zero exit, and **no** task listing/run.

## 5. Manual: per-stream independence (mixed redirection)

```sh
# stdout piped (plain list), stderr to terminal (status may color under auto):
/tmp/rune --list > list.txt        # list.txt has no ANSI
```
Expect: stdout-bound `--list` is plain because stdout is not a TTY, regardless of
stderr's TTY status.

## 6. Manual: redesigned `--help`

```sh
/tmp/rune --help            # grouped sections + colored headings on a TTY
/tmp/rune --help | cat      # same structure, no ANSI, readable
```
Expect: Usage / Common commands / Flags / Examples sections, plain-language flag
descriptions, and a worked example for run-a-task, `--list`, `--choose`, `serve`.
The piped form matches the new help golden baseline.

## 7. Diagnostics alignment (SC-003)

```sh
printf 'task t {\n  @echo {{nope}}\n}\n' > /tmp/Runefile
/tmp/rune -f /tmp/Runefile t            # colored caret on a TTY
/tmp/rune -f /tmp/Runefile t 2>&1 | cat # plain; caret under the same columns
```
Expect: caret span sits under the offending text identically in both modes; only
its color differs.

## Acceptance traceability

| Spec | Verified by |
|------|-------------|
| FR-003, FR-010, SC-001 | §2 goldens (piped / `NO_COLOR` / `--color=never`) |
| FR-004 (per-stream) | §5 |
| FR-006–FR-009 (flag) | §3, §4 |
| FR-013, SC-002 (`--list` align) | §2, §3 |
| FR-014–FR-016 (status/echo) | §2 + integration |
| FR-017, FR-018, SC-003 (diag) | §2, §7 |
| FR-019–FR-021, SC-006 (help) | §6 |
| FR-022 (forced-color test) | §3 (`--color=always` ANSI present) |
