# Quickstart: Validating Grouped Sections in the Interactive Task Picker

Validates the behaviors defined in `contracts/tui-picker-grouping.md` against
a running build. Requires an interactive terminal (the picker refuses to open
otherwise, per the existing `--choose` contract).

## Prerequisites

- A built `rune` binary: `go run ./cmd/rune` or `go build -o rune ./cmd/rune`.
- A scratch Runefile exercising groups, e.g.:

  ```text
  [group("build")]
  [doc("compile the binary")]
  build:
      @go build ./...

  [group("build")]
  [doc("run the linter")]
  lint:
      @golangci-lint run

  [group("test")]
  [doc("run the test suite")]
  test:
      @go test ./...

  [group("release")]
  [doc("cut a release")]
  release:
      @echo release

  [doc("say hello (ungrouped)")]
  hello:
      @echo hi
  ```

## Scenario 1 — Sections match `--list` (US1, FR-001–FR-003, SC-001, SC-003)

```sh
rune --list
```

Note the section order and membership (`[build]` → build, lint; `[test]` →
test; `[release]` → release; ungrouped → hello).

```sh
rune --choose
```

Expected: the picker shows the same sections, in the same order, with the
same membership — no task missing, none duplicated, none appearing outside
its `--list` section.

## Scenario 2 — No groups is a no-op (FR-004, SC-002)

Using a Runefile with no `group(...)` attributes at all:

```sh
rune --choose
```

Expected: a single flat list, no section headers — indistinguishable from the
picker's behavior before this feature.

## Scenario 3 — Navigation skips headers (US2, FR-005)

With the grouped Runefile from Prerequisites:

```sh
rune --choose
```

Press `↓` repeatedly from the first task to the last. Expected: the highlight
only ever rests on a task name (`build`, `lint`, `test`, `release`, `hello`),
never on a `[build]`/`[test]`/`[release]` header line.

## Scenario 4 — Filtering hides empty sections (US2, FR-006)

```sh
rune --choose
```

Type `rel` (matches only `release`'s description/name, or adjust to a query
that matches one section). Expected: only the `[release]` header and the
`release` task remain visible; the `[build]` and `[test]` headers disappear
along with their now-unmatched tasks.

## Scenario 5 — Selection runs unaffected by section (FR-008)

```sh
rune --choose
```

Select `lint` (inside the `[build]` section) and press `Enter`. Expected:
`golangci-lint run` executes exactly as `rune lint` would — section
membership has no effect on execution or exit code.

## Automated coverage

The above is also asserted by:
- `internal/tui` unit/model tests (cursor never rests on a header; filter
  hides empty sections) — see `research.md` R5.
- `internal/cli` golden/integration tests comparing `--choose`'s section
  derivation against `--list`'s existing grouping output for the same
  Runefile fixtures, plus a byte-identical check for the no-groups case.
