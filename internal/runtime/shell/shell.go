// Package shell runs (sh) task bodies through the pure-Go mvdan/sh interpreter.
// It NEVER shells out to the system shell (Principle V), so default-task
// behavior is identical across Linux, macOS, and Windows. Body lines run on a
// single persistent Runner so cd and exported variables carry across lines.
package shell

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"

	"github.com/rune-task-runner/rune/internal/token"
)

// Line is one interpolated body line plus its per-line sigils.
type Line struct {
	Text            string
	NoEcho          bool
	ContinueOnError bool
	Span            token.Span
}

// Options configures a shell run.
type Options struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Dir    string   // working directory ("" => current)
	Env    []string // KEY=VALUE pairs (nil => inherit os.Environ)
	Quiet  bool     // suppress command echo
}

// ExecError reports a body-line failure: the task, the offending line, and the
// underlying exit code (or shell error).
type ExecError struct {
	Task string
	Line string
	Span token.Span
	Code int
	Err  error
}

func (e *ExecError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("task %q failed at %q: %v", e.Task, e.Line, e.Err)
	}
	return fmt.Sprintf("task %q failed at %q: exit status %d", e.Task, e.Line, e.Code)
}

func (e *ExecError) Unwrap() error { return e.Err }

// Run executes the body lines sequentially on one interpreter, echoing each
// command to stderr unless suppressed, honoring @ (no echo) and - (continue on
// error) per line, and returning an *ExecError on the first uncontinued failure.
func Run(ctx context.Context, taskName string, lines []Line, opts Options) error {
	dir := opts.Dir
	if dir == "" {
		if wd, err := os.Getwd(); err == nil {
			dir = wd
		}
	}
	env := opts.Env
	if env == nil {
		env = os.Environ()
	}
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	runner, err := interp.New(
		interp.StdIO(opts.Stdin, stdout, stderr),
		interp.Dir(dir),
		interp.Env(expand.ListEnviron(env...)),
	)
	if err != nil {
		return &ExecError{Task: taskName, Err: err}
	}

	parser := syntax.NewParser()
	for _, ln := range lines {
		if strings.TrimSpace(ln.Text) == "" {
			continue
		}
		if !ln.NoEcho && !opts.Quiet {
			fmt.Fprintln(stderr, ln.Text)
		}
		prog, perr := parser.Parse(strings.NewReader(ln.Text), taskName)
		if perr != nil {
			if ln.ContinueOnError {
				continue
			}
			return &ExecError{Task: taskName, Line: ln.Text, Span: ln.Span, Err: perr}
		}
		rerr := runner.Run(ctx, prog)
		if rerr == nil {
			continue
		}
		if ln.ContinueOnError {
			continue
		}
		var status interp.ExitStatus
		if errors.As(rerr, &status) {
			return &ExecError{Task: taskName, Line: ln.Text, Span: ln.Span, Code: int(status)}
		}
		if errors.Is(rerr, context.Canceled) {
			return rerr
		}
		return &ExecError{Task: taskName, Line: ln.Text, Span: ln.Span, Err: rerr}
	}
	return nil
}
