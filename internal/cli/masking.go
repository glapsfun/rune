package cli

import (
	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/mask"
)

// envAttrPairs returns a task's [env("K","V")] attributes as KEY=value pairs —
// the single definition shared by the real task environment (engine.taskEnv)
// and the mask-set derivation, so the two can never diverge.
func envAttrPairs(task *ast.Task) []string {
	var pairs []string
	for _, a := range task.Attributes {
		if a.Kind == ast.AttrEnv && a.Str != "" {
			pairs = append(pairs, a.Str+"="+a.Str2)
		}
	}
	return pairs
}

// deriveMaskSet builds the run's secret-value set from the effective
// environment plus every task's [env] values (the set is engine-wide because
// all tasks share the same output streams). It is derived once per run — and
// once per adapter on the MCP path, where it is reused across tool calls.
func deriveMaskSet(env []string, tasks map[string]*ast.Task, declared, exempt []string) *mask.Set {
	full := env[:len(env):len(env)]
	for _, t := range tasks {
		full = append(full, envAttrPairs(t)...)
	}
	return mask.NewSet(full, declared, exempt)
}

// maskOptions wraps opts.Stdout/Stderr in emission-time masking writers — the
// single choke point covering task passthrough, command echo, Rune's own
// status lines, the agent write-back, and the MCP adapter's buffers.
//
// The returned flush func emits any withheld stream tail. It must only be
// called once no producer can still be writing (i.e. after scheduler.Run has
// returned) — flushing earlier could emit a parallel task's in-flight secret
// prefix verbatim. For an empty set the writers are left untouched, so
// secret-free runs stay byte-identical, and the flush func is a no-op.
func maskOptions(opts Options, set *mask.Set) (Options, func()) {
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

// maskErr wraps err so its rendered message is masked while the error chain
// stays intact for errors.Is/As. This covers the one surface the wrapped
// writers cannot reach: the final "rune: ..." banner (cmd/rune and the watch
// loop) prints err.Error() to the raw process stderr, and an executor error
// embeds the interpolated failing command line.
func maskErr(err error, set *mask.Set) error {
	if err == nil || set.Empty() {
		return err
	}
	return &maskedError{err: err, set: set}
}

type maskedError struct {
	err error
	set *mask.Set
}

func (m *maskedError) Error() string { return m.set.MaskString(m.err.Error()) }
func (m *maskedError) Unwrap() error { return m.err }
