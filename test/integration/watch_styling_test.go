package integration

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// 014 US1/US2 (C1, C3): watch mode's per-iteration failure banner uses the
// shared error styling and its chrome lines ("watching …") are dimmed; plain
// mode stays byte-frozen. Watch keeps running after a styled failure.

// runWatch starts `rune [args…]` and hard-stops it after the timeout,
// returning whatever both streams accumulated (watch never exits on its own).
func runWatch(t *testing.T, dir string, timeout time.Duration, args ...string) (stdout, stderr string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, runeBin, args...)
	cmd.Dir = dir
	cmd.WaitDelay = 2 * time.Second
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	_ = cmd.Run() // killed by the deadline by design
	return out.String(), errb.String()
}

func TestWatchBannerStyled(t *testing.T) {
	dir := writeRunefile(t, "boom:\n    @exit 7\n")

	_, styled := runWatch(t, dir, 2*time.Second, "--color=always", "--watch", "boom")
	_, plain := runWatch(t, dir, 2*time.Second, "--color=never", "--watch", "boom")

	// Both modes show the chrome line and the per-iteration failure banner.
	for name, s := range map[string]string{"styled": styled, "plain": plain} {
		if !strings.Contains(stripANSI(s), "watching ") {
			t.Fatalf("%s: missing watch chrome: %q", name, s)
		}
		if !strings.Contains(stripANSI(s), `rune: task "boom" failed`) {
			t.Fatalf("%s: missing per-iteration failure banner: %q", name, s)
		}
	}
	if hasANSI(plain) {
		t.Errorf("--color=never watch output carried ANSI: %q", plain)
	}
	if !hasANSI(styled) {
		t.Errorf("--color=always watch output should carry ANSI: %q", styled)
	}
	// Ordering held: the chrome line precedes the first (failing) run, i.e. the
	// loop printed its banner and kept watching rather than exiting.
	if got := stripANSI(styled); !strings.HasPrefix(got, "watching ") {
		t.Errorf("watch chrome should precede the first run: %q", got)
	}
}

// TestWatchNoRedundantValidationBanner: a re-run that fails static validation
// (here a dependency cycle) has its diagnostics rendered by the pipeline once;
// watch must not additionally print the "rune: <msg>" banner for the returned
// ValidationError, matching the non-watch path. Regression for the code-review
// finding on the doubled banner under --watch.
func TestWatchNoRedundantValidationBanner(t *testing.T) {
	dir := writeRunefile(t, "c: c\n    @echo ran\n")
	_, stderr := runWatch(t, dir, 2*time.Second, "--watch", "c")

	// The diagnostic (with its caret) is shown once...
	if !strings.Contains(stderr, "dependency cycle: c") {
		t.Fatalf("watch did not render the cycle diagnostic: %q", stderr)
	}
	// ...but the redundant "rune: dependency cycle" banner must not appear.
	if strings.Contains(stderr, "rune: dependency cycle") {
		t.Errorf("watch printed the redundant validation banner: %q", stderr)
	}
}

// TestWatchPlainBytesFrozen: the plain watch output (chrome + banner) is
// byte-identical between piped auto mode and --color=never (FR-006).
func TestWatchPlainBytesFrozen(t *testing.T) {
	dir := writeRunefile(t, "boom:\n    @exit 7\n")
	_, piped := runWatch(t, dir, 2*time.Second, "--watch", "boom")
	_, never := runWatch(t, dir, 2*time.Second, "--color=never", "--watch", "boom")
	if hasANSI(piped) || hasANSI(never) {
		t.Fatalf("plain watch output carried ANSI: piped=%q never=%q", piped, never)
	}
	if piped != never {
		t.Errorf("plain watch output diverged:\n piped=%q\n never=%q", piped, never)
	}
}
