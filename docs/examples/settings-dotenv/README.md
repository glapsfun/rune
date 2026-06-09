# Settings & dotenv

> **Use case:** configure a project with settings and a `.env` file, and reach those values
> from task bodies — keeping secrets in the environment, not the Runefile.

**Demonstrates:** settings, dotenv, exported variables  ·  **Guide:** [Variables and settings](../../runefile.md#variables-and-settings)

**Prerequisites:** none

## Run it

```sh
rune show
```

## Expected output

```text
app=demo greeting=hello
```

`app` is a Runefile variable made available to the body by `set export`; `greeting` comes from
the `.env` file in this directory (`GREETING=hello`).

## How it works

`set export` exports Runefile variables into task environments; a project `.env` is loaded
automatically. Put secrets in the environment (or `.env`, kept out of version control) — never
as literals in the Runefile.
