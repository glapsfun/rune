package integration

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

const cliSecret = "hunter2-cli-secret"

// T013a/c — terminal env value masked on stdout, every occurrence, across lines.
func TestSecretMasking_TerminalStdoutMasked(t *testing.T) {
	src := "leaky:\n    @echo \"tok is $CLI_DEMO_TOKEN and again $CLI_DEMO_TOKEN\"\n    @echo \"second line $CLI_DEMO_TOKEN\"\n"
	dir := writeRunefile(t, src)
	r := run(t, dir, []string{"CLI_DEMO_TOKEN=" + cliSecret}, "leaky")
	if r.code != 0 {
		t.Fatalf("exit = %d, want 0; stderr=%s", r.code, r.stderr)
	}
	if strings.Contains(r.stdout, cliSecret) || strings.Contains(r.stderr, cliSecret) {
		t.Fatalf("raw secret leaked: stdout=%q stderr=%q", r.stdout, r.stderr)
	}
	if got, want := strings.Count(r.stdout, "***"), 3; got != want {
		t.Errorf("masked %d occurrences, want %d: %q", got, want, r.stdout)
	}
}

// T013b — a failing task's stderr is masked and the failure exit is preserved.
func TestSecretMasking_FailingTaskStderrMasked(t *testing.T) {
	src := "boom:\n    @echo \"failing with $CLI_DEMO_TOKEN\" >&2\n    @exit 7\n"
	dir := writeRunefile(t, src)
	r := run(t, dir, []string{"CLI_DEMO_TOKEN=" + cliSecret}, "boom")
	if r.code == 0 {
		t.Fatalf("exit = 0, want failure")
	}
	if strings.Contains(r.stderr, cliSecret) {
		t.Fatalf("raw secret leaked on stderr: %q", r.stderr)
	}
	if !strings.Contains(r.stderr, "failing with ***") {
		t.Errorf("stderr = %q, want masked failure line", r.stderr)
	}
}

// T014 — an un-suppressed command line interpolating a secret echoes masked;
// `@` and `set quiet` suppression semantics are unchanged.
func TestSecretMasking_EchoedCommandMasked(t *testing.T) {
	src := "deploy:\n    echo bearer {{env(\"ECHO_DEMO_TOKEN\")}}\n"
	dir := writeRunefile(t, src)
	r := run(t, dir, []string{"ECHO_DEMO_TOKEN=" + cliSecret}, "deploy")
	if r.code != 0 {
		t.Fatalf("exit = %d; stderr=%s", r.code, r.stderr)
	}
	if strings.Contains(r.stderr, cliSecret) || strings.Contains(r.stdout, cliSecret) {
		t.Fatalf("raw secret leaked: stdout=%q stderr=%q", r.stdout, r.stderr)
	}
	if !strings.Contains(r.stderr, "echo bearer ***") {
		t.Errorf("echoed command not masked on stderr: %q", r.stderr)
	}
	if !strings.Contains(r.stdout, "bearer ***") {
		t.Errorf("command output not masked on stdout: %q", r.stdout)
	}
}

func TestSecretMasking_EchoSuppressionUnchanged(t *testing.T) {
	// `@` still suppresses the echo entirely (no masked echo line either).
	src := "quietly:\n    @echo done {{env(\"ECHO_DEMO_TOKEN\")}}\n"
	dir := writeRunefile(t, src)
	r := run(t, dir, []string{"ECHO_DEMO_TOKEN=" + cliSecret}, "quietly")
	if strings.Contains(r.stderr, "echo") {
		t.Errorf("@-suppressed command was echoed: %q", r.stderr)
	}
	if !strings.Contains(r.stdout, "done ***") {
		t.Errorf("stdout = %q, want masked output", r.stdout)
	}

	// `set quiet` suppresses echo file-wide the same way.
	src = "set quiet\n\nloudly:\n    echo done {{env(\"ECHO_DEMO_TOKEN\")}}\n"
	dir = writeRunefile(t, src)
	r = run(t, dir, []string{"ECHO_DEMO_TOKEN=" + cliSecret}, "loudly")
	if strings.Contains(r.stderr, "echo") {
		t.Errorf("set quiet did not suppress the echo: %q", r.stderr)
	}
	if !strings.Contains(r.stdout, "done ***") {
		t.Errorf("stdout = %q, want masked output", r.stdout)
	}
}

