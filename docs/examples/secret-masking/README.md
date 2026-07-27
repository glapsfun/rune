# Secret masking

> **Use case:** run tasks that touch credentials without ever leaking them — values of
> sensitive environment variables are masked as `***` in task output, echoed commands,
> and MCP tool results sent to agents.

**Demonstrates:** built-in name patterns, `set secrets`, `set unmasked`  ·  **Guide:** [Secret masking](../../how-to/secret-masking.md)

**Prerequisites:** none

## Run it

```sh
rune show-env
```

## Expected output

```text
token  is ***
config is ***
author is ada
```

`DEMO_API_TOKEN` matches the built-in `TOKEN` pattern, so it is masked
automatically. `DEPLOY_CFG` has an innocent name and is masked because the
Runefile declares it in `set secrets`. `AUTHOR` matches the `AUTH` pattern but
is exempted with `set unmasked`, so its value prints normally.

## How it works

Rune inspects the *names* of the task's environment variables, collects the
values of sensitive ones, and replaces every occurrence with `***` in
everything it emits — the task itself still receives the real values. Masking
is always on; the only opt-out is a per-variable `set unmasked` entry. See the
[secret masking guide](../../how-to/secret-masking.md) for the exact rules and
limits.
