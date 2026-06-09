# Dependencies & post-hooks

> **Use case:** order work correctly — some tasks must run before, and some after, a task.

**Demonstrates:** dependencies, post-hooks  ·  **Guide:** [Dependencies](../../runefile.md#dependencies-and-post-hooks)

**Prerequisites:** none

## Run it

```sh
rune deploy
```

## Expected output

```text
building
testing
deploying
notify: deployed ✓
```

`deploy: build test && notify` — `build` and `test` are dependencies (run first); `notify` is
a post-hook (runs after `deploy` succeeds). Each task runs at most once per invocation.

## How it works

See the `Runefile`: prerequisites go after the colon, post-hooks after `&&`.