// T015 — --dry-run runs nothing and leaks nothing. (Dry-run prints task names
// only, never interpolated bodies; this pins that no secret can appear.)
func TestSecretMasking_DryRunLeaksNothing(t *testing.T) {
	src := "leaky:\n    echo {{env(\"CLI_DEMO_TOKEN\")}}\n"
	dir := writeRunefile(t, src)
	r := run(t, dir, []string{"CLI_DEMO_TOKEN=" + cliSecret}, "--dry-run", "leaky")
	if r.code != 0 {
		t.Fatalf("exit = %d; stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, "would run: leaky") {
		t.Errorf("dry-run notice missing: %q", r.stderr)
	}
	if strings.Contains(r.stdout, cliSecret) || strings.Contains(r.stderr, cliSecret) {
		t.Fatalf("dry-run leaked a secret: stdout=%q stderr=%q", r.stdout, r.stderr)
	}
}

// T016 — a task interrupted mid-stream has only ever emitted masked bytes
// (masking happens at emission time, not as post-processing; FR-003).
func TestSecretMasking_InterruptedTaskNeverLeaks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("relies on the external sleep command")
	}
	src := "slow:\n    @echo \"tok $INT_DEMO_TOKEN\"\n    @sleep 30\n"
	dir := writeRunefile(t, src)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, runeBin, "slow")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "INT_DEMO_TOKEN="+cliSecret)
	cmd.WaitDelay = 2 * time.Second // don't wait for the orphaned sleep to release the pipes
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	_ = cmd.Run() // the kill makes this error by design

	if strings.Contains(out.String(), cliSecret) || strings.Contains(errb.String(), cliSecret) {
		t.Fatalf("interrupted run leaked a secret: stdout=%q stderr=%q", out.String(), errb.String())
	}
	if !strings.Contains(out.String(), "tok ***") {
		t.Errorf("pre-interrupt output missing or unmasked: %q", out.String())
	}
}

// T017 — FR-011: the task itself receives the REAL value (only emissions are
// masked, the process environment is never altered).
func TestSecretMasking_TaskEnvCarriesRealValue(t *testing.T) {
	src := "check:\n    @test \"$REAL_DEMO_TOKEN\" = \"" + cliSecret + "\" && echo match\n    @echo \"visible: $REAL_DEMO_TOKEN\"\n"
	dir := writeRunefile(t, src)
	r := run(t, dir, []string{"REAL_DEMO_TOKEN=" + cliSecret}, "check")
	if r.code != 0 {
		t.Fatalf("exit = %d; stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stdout, "match") {
		t.Errorf("task did not see the real value: %q", r.stdout)
	}
	if !strings.Contains(r.stdout, "visible: ***") || strings.Contains(r.stdout, cliSecret) {
		t.Errorf("emission not masked: %q", r.stdout)
	}
}

// T018 — FR-006: masking is always on; no environment variable or flag
// disables it (the only opt-out is per-variable `set unmasked`).
func TestSecretMasking_NoDisablePath(t *testing.T) {
	src := "leaky:\n    @echo \"tok $CLI_DEMO_TOKEN\"\n"
	dir := writeRunefile(t, src)
	r := run(t, dir, []string{
		"CLI_DEMO_TOKEN=" + cliSecret,
		"RUNE_NO_MASK=1", "RUNE_UNMASK=1", "NO_MASK=1",
	}, "leaky")
	if strings.Contains(r.stdout, cliSecret) {
		t.Fatalf("an env var disabled masking: %q", r.stdout)
	}
	if !strings.Contains(r.stdout, "tok ***") {
		t.Errorf("stdout = %q, want masked", r.stdout)
	}

	help := run(t, dir, nil, "--help")
	for _, forbidden := range []string{"--no-mask", "--unmask", "--disable-mask"} {
		if strings.Contains(help.stdout, forbidden) || strings.Contains(help.stderr, forbidden) {
			t.Errorf("--help advertises a masking disable flag %q", forbidden)
		}
	}
}

// T019 — SC-003: a secret-free Runefile produces byte-identical output.
func TestSecretMasking_NoSecretsByteIdentical(t *testing.T) {
	src := "hello:\n    @echo hi\n"
	dir := writeRunefile(t, src)
	r := run(t, dir, nil, "hello")
	if r.code != 0 {
		t.Fatalf("exit = %d", r.code)
	}
	if r.stdout != "hi\n" {
		t.Errorf("stdout = %q, want %q", r.stdout, "hi\n")
	}
	if r.stderr != "" {
		t.Errorf("stderr = %q, want empty", r.stderr)
	}
}
