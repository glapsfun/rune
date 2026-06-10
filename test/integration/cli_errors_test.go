package integration

import (
	"strings"
	"testing"
)

// US3: friendly errors (did-you-mean), validated serve flags, and the `--`
// escape hatch for tasks that collide with a built-in command name.

func TestUS3_DidYouMeanCommand(t *testing.T) {
	dir := writeRunefile(t, "build:\n    @echo b\n")
	r := run(t, dir, nil, "serv") // typo of the `serve` command
	if r.code != 2 {
		t.Fatalf("`rune serv` exit = %d, want 2; stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, "did you mean") || !strings.Contains(r.stderr, "serve") {
		t.Errorf(`expected a 'did you mean "serve"?' suggestion; got: %q`, r.stderr)
	}
	// Concise: no full usage/help dump on error.
	if strings.Contains(r.stderr, "Usage:") || strings.Contains(r.stderr, "Available Commands") {
		t.Errorf("error should be concise (no usage dump); got: %q", r.stderr)
	}
}

func TestUS3_DidYouMeanTask(t *testing.T) {
	dir := writeRunefile(t, "build:\n    @echo b\n")
	r := run(t, dir, nil, "biuld") // typo of the `build` task
	if r.code != 2 {
		t.Fatalf("`rune biuld` exit = %d, want 2; stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, "did you mean") || !strings.Contains(r.stderr, "build") {
		t.Errorf(`expected a suggestion of 'build'; got: %q`, r.stderr)
	}
}

// addr/token-file without --http is a usage error (exit 2). Bounded with stdin
// so that, before validation exists, a stdio server would exit on EOF rather
// than hang the test.
func TestUS3_ServeAddrRequiresHTTP(t *testing.T) {
	dir := writeRunefile(t, "build:\n    @echo b\n")
	r := runWithStdin(t, dir, "", "serve", "--addr", ":7777")
	if r.code != 2 {
		t.Errorf("`rune serve --addr :7777` without --http exit = %d, want 2; stderr=%s", r.code, r.stderr)
	}
}

// FR-008: a task whose name collides with a built-in is reachable via `--`.
func TestUS3_DashEscapesToTask(t *testing.T) {
	dir := writeRunefile(t, "serve:\n    @echo ran-the-task\n")
	r := run(t, dir, nil, "--", "serve")
	if r.code != 0 {
		t.Fatalf("`rune -- serve` exit = %d, want 0; stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stdout, "ran-the-task") {
		t.Errorf("`rune -- serve` should run the task named serve; stdout=%q stderr=%q", r.stdout, r.stderr)
	}
}
