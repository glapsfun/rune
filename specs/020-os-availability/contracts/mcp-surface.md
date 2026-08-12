# Contract: MCP Agent Surface

**Feature**: 020-os-availability

## Tool list (`mcpAdapter.Tasks()` → `mcpserver.New` tool registration)

| Aspect | Contract |
|--------|----------|
| Exposure rule | A task is exposed as an MCP tool iff `!IsPrivate() && AvailableOn(host)` |
| Applies to | Both transports: stdio (`rune mcp`) and HTTP (`rune serve`) |
| Authz | `mcpserver/authz.go` allowlist/destructive gating iterates the same filtered `Engine.Tasks()` — an OS-mismatched task can be neither listed nor allowlisted |
| Invisibility | No trace of mismatched tasks in tool names, descriptions, or schemas (composes with the existing no-secrets rule) |

## Tool call (`mcpAdapter.Call()`)

| Aspect | Contract |
|--------|----------|
| Defense-in-depth | A call naming an OS-mismatched task fails with the same "not available on HOST" message as the CLI root gate, before any scheduling |
| Rationale | The adapter's internal task map stays complete (mismatched tasks must remain resolvable as skippable deps), so the call path re-checks |
| Unknown vs unavailable | Unknown names keep the existing `unknown task: NAME` error; unavailable names get the availability error (more truthful diagnostic) |
| Dep behavior | A tool call whose task has OS-mismatched deps runs with those deps silently skipped, exactly like the CLI path (same scheduler) |

## Documentation contract

`docs/mcp.md` states the exposure rule as: every non-private task
**available on the host OS** is exposed as an MCP tool.
