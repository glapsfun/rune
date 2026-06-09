# Parallel prerequisites

> **Use case:** speed up a pipeline by running independent prerequisites at the same time.

**Demonstrates:** parallelism  ·  **Guide:** [Attributes](../../runefile.md#attributes)

**Prerequisites:** none

## Run it

```sh
rune checks
```

## Expected output

```text
lint
test
typecheck
all checks passed
```

`lint`, `test`, and `typecheck` run concurrently (order between them may vary); `checks` runs
once they all succeed. Concurrency is bounded by the available cores.

## How it works

The `[parallel]` attribute on `checks` makes its independent dependencies run concurrently.
