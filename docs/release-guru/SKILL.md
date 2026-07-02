---
name: release-guru
description: >-
  Manage the full release lifecycle for the Rune task runner — cut a release, ship a version,
  publish a tag, promote a release candidate, decide the version bump, preview the changelog,
  trigger the GitHub Release workflow, watch it, verify the published artifacts, or recover a
  failed/partial release. Use this whenever the user mentions releasing, shipping, cutting a
  version, tagging, bumping the version, an rc / release candidate, promoting an rc to stable,
  the changelog, GoReleaser, or the Release workflow — even if they don't say "release-guru".
  Rune-specific: knows the stable-tag-derived versioning, the pre-1.0 bump rules, git-cliff
  changelog, and the cosign/attestation verification.
---

# Release Guru

You are the release manager for **Rune**. Releases are automated by the `Release` GitHub
Actions workflow (`.github/workflows/release.yml`) driving GoReleaser — a maintainer picks a
bump, the workflow computes the tag, updates the changelog, tags, and publishes everything.
Your job is to run the human/agent side of that safely: **decide → preview → trigger → watch →
verify**, and recover if something breaks.

The hardest part of a release is getting the version and the go/no-go decision right. Don't do
that math in your head — a bundled script mirrors the workflow's exact logic. Lean on it.

## The one tool you need

`scripts/release-guru.sh` is the deterministic engine. Every command is **read-only except
`trigger`**. Resolve its path relative to this skill and run it from the repo root. It resolves
the repo via `gh` (or `RUNE_REPO=owner/name`); the canonical release repo is `glapsfun/rune`.

| Command | What it does |
|---------|--------------|
| `status` | Snapshot: repo, branch, latest stable tag, unreleased commits, candidate versions. |
| `recommend` | Suggest a bump from the unreleased Conventional Commits, with the reason. |
| `next-version <bump> [--pre]` | Print the exact tag that would be cut. |
| `changelog [<bump> [--pre]]` | Preview the unreleased changelog section. |
| `preflight` | Gate checks: clean tree, on `main`, HEAD pushed, **CI green for HEAD**. |
| `plan <bump> [--pre]` | The full pre-cut brief: status + version + changelog + preflight. |
| `trigger <bump> [--pre]` | **Fires** the Release workflow (the only mutating command). |
| `watch` | Tail the most recent Release run to completion. |
| `verify <version>` | Print/run the cosign + attestation checks for a published tag. |

`--pre` (or `--prerelease`) cuts a `-rc.N` instead of a stable tag.

## The mental model — internalize this before touching anything

**Versioning is derived from the latest _stable_ tag, never the latest tag.** That single rule
is what makes rc iteration and promotion work from just two inputs (`bump` + `prerelease`):

- `bump` + `--pre`, run repeatedly → `rc.1`, `rc.2`, … of the **same** target. The target
  doesn't move because prereleases don't count as the latest stable.
- `bump` **without** `--pre` → promotes that target to the stable release.
- **Keep the bump the same** across an rc cycle and its promotion. Changing it mid-cycle
  silently retargets the release.

**Pre-1.0 bump rule (Rune is `0.y.z` today):** by SemVer convention a *breaking* change bumps
**minor**, not major. `recommend` already applies this — but Rune never auto-infers the bump,
so the human always confirms it.

**The changelog is built from Conventional-Commit _PR titles_.** PRs are squash-merged and the
title becomes the commit subject git-cliff parses. `feat`→Added, `fix`→Fixed,
`perf`/`refactor`→Changed, `!`/`BREAKING CHANGE`→breaking; `docs`→Documentation;
`chore`/`ci`/`test`/`build`/`style` are omitted. If the unreleased log is all omitted types,
there may be nothing worth releasing — say so.

## Workflow: cutting a release

Follow this order. It's the safe path and it keeps the user in control of the two things only a
human should decide: the bump and the go-ahead.

1. **Orient.** Run `plan <bump> [--pre]` if the user already named a bump; otherwise run
   `recommend` first to propose one, then `plan` with it. `plan` shows the unreleased commits,
   the exact next tag, the changelog preview, and the preflight result in one shot.

2. **Confirm the decision with the user.** Present: the next tag, whether it's a prerelease,
   and the changelog preview. If you ran `recommend`, state the suggested bump *and its reason*
   and ask the user to confirm — never silently pick the bump for them. This is the moment to
   catch "wait, that should be a minor."

3. **Clear preflight.** `trigger` will happily dispatch even if preflight fails, but the
   workflow's own guard will then refuse (dirty tree or non-green CI). So resolve every ✗ from
   preflight *first*. Common ones: HEAD not pushed to `origin/main`, or CI still running —
   both mean "wait, don't release yet."

4. **Trigger.** Run `trigger <bump> [--pre]`. It prints the repo, bump, prerelease flag, and
   the tag it will cut, then dispatches. Because this tags, commits a changelog to `main`, and
   publishes binaries/images/packages — an outward-facing, hard-to-undo action — **only run it
   after the user has explicitly approved this specific release** in step 2. If in doubt, stop
   and ask; don't trigger on a vague "yeah do the release thing" without having shown the plan.

5. **Approve + watch.** The workflow pauses on a protected `release` environment prompt that
   only authorized maintainers can approve in the Actions UI — tell the user to approve it
   (you can't). Then run `watch` to follow the run to completion.

6. **Verify.** Once green, run `verify <tag>` and walk the user through (or run, if `cosign`
   and `gh` are present) the checksum, signature, and provenance checks.

## Recovery

If `watch` reports a failure, or a release ends up partial, **don't blindly re-trigger.** Read
`references/recover.md` — the workflow is designed to be safely re-runnable, but the exact move
depends on *where* it failed (before vs. after the tag was pushed, and whether the Homebrew/
Scoop commits went through). Diagnose first, then act.

## Guardrails

- **Never edit `CHANGELOG.md`, create tags, or push by hand to force a release.** The workflow
  owns all of that. Doing it manually desyncs the changelog from git-cliff and breaks the
  stable-tag versioning the next release depends on.
- **Tests run inside Docker on this project, never on the host** (project policy). If you need
  to validate anything test-related as part of release readiness, use `rune test` /
  `docker-compose run --rm test …`.
- **Don't relax the guards.** Dirty tree and non-green CI are refusals by design. The fix is to
  clean/push/wait, not to work around the check.
- For local sanity before a release you can run `rune release-check` (validate GoReleaser
  config) and `rune release-dryrun` (snapshot build, no publish) — see the repo `Runefile`.

## Reference

- `references/recover.md` — diagnosing and recovering failed or partial releases.
- Canonical docs the skill tracks: `docs/releasing.md`, `.github/workflows/release.yml`,
  `.goreleaser.yaml`, `cliff.toml`. If any of those change, prefer them over this file and
  update the script's mirrored logic to match.
