// Package cli wires the command-line surface: Runefile resolution, the parse →
// analyze → run pipeline, task dispatch, and exit-code mapping.
package cli

import "errors"

// Exit codes (contracts/cli.md, FR-021).
const (
	ExitSuccess    = 0   // all requested tasks succeeded
	ExitTaskFail   = 1   // a task body failed
	ExitUsage      = 2   // usage error / no Runefile / unknown task / bad args
	ExitValidation = 3   // static parse/analyze error — nothing executed
	ExitInterrupt  = 130 // interrupted (SIGINT)
)

// UsageError marks an error as a usage/discovery problem (exit 2).
type UsageError struct{ Err error }

func (e *UsageError) Error() string { return e.Err.Error() }
func (e *UsageError) Unwrap() error { return e.Err }

// ValidationError marks a static parse/analyze failure (exit 3). The diagnostics
// have already been rendered to stderr by the caller; this type only carries the
// exit code intent.
type ValidationError struct{ Err error }

func (e *ValidationError) Error() string { return e.Err.Error() }
func (e *ValidationError) Unwrap() error { return e.Err }

// TaskFailure marks a task body failure (exit 1). Silent suppresses the trailing
// error banner (the [no-exit-message] attribute) without changing the exit code.
type TaskFailure struct {
	Err    error
	Silent bool
}

func (e *TaskFailure) Error() string { return e.Err.Error() }
func (e *TaskFailure) Unwrap() error { return e.Err }

// Interrupted marks a SIGINT (exit 130).
type Interrupted struct{ Err error }

func (e *Interrupted) Error() string {
	if e.Err == nil {
		return "interrupted"
	}
	return e.Err.Error()
}
func (e *Interrupted) Unwrap() error { return e.Err }

// CodeFor maps an error returned by Run to a process exit code.
func CodeFor(err error) int {
	if err == nil {
		return ExitSuccess
	}
	var ue *UsageError
	if errors.As(err, &ue) {
		return ExitUsage
	}
	var ve *ValidationError
	if errors.As(err, &ve) {
		return ExitValidation
	}
	var ti *Interrupted
	if errors.As(err, &ti) {
		return ExitInterrupt
	}
	var tf *TaskFailure
	if errors.As(err, &tf) {
		return ExitTaskFail
	}
	// Default: treat unknown errors as usage problems.
	return ExitUsage
}

// usagef builds a UsageError with a formatted message.
func usagef(format string, args ...any) error {
	return &UsageError{Err: errorf(format, args...)}
}
