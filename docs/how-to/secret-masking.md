# Secret masking

> How Rune keeps credentials out of terminal transcripts and agent chat
> histories. Part of the [guides](README.md); settings syntax in the
> [language guide](../runefile.md#secret-masking).

## Concept

If a task prints an environment variable that holds a credential — an `env`
dump, a debug line, a tool echoing its config on failure — Rune replaces the
value with `***` in **everything it emits**: task stdout and stderr, echoed
command lines, Rune's own status and error lines, and MCP tool results sent to
agents. The task itself always receives the real value; only Rune's output is
transformed. Masking is always on — there is no flag or environment variable
that disables it.

Secrets are identified by **name**, the way CI systems mask registered
secrets. Any variable in the task's environment whose name contains one of
these patterns (case-insensitive) is masked automatically:

```text
TOKEN · SECRET · PASSWORD · PASSWD · APIKEY · API_KEY ·
PRIVATE_KEY · ACCESS_KEY · CREDENTIAL · AUTH
```

A few ubiquitous, definitively non-secret names that the `AUTH` pattern would
otherwise catch are exempt by default: `SSH_AUTH_SOCK`, `GIT_AUTHOR_NAME`,
`GIT_AUTHOR_EMAIL`, and `GIT_AUTHOR_DATE` (declaring one in `set secrets`
re-includes it).

## Syntax

Two settings give the Runefile the last word:

```rune
# Mask variables the built-in patterns miss:
set secrets := ["DEPLOY_CFG", "UPLOAD_URL"]

# Exempt a false positive (e.g. OAUTH_METHOD=oauth2 is not a credential):
set unmasked := ["OAUTH_METHOD"]
```

Names in both lists match environment variables case-insensitively, just like
the built-in patterns. Listing the same name in both is a static error
(reported with both source positions, nothing runs). A listed name that isn't
present in the environment is simply inert.

## Runnable example

See **[examples/secret-masking](../examples/secret-masking/README.md)** — a
pattern-matched token masked automatically, plus an `unmasked` exemption.

## What is (and isn't) guaranteed

- **Verbatim occurrences** of a secret's value are masked everywhere,
  including values split across output buffers and each line of a multi-line
  value (PEM keys). An interrupted task can never have emitted an unmasked
  window — masking happens at emission time, not afterwards.
- **Transformed values are not masked**: if a task base64- or URL-encodes a
  secret, or splits it itself, the transformed text passes through. Masking
  is a safety net, not encryption — don't print secrets on purpose.
- **Values shorter than 4 bytes are not masked** — replacing a 1-character
  value would mangle unrelated output.
- **Detection is by name, not content.** A credential in a variable with an
  innocent name is masked only if you declare it in `set secrets`.

## Pitfalls

- **Over-matching `AUTH`.** Names like `OAUTH_METHOD` match the `AUTH`
  pattern; exempt them with `set unmasked` rather than renaming.
- **Same value, two names.** If a non-secret variable happens to hold the
  same value as a secret one, occurrences are masked regardless — the mask
  set is value-based once names are resolved.
- **Secrets still belong in the environment.** Masking complements — never
  replaces — the rule that secrets live in `.env` or the environment, not in
  a committed Runefile (see [Settings & dotenv](settings-and-dotenv.md)).
- **Task output flows through Rune when masking is active.** With at least one
  secret in the environment, child processes write to a pipe instead of
  inheriting the terminal directly: tools that auto-detect a TTY fall back to
  their non-interactive output, and a background process a task leaves running
  can keep the run open until it releases the stream. Secret-free environments
  are unaffected.

## Next steps

- [Settings & dotenv](settings-and-dotenv.md) — where configuration and
  secrets come from.
- [AI agents (MCP)](../mcp.md) — the agent surface this protects.
