# Contract: Runefile Grammar (EBNF)

**Feature**: 001-rune-task-runner | **Date**: 2026-06-08

Non-normative reference grammar for the `Runefile` language; the hand-written parser
(`internal/parser`) is the source of truth. This seeds `docs/GRAMMAR.md`. Bodies use
**significant indentation** (Clarification Q4 / FR-002). `NEWLINE`, `INDENT`, `DEDENT` are
produced by the lexer.

```ebnf
File        = { Item } ;
Item        = Comment | Setting | Assignment | Import | Mod | Task ;

Comment     = "#" { any-char } NEWLINE ;          (* run of comments above a Task = its doc *)

Setting     = "set" Name [ ":=" Value ] NEWLINE ; (* bare form ≡ ":= true" *)
Value       = Expr | List ;
List        = "[" [ Expr { "," Expr } ] "]" ;

Assignment  = Name ":=" Expr NEWLINE ;

Import      = "import" [ "?" ] StringLit NEWLINE ;
Mod         = "mod" Name [ StringLit ] NEWLINE ;

Task        = { Attribute } Signature ":" [ Deps ] [ "&&" PostHooks ]
              NEWLINE INDENT Body DEDENT ;
Signature   = Name { Param } [ "(" Executor ")" ] ;
Param       = Name [ "=" Expr ]            (* defaulted *)
            | "+" Name                     (* variadic, one-or-more *)
            | "*" Name ;                   (* variadic, zero-or-more *)
Executor    = "sh" | "python" | "node" | "agent" | Name ;
Deps        = DepCall { DepCall } ;
PostHooks   = DepCall { DepCall } ;
DepCall     = Name | "(" Name { Expr } ")" ;     (* parenthesized form passes args *)

Body        = { BodyLine } ;
BodyLine    = [ "@" ] [ "-" ] Text NEWLINE ;     (* @=no-echo, -=continue-on-error *)
Text        = { TextChar | Interpolation } ;
Interpolation = "{{" Expr "}}" ;                 (* "{{{{" escapes a literal brace *)

Attribute   = "[" AttrItem { "," AttrItem } "]" NEWLINE ;
AttrItem    = "private"
            | "confirm" [ "(" StringLit ")" ]
            | "group" "(" StringLit ")"
            | "parallel"
            | "linux" | "macos" | "windows" | "unix"
            | "no-cd"
            | "network"                  (* sets MCP openWorldHint *)
            | "no-exit-message"          (* suppress the trailing error banner *)
            | "working-directory" "(" StringLit ")"
            | "env" "(" StringLit "," StringLit ")"
            | "doc" "(" StringLit ")"
            | "script" "(" StringLit ")"
            | "cache" "(" "inputs" "=" List [ "," "outputs" "=" List ] ")" ;

(* ---- Expression sublanguage (Pratt-parsed; total, non-Turing-complete) ---- *)
Expr        = Conditional | Or ;
Conditional = "if" Expr CmpOp Expr "{" Expr "}"
              { "else" "if" Expr CmpOp Expr "{" Expr "}" }
              "else" "{" Expr "}" ;
Or          = Concat ;
Concat      = Unary { ("+" | "/") Unary } ;       (* "+"=concat, "/"=path-join *)
Unary       = Primary ;
Primary     = StringLit | FuncCall | VarRef | "(" Expr ")" ;
FuncCall    = Name "(" [ Expr { "," Expr } ] ")" ;
VarRef      = Name ;
CmpOp       = "==" | "!=" | "=~" ;                (* =~ = regex match *)

StringLit   = "'" { ch } "'" | "\"" { ch } "\""
            | "'''" { ch } "'''" | "\"\"\"" { ch } "\"\"\"" ;  (* triple = de-dented *)
Name        = letter { letter | digit | "_" | "-" } ;
```

## Notes & invariants

- **No loops/recursion** in `Expr` (Principle III). Conditionals and function calls are the only
  control/branching constructs.
- Path-join `/` always emits forward slashes, even on Windows (Principle V).
- A comment run immediately above a Task with no blank line is its doc comment (FR-008).
- Indentation must be consistent within a Task body; mixing tabs and spaces within one body is a
  lexer-level located error (FR-002).
- Settings may each appear at most once (analyzer, FR-010).
- This grammar is guarded by the compatibility-corpus test (Principle VI): changes that alter how
  existing fixtures parse must be deliberate and versioned (FR-033).
