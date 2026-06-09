# CI/CD pipeline

> **Use case:** run the same lint/test/build gate locally and in CI, with a gated deploy and
> predictable exit codes.

**Demonstrates:** dependencies, dry-run, confirm gating, exit codes  ·  **Guide:** [CLI reference](../../cli.md)

**Prerequisites:** none

## Run it

```sh
rune ci
```

## Expected output

```text
golangci-lint run
go test ./...
go build ./...
CI passed ✓
```

In CI, `rune ci` exits non-zero if any step fails, so the pipeline fails fast. Preview without
running via `rune --dry-run ci`, and list tasks with `rune --list`. `deploy` is gated with
`[confirm]`, so an unattended run needs `--yes`.

## How it works

`ci: lint test build` chains the gate; `[confirm("…")]` on `deploy` blocks accidental or
unattended production deploys. Exit codes: `0` success, `1` task failure, `2` usage, `3`
validation.
