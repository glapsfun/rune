# Parallelism

> How to run independent prerequisites at the same time. Part of the [guides](README.md); full
> syntax in the [language guide](../runefile.md#attributes).

## Concept

By default a task's dependencies run **sequentially**. Adding the `[parallel]` attribute lets
its **independent** dependencies run **concurrently**, bounded by the available cores. The task
itself still runs only after all its dependencies succeed.

## Syntax

```rune
[parallel]
checks: lint test typecheck
    @echo "all checks passed"

lint:
    @echo "lint"
test:
    @echo "test"
typecheck:
    @echo "typecheck"
```

`lint`, `test`, and `typecheck` run together; `checks` runs once they all finish.

## Runnable example

See **[examples/parallel](../examples/parallel/README.md)**.

## Pitfalls

- **Order between parallel tasks is not guaranteed.** Don't rely on `lint` finishing before
  `test` — if there's a real ordering need, make one depend on the other.
- **Interleaved output.** Concurrent tasks' output can interleave; use task-level logging if
  you need to attribute lines.
- **Shared mutable state.** Parallel tasks that write the same file will race — keep them
  independent, or serialize them with a dependency.

## Next steps

- [Dependencies & post-hooks](dependencies-and-hooks.md) — the ordering model parallelism builds on.
- [Caching](caching.md) — combine with `[cache]` to skip unchanged parallel steps.
