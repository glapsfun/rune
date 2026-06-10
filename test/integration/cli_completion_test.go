package integration

import (
	"strings"
	"testing"
)

// US2: dynamic task-name completion (with descriptions) + per-shell scripts.
// `rune __complete <args>` is Cobra's hidden completion driver: it prints one
// candidate per line followed by a ":<directive>" line.

func TestUS2Complete_TaskNamesWithDescriptions(t *testing.T) {
	dir := writeRunefile(t, "# Build it.\nbuild:\n    @echo b\n\n# Test it.\ntest:\n    @echo t\n")
	r := run(t, dir, nil, "__complete", "")
	if r.code != 0 {
		t.Fatalf("`rune __complete \"\"` exit = %d, want 0; stderr=%s", r.code, r.stderr)
	}
	for _, want := range []string{"build", "test", "serve", "version"} {
		if !strings.Contains(r.stdout, want) {
			t.Errorf("completion should suggest %q; got:\n%s", want, r.stdout)
		}
	}
	if !strings.Contains(r.stdout, "Build it.") {
		t.Errorf("completion should include task descriptions; got:\n%s", r.stdout)
	}
	// ShellCompDirectiveNoFileComp == 4; Cobra prints it as a trailing ":4".
	if !strings.Contains(r.stdout, ":4") {
		t.Errorf("expected NoFileComp directive (:4); got:\n%s", r.stdout)
	}
}

func TestUS2Complete_GracefulWithoutRunefile(t *testing.T) {
	dir := t.TempDir() // no Runefile
	r := run(t, dir, nil, "__complete", "")
	if r.code != 0 {
		t.Fatalf("`rune __complete` without a Runefile exit = %d, want 0; stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stdout, "serve") {
		t.Errorf("built-in commands should still complete without a Runefile; got:\n%s", r.stdout)
	}
}

func TestUS2Completion_ScriptGenerates(t *testing.T) {
	dir := writeRunefile(t, "build:\n    @echo b\n")
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		r := run(t, dir, nil, "completion", shell)
		if r.code != 0 {
			t.Errorf("`rune completion %s` exit = %d, want 0; stderr=%s", shell, r.code, r.stderr)
		}
		if strings.TrimSpace(r.stdout) == "" {
			t.Errorf("`rune completion %s` produced no script", shell)
		}
	}
}

// FR-013 (analyze finding G1): an unsupported shell must fail clearly.
func TestUS2Completion_UnsupportedShellErrors(t *testing.T) {
	dir := writeRunefile(t, "build:\n    @echo b\n")
	r := run(t, dir, nil, "completion", "tcsh")
	if r.code == 0 {
		t.Errorf("completion for unsupported shell should fail; got exit 0, stdout:\n%s", r.stdout)
	}
}
