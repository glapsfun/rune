# Recovering a failed or partial release

The Release workflow is built to be safely re-runnable, but **where** it failed decides the
right move. Diagnose before acting — a blind re-trigger can double-tag or leave the changelog
ahead of the artifacts.

## Step 1: find out how far it got

```sh
gh run view <run-id> -R glapsfun/rune            # which step failed
git fetch --tags origin
git tag --list 'v<target>*'                      # did the tag get created + pushed?
gh release view v<target> -R glapsfun/rune       # is there a (draft?) GitHub release?
```

The workflow's ordered checkpoints, each safe to cross twice:

1. **Guards** (clean tree / green CI) — nothing published yet. Fix the cause and re-trigger.
2. **Compute version** — refuses if the tag already exists. Idempotent.
3. **Changelog commit to `main`** + **tag push** — the first durable side effects.
4. **GoReleaser** — binaries, GHCR image, sigs, SBOMs, provenance.
5. **Homebrew cask / Scoop manifest** commits — the one spot that may need manual cleanup.

## Step 2: pick the recovery by failure point

**Failed at guards or version compute (before the tag):**
Nothing was published. Fix the underlying issue (push HEAD, wait for green CI, resolve a
dirty tree) and re-run the workflow with the *same* inputs. `--clean` wipes `dist/` first, so
there's no stale-artifact risk.

**Failed after the tag was pushed but during/after GoReleaser:**
GitHub Releases are upserted and GHCR pushes are content-addressed, so **re-running is safe** —
re-trigger the workflow (it will find the tag exists and… refuse at compute). Because compute
refuses an existing tag, the clean recovery is one of:

- **Re-run GoReleaser against the existing tag** locally or by re-dispatching only the publish
  step, *or*
- **Delete the tag and the (draft) release, then re-trigger** from scratch:
  ```sh
  gh release delete v<target> -R glapsfun/rune --yes    # remove the release first
  git push origin :refs/tags/v<target>                  # delete the remote tag
  git tag -d v<target>                                   # and the local tag
  ```
  Then re-run the workflow. Only do the delete path if publishing clearly did not complete —
  never delete a tag whose artifacts are already public and may be depended on.

**Homebrew tap / Scoop bucket left half-updated:**
Check `glapsfun/homebrew-tap` and `glapsfun/scoop-bucket` before re-running — a mid-way failure
can leave a bad or duplicate commit there. Revert/clean it manually, then re-run; `skip_upload:
auto` keeps prereleases off the stable channel, so an rc won't touch these.

## Step 3: confirm the recovery

After any recovery, run `release-guru.sh verify <tag>` and confirm the checksum, signature, and
provenance all pass. A release isn't recovered until verification is green.

## When in doubt

Stop and surface the state to the user (which tags exist, whether the release is drafted,
whether the tap/bucket commits landed) rather than guessing. A wrong tag or a changelog that's
ahead of the published artifacts is more painful to unwind than a paused release.
