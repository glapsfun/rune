# OS filtering

> How to scope tasks to an operating system. Part of the [guides](README.md); full syntax in
> the [language guide](../runefile.md#attributes).

## Concept

An OS attribute restricts a task to one platform: it is **hidden from `--list` and dispatch**
on other operating systems. This keeps platform-specific setup in one Runefile without
cluttering every machine's task list. The `os()` and `arch()` built-ins report the current
platform for inline branching.

## Syntax

```rune
[linux]
setup-linux:
    apt-get install -y build-essential

[macos]
setup-macos:
    brew install coreutils

[windows]
setup-windows:
    choco install make

# Available everywhere; reports the platform.
info:
    @echo "on {{os()}}/{{arch()}}"
```

`[unix]` covers Linux and macOS together.

## Runnable example

See **[examples/os-filtering](../examples/os-filtering/README.md)** — `rune --list` shows only
the setup task for your current OS.

## Pitfalls

- **Requesting an off-OS task reports why it's unavailable** rather than silently doing
  nothing.
- **Prefer the cross-platform `sh` default** for everything that *can* be portable; reserve OS
  filtering for genuinely platform-specific steps (package managers, paths).
- **Don't hard-code path separators.** Use the `/` path-join operator, which emits forward
  slashes on every OS.

## Next steps

- [Executors](executors.md) — cross-platform shell vs. platform tools.
- [CLI reference](../cli.md) — how filtered tasks appear in `--list`.
