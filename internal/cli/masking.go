package cli

import (
	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/mask"
)

// applyMasking derives the run's secret-value set from the effective
// environment plus every task's [env("K","V")] values (the set is engine-wide
// because all tasks share the same output streams) and, when any value is
// tracked, wraps opts.Stdout/Stderr in emission-time masking writers — the
// single choke point covering task passthrough, command echo, Rune's own
// status lines, the agent write-back, and the MCP adapter's buffers.
//
// The returned flush func emits any withheld stream tail. It must only be
// called once no producer can still be writing (i.e. after scheduler.Run has
// returned) — flushing earlier could emit a parallel task's in-flight secret
// prefix verbatim. When nothing is tracked the writers are left untouched, so
// secret-free runs stay byte-identical, and the flush func is a no-op.
func applyMasking(opts Options, env []string, tasks map[string]*ast.Task, declared, exempt []string) (Options, func()) {
	full := env[:len(env):len(env)]
	for _, t := range tasks {
		for _, a := range t.Attributes {
			if a.Kind == ast.AttrEnv && a.Str != "" {
				full = append(full, a.Str+"="+a.Str2)
			}
		}
	}
	set := mask.NewSet(full, declared, exempt)
	if set.Empty() {
		return opts, func() {}
	}
	stdout := mask.NewWriter(opts.Stdout, set)
	stderr := mask.NewWriter(opts.Stderr, set)
	opts.Stdout = stdout
	opts.Stderr = stderr
	return opts, func() {
		_ = stdout.Flush()
		_ = stderr.Flush()
	}
}
