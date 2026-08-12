# Quickstart: Validating Accurate Version Reporting

Runnable scenarios proving the feature end-to-end. Contract IDs (C1–C8)
refer to [contracts/version-output.md](contracts/version-output.md).

**Prerequisites**: Docker running (Rancher Desktop) for the test suite; Go
toolchain on the host for the build-mode checks; repo root as cwd.

## V1 — Unit + regression suites (C1–C8 logic)

```sh
docker-compose run --rm test go test ./...
docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...
```

**Expected**: all green, including the new `cmd/rune` build-info unit tests
and the untouched `test/integration/minimum_version_test.go` (SC-004).

## V2 — Go-toolchain install reports the release version (C4, C5; FR-001)

**Note**: `go install …@v0.4.3` fetches the *v0.4.3 source*, which predates
this fix and will always report `dev` — it cannot validate the change.
Pre-release, install the fix itself at a pushed commit (module-cache build,
same code path as `@latest`); the expected value is then a pseudo-version
(C5), not a clean release version.

Pre-release (fix commit pushed to GitHub):

```sh
bin=$(mktemp -d)
sha=$(git rev-parse HEAD)   # a commit containing the fix, already pushed
GOBIN=$bin go install github.com/rune-task-runner/rune/cmd/rune@"$sha"
"$bin"/rune --version
```

**Expected**: `rune version 0.4.4-0.<timestamp>-<sha12> (commit none)` —
a Go pseudo-version, not `dev` (C5; visibly not a clean release).

Post-release (first release containing the fix, the real-world C4 check):

```sh
bin=$(mktemp -d)
GOBIN=$bin go install github.com/rune-task-runner/rune/cmd/rune@latest
"$bin"/rune --version
```

**Expected**: `rune version <X.Y.Z> (commit none)` for the newest tag —
not `dev` (FR-001, SC-001).

## V3 — Checkout builds stay `dev` (C6; FR-003)

```sh
go run ./cmd/rune --version
go build -o /tmp/rune-dev ./cmd/rune && /tmp/rune-dev --version
```

**Expected**: both print `rune version dev (commit none)`, even at a
tagged, clean commit.

## V4 — Release artifact output byte-identical (C3; FR-002, SC-003)

```sh
rune --version          # installed 0.4.3 Homebrew binary
go build -trimpath -ldflags "-s -w -X main.version=9.9.9 -X main.commit=abc1234" -o /tmp/rune-rel ./cmd/rune
/tmp/rune-rel --version
```

**Expected**: `rune version 0.4.3 (commit 13d059a)` and
`rune version 9.9.9 (commit abc1234)` — ldflags always wins; format
unchanged.

## V5 — Gate engages for real versions, bypasses dev (C7, C8; FR-004)

Setup — a Runefile that requires 9.0.0, and two stamped binaries (reuse
`/tmp/rune-rel` = 9.9.9 from V4):

```sh
d=$(mktemp -d); printf 'set minimum_version := "9.0.0"\nhi:\n    @echo HI\n' > "$d/Runefile"
go build -ldflags "-X main.version=0.4.3 -X main.commit=abc1234" -o /tmp/rune-old ./cmd/rune
```

(a) Satisfying version passes; dev build bypasses (C8):

```sh
(cd "$d" && /tmp/rune-rel hi; echo "exit=$?")   # 9.9.9 ≥ 9.0.0 → prints HI, exit 0
repo=$PWD; (cd "$d" && go run "$repo"/cmd/rune hi; echo "exit=$?")  # dev → gate bypassed, prints HI
```

(b) Older real version is blocked with the standard diagnostic (C7):

```sh
(cd "$d" && /tmp/rune-old hi; echo "exit=$?")
(cd "$d" && /tmp/rune-old version --check; /tmp/rune-old version --check --json)
```

**Expected (b)**: mismatch diagnostic with `installed version: 0.4.3`,
`required version:  9.0.0`, `nothing was executed`, exit 3; both `--check`
forms report the same 0.4.3 and the unsatisfied requirement (FR-006).

## V6 — Full quality gates

```sh
rune lint && rune docs-check && rune release-dryrun
```

**Expected**: golangci-lint clean (Principle VIII), docs harness green
(S4's `docs/installation.md` edit validated), GoReleaser snapshot succeeds
(release pipeline untouched).
