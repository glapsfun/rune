# Dependencies & post-hooks

> How to make tasks run in the right order. Part of the [guides](README.md); full syntax in the
> [language guide](../runefile.md#dependencies-and-post-hooks).

## Concept

A task can declare **dependencies** (tasks that must run *before* it) and **post-hooks**
(tasks that run *after* it, only if it succeeds). Within one invocation, each task runs **at
most once**, even if several tasks depend on it — so a shared `build` step isn't repeated.

## Syntax

Dependencies go after the colon; post-hooks after `&&`:

```rune
deploy: build test && notify
    @echo "deploying"
```

Pass arguments to a dependency with the parenthesized form:

```rune
release: (build "release")
    @echo "releasing"

build target="debug":
    @echo "building {{target}}"
```

## Runnable example

See **[examples/dependencies](../examples/dependencies/README.md)** — `rune deploy` runs
`build` and `test` first, then `notify` after.

## Pitfalls

- **Cycles are rejected up front.** `a: b` and `b: a` fail static analysis with the cycle path
  (`a → b → a`) and exit code `3` — nothing runs. See [Troubleshooting](../troubleshooting.md).
- **Each task runs once per invocation.** If you actually need a step to run twice, model it as
  two tasks; dependency memoization is by design.
- **Post-hooks run only on success.** If the task fails, its `&&` hooks do not run.

## Next steps

- [Parameters](parameters.md) — pass values into tasks and dependencies.
- [Parallelism](parallelism.md) — run independent dependencies concurrently.
