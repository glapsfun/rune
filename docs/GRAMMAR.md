# Runefile Grammar

> **Non-normative.** The hand-written parser (`internal/parser`) is the source of
> truth. This document is generated/maintained from
> [`specs/001-rune-task-runner/contracts/grammar.md`](../specs/001-rune-task-runner/contracts/grammar.md).
>
> **Sync discipline (Constitution Principle VI):** any PR that changes DSL
> surface MUST update this file *and* the golden/integration fixtures
> (`testdata/lexer`, `testdata/parser`, `testdata/fmt`, `testdata/corpus`) in the
> same PR. The compatibility-corpus test (`test/corpus`) guards against silent
> grammar drift.

Bodies use **significant indentation**. `NEWLINE`, `INDENT`, `DEDENT` are
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
Expr        = Conditional | Concat ;
Conditional = "if" Expr CmpOp Expr "{" Expr "}"
              { "else" "if" Expr CmpOp Expr "{" Expr "}" }
              "else" "{" Expr "}" ;
Concat      = Primary { ("+" | "/") Primary } ;   (* "+"=concat, "/"=path-join *)
Primary     = StringLit | FuncCall | VarRef | "(" Expr ")" ;
FuncCall    = Name "(" [ Expr { "," Expr } ] ")" ;
VarRef      = Name ;
CmpOp       = "==" | "!=" | "=~" ;                (* =~ = regex match *)

StringLit   = "'" { ch } "'" | "\"" { ch } "\""
            | "'''" { ch } "'''" | "\"\"\"" { ch } "\"\"\"" ;  (* triple = de-dented *)
Name        = letter { letter | digit | "_" | "-" } ;
```

## Backward compatibility

Breaking changes to Runefile semantics are opt-in per file via a version pragma:

```rune
set rune_version := "1"
```

The default interpretation never changes under a user without such a pragma
(Constitution Governance / FR-033).
