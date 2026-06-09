# Content-hash caching (opt-in)

> **Use case:** skip an expensive build step when its inputs haven't changed — explicitly,
> never by guessing from timestamps.

**Demonstrates:** content-hash caching  ·  **Guide:** [Caching](../../runefile.md#caching-opt-in)

**Prerequisites:** none

## Run it

```sh
rune build
```

## Expected output

```text
building (cached when VERSION is unchanged)
```

Run `rune build` again **without** changing `VERSION` and Rune reports the task as **cached**
and skips it. Edit `VERSION` (or run `rune --clear-cache`) and it builds again.

## How it works

`[cache(inputs = ["VERSION"], outputs = ["dist/app"])]` fingerprints the declared inputs, the
task body, the resolved variables, and the executor. A hit is logged, never silent — caching
is opt-in, consistent with Rune's "always run unless you asked to cache" model.
