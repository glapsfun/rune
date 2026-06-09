// Package interp runs python/node/custom (and `set shell`-override) task bodies
// by writing the interpolated body to a temp file and exec-ing the configured
// interpreter against it (the just shebang/[script] model). No language runtimes
// are embedded (Principle V). A missing interpreter is an actionable error.
package interp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/rune-task-runner/rune/internal/token"
)

// Options configures an interpreter run.
type Options struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Dir    string
	Env    []string
}

// MissingInterpreterError reports that the configured interpreter is not on PATH.
type MissingInterpreterError struct {
	Task string
	Name string
	Err  error
}

func (e *MissingInterpreterError) Error() string {
	return fmt.Sprintf("task %q: interpreter %q not found on PATH — install it or set its command (e.g. `set python := [...]`)", e.Task, e.Name)
}

// ExecError reports a non-zero interpreter exit.
type ExecError struct {
	Task string
	Name string
	Code int
	Span token.Span
	Err  error
}

func (e *ExecError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("task %q (%s) failed: %v", e.Task, e.Name, e.Err)
	}
	return fmt.Sprintf("task %q (%s) failed: exit status %d", e.Task, e.Name, e.Code)
}

func (e *ExecError) Unwrap() error { return e.Err }

// Run writes script to a temp file and execs command against it, streaming I/O.
func Run(ctx context.Context, taskName, script string, command []string, span token.Span, opts Options) error {
	if len(command) == 0 {
		return &ExecError{Task: taskName, Err: errors.New("no interpreter configured")}
	}
	bin := command[0]
	resolved, err := exec.LookPath(bin)
	if err != nil {
		return &MissingInterpreterError{Task: taskName, Name: bin, Err: err}
	}

	f, err := os.CreateTemp("", "rune-script-*")
	if err != nil {
		return &ExecError{Task: taskName, Name: bin, Err: err}
	}
	tmp := f.Name()
	defer func() { _ = os.Remove(tmp) }()
	if _, err := f.WriteString(script); err != nil {
		_ = f.Close()
		return &ExecError{Task: taskName, Name: bin, Err: err}
	}
	if err := f.Close(); err != nil {
		return &ExecError{Task: taskName, Name: bin, Err: err}
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(tmp, 0o700)
	}

	args := append(append([]string{}, command[1:]...), tmp)
	cmd := exec.CommandContext(ctx, resolved, args...)
	cmd.Dir = opts.Dir
	cmd.Env = opts.Env
	cmd.Stdin = opts.Stdin
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return &ExecError{Task: taskName, Name: bin, Code: ee.ExitCode(), Span: span}
		}
		return &ExecError{Task: taskName, Name: bin, Err: err, Span: span}
	}
	return nil
}
