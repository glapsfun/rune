package integration

import (
	"strings"
	"testing"
)

// US2: static validation gates execution. Each error exits 3, runs nothing, and
// prints a located message with a caret.

func TestUS2_UnknownTaskExits3(t *testing.T) {
	dir := writeRunefile(t, "a: b\n    @echo ran-a\n")
	r := run(t, dir, nil, "a")
	if r.code != 3 {
		t.Fatalf("exit = %d, want 3; stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, "unknown task: b") {
		t.Errorf("stderr missing 'unknown task: b': %q", r.stderr)
	}
	// Located message + caret, nothing executed.
	if !strings.Contains(r.stderr, "Runefile:1:") {
		t.Errorf("stderr missing file:line:col: %q", r.stderr)
	}
	if !strings.Contains(r.stderr, "^") {
		t.Errorf("stderr missing caret: %q", r.stderr)
	}
	if strings.Contains(r.stdout, "ran-a") {
		t.Errorf("task body executed despite validation failure: %q", r.stdout)
	}
}

func TestUS2_SelfCycleExits3(t *testing.T) {
	dir := writeRunefile(t, "c: c\n    @echo ran-c\n")
	r := run(t, dir, nil, "c")
	if r.code != 3 {
		t.Fatalf("exit = %d, want 3; stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, "dependency cycle: c → c") {
		t.Errorf("stderr missing cycle: %q", r.stderr)
	}
	if strings.Contains(r.stdout, "ran-c") {
		t.Errorf("task executed despite cycle: %q", r.stdout)
	}
}

func TestUS2_UndefinedVariableExits3(t *testing.T) {
	dir := writeRunefile(t, "greet:\n    @echo {{undefined_var}}\n")
	r := run(t, dir, nil, "greet")
	if r.code != 3 {
		t.Fatalf("exit = %d, want 3; stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, "undefined variable: undefined_var") {
		t.Errorf("stderr missing undefined var: %q", r.stderr)
	}
}

func TestUS2_AllDiagnosticsAndNothingRuns(t *testing.T) {
	src := "a: b\n    @echo ran-a\nc: c\n    @echo ran-c\ngreet:\n    @echo {{undefined_var}}\n"
	dir := writeRunefile(t, src)
	r := run(t, dir, nil, "a")
	if r.code != 3 {
		t.Fatalf("exit = %d, want 3", r.code)
	}
	for _, want := range []string{"unknown task: b", "dependency cycle", "undefined variable: undefined_var"} {
		if !strings.Contains(r.stderr, want) {
			t.Errorf("stderr missing %q: %s", want, r.stderr)
		}
	}
	if strings.Contains(r.stdout, "ran-") {
		t.Errorf("a task ran despite validation errors: %q", r.stdout)
	}
}

func TestUS2_ParseErrorExits3(t *testing.T) {
	// A malformed expression is a parse error, still routed through exit 3.
	dir := writeRunefile(t, "x := \n")
	r := run(t, dir, nil, "--list")
	if r.code != 3 {
		t.Errorf("exit = %d, want 3; stderr=%s", r.code, r.stderr)
	}
}
