# Installation

> Get the `rune` binary onto your machine. Already installed? Jump to
> [Getting started](getting-started.md).

Rune ships as a single, statically-linked binary with no runtime dependencies. Pick the
method that fits your platform.

## Prebuilt binary (recommended)

Download the archive for your OS/architecture from the
[GitHub Releases](https://github.com/rune-task-runner/rune/releases) page, extract it, and
put the `rune` binary on your `PATH`.

**Linux / macOS:**

```sh
# Replace VERSION, OS (linux|darwin), ARCH (amd64|arm64) as appropriate.
curl -sSfL -o rune.tar.gz \
  https://github.com/rune-task-runner/rune/releases/download/vVERSION/rune_VERSION_OS_ARCH.tar.gz
tar -xzf rune.tar.gz rune
sudo mv rune /usr/local/bin/
rune --version
```

**Windows (PowerShell):** download the `..._windows_amd64.zip` archive, extract `rune.exe`,
and add its folder to your `PATH`.

Every release also publishes a `checksums.txt`; verify your download against it.

## From source (Go)

Requires Go 1.24 or newer (the project builds on Go 1.25):

```sh
go install github.com/rune-task-runner/rune/cmd/rune@latest
```

This installs `rune` into `$(go env GOPATH)/bin`. Ensure that directory is on your `PATH`.

To build from a checkout:

```sh
git clone https://github.com/rune-task-runner/rune
cd rune
go build -o rune ./cmd/rune
```

Rune is pure Go with `CGO` disabled, so cross-compilation is trivial:

```sh
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o rune ./cmd/rune
```

## Container

Run Rune with no local toolchain using the official image. See the
[Docker guide](docker.md) for full details (mounting your project, passing arguments, and
the minimal-image limitations):

```sh
docker run --rm -v "$PWD":/work ghcr.io/rune-task-runner/rune --list
```

## Shell completions

Rune can generate completion scripts:

```sh
rune completion bash   # or: zsh | fish | powershell
```

Install the output per your shell's convention (for example, source it from your shell rc
file, or place it in your completions directory).

## Verify

```sh
rune --version
rune --help
```

Next: the [getting-started guide](getting-started.md).
