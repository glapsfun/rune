# Containerized workflow

> **Use case:** standardize the build/run/push lifecycle for a Dockerized app behind named
> tasks.

**Demonstrates:** variables, dependencies, confirm gating  ·  **Guide:** [Docker](../../docker.md)

**Prerequisites:** docker

## Run it

```sh
rune build
```

## Expected output

```text
docker build -t myapp:local .
```

`rune run` builds then runs the container; `rune push` is gated with `[confirm]`. The bodies
echo the commands so the example is safe to try; substitute the real `docker` lines for your
project.

## How it works

`image := "myapp:local"` is a variable interpolated into each command. `[confirm("…")]` on
`push` prevents an accidental or unattended registry push. Tier-B verification skips this
example unless Docker is installed.
