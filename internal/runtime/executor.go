// Package runtime selects and drives the executor for a task body: the default
// pure-Go shell (internal/runtime/shell), a temp-file interpreter for
// python/node/custom or a `set shell` override (internal/runtime/interp), or the
// AI-agent driver (internal/runtime/agent).
package runtime

import (
	"strings"

	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/config"
)

// Kind identifies which executor runs a task body.
type Kind int

const (
	KindShell  Kind = iota // mvdan/sh (default, never the system shell)
	KindInterp             // temp-file + exec a real interpreter
	KindAgent              // drive an installed agent CLI (US4)
)

// Selection is the resolved executor for a task.
type Selection struct {
	Kind    Kind
	Command []string // interpreter command (KindInterp): e.g. ["python3"]
	Display string   // human-facing name for diagnostics
}

// Select resolves a task's executor from its declared executor, any
// [script("...")] attribute, and the module settings. A `set shell := [...]`
// override switches even default bodies to temp-file exec of the named shell.
func Select(task *ast.Task, settings config.Settings) Selection {
	if attr := task.Attr(ast.AttrScript); attr != nil && attr.Str != "" {
		return Selection{Kind: KindInterp, Command: splitCommand(attr.Str), Display: attr.Str}
	}
	switch task.Executor {
	case "", ast.ExecSh:
		if len(settings.Shell) > 0 {
			return Selection{Kind: KindInterp, Command: settings.Shell, Display: "shell:" + settings.Shell[0]}
		}
		return Selection{Kind: KindShell, Display: "sh"}
	case ast.ExecPython:
		return Selection{Kind: KindInterp, Command: defaultCmd(settings.Python, "python3"), Display: "python"}
	case ast.ExecNode:
		return Selection{Kind: KindInterp, Command: defaultCmd(settings.Node, "node"), Display: "node"}
	case ast.ExecAgent:
		return Selection{Kind: KindAgent, Display: "agent"}
	default:
		// A custom executor name is treated as the interpreter command itself.
		return Selection{Kind: KindInterp, Command: []string{task.Executor}, Display: task.Executor}
	}
}

func defaultCmd(list []string, def string) []string {
	if len(list) > 0 {
		return list
	}
	return []string{def}
}

func splitCommand(s string) []string {
	return strings.Fields(s)
}
