# Quickstart: Validating Colorful CLI Output Everywhere

End-to-end validation guide for feature 014. Contracts referenced:
[styled-output.md](contracts/styled-output.md); surface inventory:
[data-model.md](../014-colorful-cli-output/data-model.md).

## Prerequisites

```sh
go build -o /tmp/rune ./cmd/rune          # or: rune build (dogfood)
cd $(mktemp -d)
cat > Runefile <<'EOF'
# Build the thing
build:
    echo building

# Fail on purpose
[group("dev")]
boom:
    exit 1
EOF
```

Full test suite (authoritative; runs the invariance matrix from D10):

```sh
docker-compose run --rm test go test ./...
```

## Scenario 1 — Failure banner is styled (US1 / C1)

```sh
/tmp/rune --color=always boom 2>&1 | cat -v | grep 'rune:'
```

**Expect**: escape sequences around the `rune:` prefix (e.g. `^[[1;31m`),
message text plain, exit code 1 preserved (`echo $?` after a non-piped run).

```sh
/tmp/rune --color=never boom 2>err.txt; cat -v err.txt
```

**Expect**: `rune: task "boom" failed …` with zero `^[` sequences —
byte-identical to the pre-014 banner.

## Scenario 2 — Warnings are uniform (US1 / C2)

Add `set minimum_version := "99.0.0"` to the Runefile, then:

```sh
/tmp/rune --color=always --ignore-version build 2>&1 | cat -v | grep warning
```

**Expect**: `warning` label carries the same yellow styling as the cache-write
warning; with `NO_COLOR=1` the line is exactly
`warning: ignoring Runefile minimum Rune version 99.0.0; running <version>`.

## Scenario 3 — Status notices join the theme (US2 / C3)

```sh
/tmp/rune --color=always --fmt 2>&1 | cat -v          # formatted: styled label
/tmp/rune --color=always --clear-cache 2>&1 | cat -v  # cleared: styled label
/tmp/rune --color=always --watch build                # watching…/change detected… muted; Ctrl-C to stop
/tmp/rune --color=always                              # overview: styled header + styled list
```

**Expect**: labels/chrome styled per C3's table; paths and values plain; with
`--color=never` every line matches its pre-014 bytes.

## Scenario 4 — Analyze diagnostic parity (US2 / C4)

Break the Runefile (e.g. reference an undefined variable), then:

```sh
/tmp/rune analyze --color=always | cat -v     # caret + severity color, like a run
/tmp/rune analyze --json | grep -c $'\x1b'    # expect 0 (and JSON unchanged)
/tmp/rune build 2>&1 | cat -v                 # same diagnostic rendering as analyze
```

**Expect**: `analyze` text output visually identical to the run-path
diagnostic (severity word colored, faint locator, caret span) while keeping
analyze's coded severity token (`error[RUNE2001]`); exit 3 with error
diagnostics; JSON byte-identical to pre-014. Both `--color` positions work —
the flag is persistent (audit finding A13).

## Scenario 5 — Uniform help (US3 / C5)

```sh
for c in "" serve version analyze lsp completion; do
  /tmp/rune $c --help | head -5
  /tmp/rune $c --help | grep -c $'\x1b'   # 0 when piped
done
/tmp/rune serve --help > /dev/tty         # headings colored on a TTY
```

**Expect**: every help screen shows the grouped layout (description, Usage,
Examples, Flags); piped help has zero escapes; on a TTY headings use the same
accent as root help.

## Scenario 6 — Machine surfaces stay clean (C7)

```sh
/tmp/rune --color=always --dump | grep -c $'\x1b'            # 0
/tmp/rune --color=always --dump --format json | grep -c $'\x1b'  # 0
/tmp/rune completion zsh | grep -c $'\x1b'                   # 0
/tmp/rune version --json | grep -c $'\x1b'                   # 0
```

Also verify an MCP tool call's result carries no ANSI (integration test
`internal/cli` MCP suite covers this; manual check via `rune serve --mcp` and
an MCP client is optional).

## Scenario 7 — Masking through styled banners (C1, FR-011)

```sh
MY_TOKEN=hunter2token /tmp/rune --color=always boom-with-secret 2>&1 | cat -v
```

(where the task's failure message embeds `$MY_TOKEN`)

**Expect**: `***` in the styled banner; the literal secret never appears; ANSI
present around `rune:`.

## Scenario 8 — Alignment invariance (C6)

Add a task with a doc comment and a long name; compare column of `#` between
`--color=always` and `--color=never` output of `--list` (visually or via
ANSI-strip): **identical**. (Implementation note: the grammar restricts task
names to ASCII, so the feared wide-rune misalignment is unreachable; parity is
pinned by `TestListTasksColoredAlignmentMatchesPlain` instead of a CJK fixture,
and no display-width dependency was added.)

## Sign-off checklist (completed 2026-07-23)

- [x] All 8 scenarios pass locally (S2 verified with the `-tags runetest`
      binary + `RUNE_TEST_VERSION`, since a dev build satisfies any minimum)
- [x] `docker-compose run --rm test go test ./...` green
- [x] `docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...` green
- [x] Golden diffs: **none at all** — analyze/help reformats are covered by
      integration tests, no golden files exist for those surfaces (SC-002)
- [x] `golangci-lint run` zero issues; `gofmt -l` clean; `go vet` clean;
      `go mod tidy` no-op; docs tests green
- [x] SC-001 checklist in [data-model.md](../014-colorful-cli-output/data-model.md)
      surfaces table: every S1–S13 verified styled (S13 = alignment parity
      pinned by test); audit A1–A13 dispositions confirmed in cli-audit.md
