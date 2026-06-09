# Phase 1 Data Model: Rune Task Runner

**Date**: 2026-06-08 | **Feature**: 001-rune-task-runner

This is the internal domain model — the AST produced by the parser (`internal/ast`) plus the
runtime structures used by the analyzer, evaluator, scheduler, cache, and MCP server. It maps
spec **Key Entities** and **Functional Requirements** to concrete types. Fields are described
language-agnostically; every node carries a `Span` for diagnostics (Principle II).

## Common

- **Position**: `{ offset int, line int, col int }` — byte offset + 1-based line/column.
- **Span**: `{ file string, start Position, end Position }` — attached to every AST node and
  every Diagnostic.

## Parse-time AST

### File (Module)  *(spec: Task file, Module)*

The root of one parsed `.rune` file. A project may be a tree of Files via `import`/`mod`.

| Field | Type | Notes |
|-------|------|-------|
| `path` | string | absolute path of this file |
| `settings` | `[]Setting` | at most one per setting name (validated) |
| `assignments` | `[]Assignment` | variables; order-insensitive within the module |
| `tasks` | `[]Task` | declaration order preserved (for `--list` & fmt) |
| `imports` | `[]Import` | spliced files |
| `mods` | `[]Mod` | namespaced submodules |
| `span` | Span | |

Relationships: a File **owns** its settings/assignments/tasks; **references** other Files via
imports (spliced into the same namespace) and mods (a child namespace).

### Setting  *(spec: Setting; FR-010)*

| Field | Type | Notes |
|-------|------|-------|
| `name` | enum | `shell`, `dotenv`, `default`, `export`, `working-directory`, `quiet`, `fallback`, `python`, `node`, `agent_provider`, … |
| `value` | `Expr` \| bool | boolean form (`set export`) ≡ `true` |
| `span` | Span | |

Validation: each name appears **at most once** per module.

### Assignment (Variable)  *(spec: Variable; FR-006)*

| Field | Type | Notes |
|-------|------|-------|
| `name` | string | unique within module |
| `expr` | `Expr` | statically evaluated, no parse-time side effects |
| `span` | Span | |

Overridable at run time via `name=value` / `--set name value` (CLI layer injects an override
map consulted before file assignments).

### Task (Recipe)  *(spec: Task; FR-002, FR-003, FR-004, FR-008, FR-009)*

| Field | Type | Notes |
|-------|------|-------|
| `name` | string | unique within its module/namespace |
| `doc` | string | from preceding comment run or `[doc("…")]` |
| `params` | `[]Param` | positional; ≤1 trailing variadic |
| `executor` | Executor | `sh`(default) \| `python` \| `node` \| `agent` \| custom |
| `deps` | `[]DepCall` | prior dependencies |
| `postHooks` | `[]DepCall` | run after, on success (`&&`) |
| `attributes` | `[]Attribute` | private/confirm/group/parallel/os/working-directory/env/cache/script |
| `body` | `[]BodyLine` | significant-indentation block |
| `span` | Span | |

State (runtime, per `(task, resolved-args)` key): `Pending → Scheduled → (Running | SkippedCached) → (Succeeded | Failed)`. Memoized: a key executes **at most once** per invocation (FR-005).

### Param  *(spec: Parameter; FR-003)*

| Field | Type | Notes |
|-------|------|-------|
| `name` | string | |
| `kind` | enum | `required` \| `defaulted` \| `variadicPlus` (`+p`, ≥1) \| `variadicStar` (`*p`, ≥0) |
| `default` | `Expr?` | only for `defaulted` |
| `span` | Span | |

Validation: at most one variadic, and it MUST be the last param; defaulted params MUST follow
required ones.

### DepCall (Dependency / Post-hook)  *(spec: Dependency / Post-hook; FR-004)*

| Field | Type | Notes |
|-------|------|-------|
| `taskName` | string | may be namespaced (`docker::push`) |
| `args` | `[]Expr` | arguments passed to the dependency |
| `span` | Span | |

Validation: `taskName` MUST resolve to a known task; `len(args)` MUST satisfy the target's arity.

### Attribute  *(spec: Attribute; FR-009)*

