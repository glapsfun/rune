package config

import (
	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/diag"
	"github.com/rune-task-runner/rune/internal/eval"
)

// Settings holds the resolved module settings that influence execution.
type Settings struct {
	WorkingDir string   // set working-directory := "path"
	Quiet      bool     // set quiet
	Export     bool     // set export (export all variables to the task env)
	Fallback   bool     // set fallback
	Dotenv     string   // set dotenv := ".env"
	Shell      []string // set shell := ["bash", "-cu"]
	Python     []string // set python := ["python3"]
	Node       []string // set node := ["node"]
	AgentCmd   []string // set agent_cmd := ["claude", "-p"]
}

// ResolveSettings evaluates a file's settings into a Settings value. String and
// list settings are evaluated with ev; boolean settings use their bare form.
func ResolveSettings(f *ast.File, ev *eval.Evaluator) (Settings, diag.List) {
	var s Settings
	var diags diag.List

	evalStr := func(set *ast.Setting) string {
		if set.Value == nil {
			return ""
		}
		v, err := ev.Eval(set.Value)
		if err != nil {
			diags.Add(diag.New(err.Span, err.Msg))
			return ""
		}
		return v
	}
	evalList := func(set *ast.Setting) []string {
		if set.List == nil {
			if set.Value != nil {
				return []string{evalStr(set)}
			}
			return nil
		}
		out := make([]string, 0, len(set.List))
		for _, e := range set.List {
			v, err := ev.Eval(e)
			if err != nil {
				diags.Add(diag.New(err.Span, err.Msg))
				continue
			}
			out = append(out, v)
		}
		return out
	}

	for _, set := range f.Settings {
		switch set.Name {
		case "working-directory":
			s.WorkingDir = evalStr(set)
		case "quiet":
			s.Quiet = true
		case "export":
			s.Export = true
		case "fallback":
			s.Fallback = true
		case "dotenv":
			s.Dotenv = evalStr(set)
		case "shell":
			s.Shell = evalList(set)
		case "python":
			s.Python = evalList(set)
		case "node":
			s.Node = evalList(set)
		case "agent_cmd", "agent_provider":
			s.AgentCmd = evalList(set)
		}
	}
	return s, diags
}
