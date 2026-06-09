# Polyglot repository

> **Use case:** a repo whose tasks span languages — shell, Python, and Node — all driven from
> one Runefile.

**Demonstrates:** executors (sh / python / node)  ·  **Guide:** [Executors](../../runefile.md#executors)

**Prerequisites:** python3, node

## Run it

```sh
rune all
```

## Expected output

```text
shell step
python step
node step
polyglot pipeline complete
```

## How it works

`all` depends on three tasks, each declaring a different executor: the default shell,
`(python)`, and `(node)`. Rune runs each body under the right interpreter (which must be
installed). Tier-B verification skips this example unless both Python and Node are present.
