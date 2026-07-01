# Caching

> How to skip an expensive task safely. Part of the [guides](README.md); full syntax in the
> [language guide](../runefile.md#caching-opt-in).

## Concept

Rune is a command runner, not a build system: **a task always runs when you ask** — it never
skips work based on file timestamps. Caching is an **explicit, per-task opt-in**. When you add
`[cache(...)]`, Rune computes a fingerprint over the declared **inputs**, the **task body**,
the **resolved variables**, and the **executor**. If nothing changed and the declared outputs
exist, the task is skipped — and the skip is **logged** ("cached"), never silent.

## Syntax

```rune
[cache(inputs = ["go.mod", "go.sum", "**/*.go"], outputs = ["dist/app"])]
build:
    go build -o dist/app ./...
```

Clear stored fingerprints with `rune --clear-cache`.

## Runnable example

See **[examples/caching](../examples/caching/README.md)** — run `rune build` twice; the second
run reports the task as cached.

## Pitfalls

- **Missing outputs force a run.** If a declared output is absent, the task runs even on a
  fingerprint match — you can't "cache" your way out of a deleted artifact.
- **Unresolvable inputs force a run** rather than a false skip.
- **Caching is never the default.** If you didn't add `[cache]`, the task runs every time —
  that's the point (no `.PHONY` traps, no stale "up-to-date" guesses).
- **The body is part of the fingerprint.** Editing the task's commands invalidates the cache,
  as it should.

## Next steps

- [Dependencies & post-hooks](dependencies-and-hooks.md) — cached steps still respect order.
- [CLI reference](../cli.md) — `--clear-cache`.
