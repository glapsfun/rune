# Quickstart — Validating the Idiomatic Go Refactor

Runnable scenarios proving each user story. Per global policy the Go suite runs **inside
Docker** (`docker-compose run --rm test …`), never on the host. Commands assume repo root.

Cross-refs: rubric → `contracts/review-rubric.md`; invariants → `contracts/preservation-
invariants.md`; lint set → `contracts/lint-gate.md`; finding model → `data-model.md`.

---

## Prerequisites

- Docker + standalone `docker-compose` (the test/lint harness).
- Network access for `go run` of `golangci-lint`, `benchstat` (and the `goleak` test dep).

---

## US1 — Skill-governed review (P1)

**Goal**: a complete, traceable findings report (SC-001).

```sh
# The report exists and follows the schema
test -f specs/003-idiomatic-go-refactor/review.md
# Every Go package appears (reviewed or clean):
for p in $(go list ./... | sed 's#.*/rune/##'); do grep -q "$p" specs/003-idiomatic-go-refactor/review.md || echo "MISSING coverage: $p"; done
# Spot-check a finding has all required fields (skill, rule, loc, sev, rec)
grep -E 'F-[0-9]+ \| skill=.* \| rule=.* \| loc=.*:[0-9]+ \| sev=S[1-4]' specs/003-idiomatic-go-refactor/review.md | head
```

**Expected**: report present; every package covered; findings carry skill+rule+location+
severity, ordered S1→S4.

---

## US2 — Correctness & safety remediation (P1)

**Goal**: all S1 findings fixed; race- and leak-clean; behavior preserved (SC-002/003/004).

```sh
# Full suite green, no behavior change, NO golden regenerated:
docker-compose run --rm test go test ./...
git status --porcelain testdata/        # → empty

# Race + leak clean (goleak TestMain in scheduler/mcpserver/cli):
docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...

# The specific S1 fixes are in place:
grep -n 'srv.Serve' mcpserver/transport.go        # error is captured/surfaced, not `_ =`
grep -n 'cache.Store' internal/cli/run.go          # failure logged, not silently `_ =`
grep -nc 'goleak' internal/runtime/scheduler/*_test.go mcpserver/*_test.go internal/cli/*_test.go
```

**Expected**: suite + `-race` green, 0 races, 0 leaks; no golden changed; S1 sites remediated.

---

## US3 — Idiomatic design & structure (P2)

**Goal**: `init()`/globals gone, naming fixed, layout intact, suite still green (SC-006).

```sh
# Zero init() in non-test code (or each justified):
grep -rn --include='*.go' '^func init()' cmd internal mcpserver | grep -v _test    # → empty

# The eval builtins global is no longer init()-populated:
grep -n 'sync.OnceValue\|func newBuiltins\|func init' internal/eval/builtins.go

# Naming fix: no global shadows a package/predeclared name:
grep -n 'var unsafe' internal/cache/cache.go    # → gone/renamed

# Locked package layout intact (names unchanged):
go list ./... | grep -E 'internal/(token|lexer|ast|parser|analyzer|eval|diag|config|dotenv|cache|cli|runtime)|mcpserver'

# Behavior still preserved:
docker-compose run --rm test go test ./...
```

**Expected**: zero unjustified `init()`/globals; locked packages unchanged; suite green.

---

## US4 — Encode skills as gates (P2)

**Goal**: expanded linter set passes clean and is live (SC-005).

```sh
# The new linters are configured:
grep -E 'errorlint|contextcheck|noctx|bodyclose|predeclared|wastedassign|revive|gocritic' .golangci.yml

# The refactored tree passes the expanded gate with zero issues:
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./...   # → "0 issues."

# (Proof it's live) a seeded violation is caught — e.g. add a dead assignment and re-run:
#   expect wastedassign/ineffassign to flag it; then revert.
```

**Expected**: linters present; `0 issues` on the clean tree; seeded violation flagged.

---

## US5 — Benchmark-gated performance (P3)

**Goal**: hot-path benchmarks exist; only benchstat-proven changes merge (SC-008).

```sh
# Benchmarks exist on the hot paths:
grep -rn --include='*_test.go' 'func Benchmark' internal/lexer internal/parser internal/eval internal/runtime/scheduler

# Establish a baseline (run in Docker):
docker-compose run --rm test go test -run=xxx -bench=. -benchmem ./internal/lexer ./internal/parser ./internal/eval ./internal/runtime/scheduler | tee bench-before.txt

# Any optimization PR must show a benchstat win:
#   go run golang.org/x/perf/cmd/benchstat@latest bench-before.txt bench-after.txt
```

**Expected**: benchmarks present for lexer/parser/eval/scheduler; a baseline is recordable;
any merged perf change carries a benchstat improvement + `perf(scope):` commit.

---

## Done-when (maps to Success Criteria)

- [ ] SC-001 all packages reviewed, findings well-formed · [ ] SC-002 all S1 fixed
- [ ] SC-003 `-race` + goleak clean · [ ] SC-004 suite green, no golden regenerated
- [ ] SC-005 expanded gate clean + live · [ ] SC-006 zero unjustified init()/globals
- [ ] SC-007 exported identifiers documented · [ ] SC-008 hot-path benchmarks + benchstat-gated
- [ ] SC-009 mcpserver/CLI contract tests pass unmodified · [ ] SC-010 every fix → finding id