| Field | Type | Notes |
|-------|------|-------|
| `kind` | enum | `private`, `confirm`, `group`, `parallel`, `linux`/`macos`/`windows`/`unix`, `no-cd`, `working-directory`, `env`, `doc`, `script`, `cache` |
| `args` | `[]AttrArg` | e.g. `confirm("prompt")`, `group("name")`, `cache(inputs=[…], outputs=[…])` |
| `span` | Span | |

`cache` carries a **CacheSpec**: `{ inputs []Expr (globs), outputs []Expr }`.

### BodyLine  *(spec: body interpolation & sigils; FR-018)*

| Field | Type | Notes |
|-------|------|-------|
| `raw` | string | text with `{{ … }}` interpolation placeholders |
| `noEcho` | bool | leading `@` |
| `continueOnError` | bool | leading `-` |
| `span` | Span | |

### Expr (interface)  *(spec: expression sublanguage; FR-007)* — total, non-Turing-complete

Variants: `StringLit` (single/double/triple, with de-dent for triple), `Concat(+)`,
`PathJoin(/)`, `Conditional(if/else if/else)`, `Comparison(==, !=)`, `RegexMatch(=~)`,
`FuncCall(name, args)`, `VarRef(name)`, `ParamRef(name)`. **No** loops, recursion, or
user-defined functions.

Built-in functions (registry, FR-007): `env`, `os`, `arch`, `os_family`, `num_cpus`, path ops
(`join`,`clean`,`extension`,`file_name`,`file_stem`,`parent_dir`,`absolute_path`), string ops
(`uppercase`,`lowercase`,`trim`,`replace`,`replace_regex`, case-conversions), `path_exists`,
`read`, hashing (`sha256`,`sha256_file`), `uuid`, `datetime`, `error`, `require`/`which`,
`quote`.

### Import / Mod  *(spec: Module / Import; FR-011)*

- **Import**: `{ path Expr, optional bool (import?), span }` — definitions spliced into the
  current namespace; order-insensitive; name collisions are a reported conflict.
- **Mod**: `{ name string, path Expr?, span }` — loads another File as a child namespace
  addressable as `name::task` / `name task`; child inherits parent env, has its own settings.

## Runtime model

### ResolvedTask

`{ task *Task, args map[string]Value, namespace string }` — the memoization key is
`(namespace, task.name, canonical(args))`.

### ExecutionPlan / Node

A DAG built from `deps`/`postHooks`. Node: `{ resolved ResolvedTask, dependsOn []NodeID,
postOf NodeID? }`. The scheduler topo-sorts, detects cycles (reported pre-execution), and runs
`[parallel]` siblings concurrently via `errgroup` bounded by `num_cpus()`.

### Value

Evaluator output: strings (the only first-class value type) plus the boolean used by `set`
flags and conditionals. Lists exist only as setting/attribute arguments (e.g. `set shell`,
`cache(inputs=[…])`), not as general expression values — keeping the language total.

### CacheFingerprint  *(spec: Cache fingerprint; FR-015, FR-020)*

`{ key string (namespace::task), hash string (sha256 hex), inputs []string, outputs []string,
executor string, createdAt string }` — stored as JSON under `.rune/cache/`. See
`contracts/cache-fingerprint.md`.

### ExposedTool  *(spec: Exposed tool; FR-025, FR-028)*

Derived from a non-private Task: `{ name (namespaced with __), description (doc), inputSchema
(from params: defaulted→optional, variadic→array), annotations { destructiveHint (=has
[confirm]/destructive), openWorldHint (network) } }`.

### Diagnostic  *(spec: located errors; FR-013)*

`{ severity (error|warning), span Span, message string, snippet string (source line + caret
underline) }`. Rendered by `internal/diag`.

## Cross-cutting validation rules (analyzer — run before any execution, FR-012/FR-014)

1. Every `VarRef`/`ParamRef` resolves to a known variable/param (else undefined-name error).
2. Every `DepCall.taskName` (and CLI-invoked task) resolves to a known task (unknown-task error).
3. Dependency graph is acyclic; on a cycle, report the full cycle path.
4. Arity: every CLI invocation and `DepCall` satisfies the target task's param arity.
5. Each setting name appears at most once; variadic param is last; defaulted params follow
   required ones.
6. `import` name collisions are reported, not silently resolved.
7. Body indentation is consistent within a task; tab/space mixing within one body is an error.
8. On any error: emit all diagnostics, execute nothing, exit non-zero.
