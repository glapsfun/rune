package integration

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// escRe matches SGR (color) escape sequences emitted by the styled surfaces.
var escRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

// stripANSI removes SGR escapes, recovering the visible text. Stripped styled
// output must equal the plain output byte-for-byte (zero-width emphasis).
func stripANSI(s string) string { return escRe.ReplaceAllString(s, "") }

// hasANSI reports whether s contains any ESC byte.
func hasANSI(s string) bool { return strings.ContainsRune(s, '\x1b') }

// assertPlainInvariant runs args three ways that MUST all produce identical,
// ANSI-free output: piped (default non-TTY), NO_COLOR=1, and --color=never
// (FR-010, SC-001). Global flags must precede positionals, so --color=never is
// prepended. Returns the shared plain result for further assertions.
func assertPlainInvariant(t *testing.T, dir string, args ...string) result {
	t.Helper()
	piped := run(t, dir, nil, args...)
	noColor := run(t, dir, []string{"NO_COLOR=1"}, args...)
	never := run(t, dir, nil, append([]string{"--color=never"}, args...)...)

	for _, c := range []struct {
		name string
		res  result
	}{{"piped", piped}, {"NO_COLOR", noColor}, {"--color=never", never}} {
		if hasANSI(c.res.stdout) {
			t.Errorf("%s: stdout contained ANSI escape: %q", c.name, c.res.stdout)
		}
		if hasANSI(c.res.stderr) {
			t.Errorf("%s: stderr contained ANSI escape: %q", c.name, c.res.stderr)
		}
	}
	if noColor.stdout != piped.stdout || never.stdout != piped.stdout {
		t.Errorf("stdout diverged across off-modes:\n piped=%q\n no_color=%q\n never=%q", piped.stdout, noColor.stdout, never.stdout)
	}
	if noColor.stderr != piped.stderr || never.stderr != piped.stderr {
		t.Errorf("stderr diverged across off-modes:\n piped=%q\n no_color=%q\n never=%q", piped.stderr, noColor.stderr, never.stderr)
	}
	// Exit codes must be identical across color modes (SC-005).
	if noColor.code != piped.code || never.code != piped.code {
		t.Errorf("exit code diverged across off-modes: piped=%d no_color=%d never=%d", piped.code, noColor.code, never.code)
	}
	return piped
}

// assertStyledStderrMatchesPlain checks that a --color=always run emits ANSI on
// stderr and that stripping it recovers the plain run's stderr byte-for-byte.
func assertStyledStderrMatchesPlain(t *testing.T, styled, plain result) {
	t.Helper()
	if !hasANSI(styled.stderr) {
		t.Errorf("expected ANSI on styled stderr, got %q", styled.stderr)
	}
	if got := stripANSI(styled.stderr); got != plain.stderr {
		t.Errorf("styled stderr (stripped) != plain:\n stripped=%q\n plain   =%q", got, plain.stderr)
	}
}

// cacheRunefile is a cached task whose body echoes a command; exercises the
// running:/cached: labels and command-echo dimming.
const cacheRunefile = "[cache(inputs = [\"in.txt\"], outputs = [\"out.txt\"])]\nbuild:\n    cp in.txt out.txt\n"

func setupCacheDir(t *testing.T) string {
	t.Helper()
	dir := writeRunefile(t, cacheRunefile)
	if err := os.WriteFile(filepath.Join(dir, "in.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestUS3_RunOutputStyled: status labels are styled and command echo / cache
// notices are dimmed, while stripping ANSI yields the exact plain bytes and the
// stream assignment is unchanged (FR-014, FR-016).
func TestUS3_RunOutputStyled(t *testing.T) {
	ds, dp := setupCacheDir(t), setupCacheDir(t)

	// First run: "running: build" label + dimmed echo of "cp in.txt out.txt".
	s1 := run(t, ds, nil, "--color=always", "build")
	p1 := run(t, dp, nil, "build")
	if s1.code != 0 || p1.code != 0 {
		t.Fatalf("first run exit: styled=%d plain=%d (%q)", s1.code, p1.code, s1.stderr)
	}
	assertStyledStderrMatchesPlain(t, s1, p1)
	if !strings.Contains(p1.stderr, "running: build") {
		t.Errorf("first run missing 'running: build': %q", p1.stderr)
	}
	if !strings.Contains(p1.stderr, "cp in.txt out.txt") {
		t.Errorf("first run missing echoed command: %q", p1.stderr)
	}

	// Second run: "cached: build" (dimmed), no echo.
	s2 := run(t, ds, nil, "--color=always", "build")
	p2 := run(t, dp, nil, "build")
	assertStyledStderrMatchesPlain(t, s2, p2)
	if !strings.Contains(p2.stderr, "cached: build") {
		t.Errorf("second run missing 'cached: build': %q", p2.stderr)
	}
}

// TestUS3_DryRunLabelStyled: the "would run:" dry-run notice is styled and
// strips back to the plain bytes.
func TestUS3_DryRunLabelStyled(t *testing.T) {
	dir := setupCacheDir(t)
	s := run(t, dir, nil, "--color=always", "--dry-run", "build")
	p := run(t, dir, nil, "--dry-run", "build")
	assertStyledStderrMatchesPlain(t, s, p)
	if !strings.Contains(p.stderr, "would run: build") {
		t.Errorf("dry-run missing 'would run: build': %q", p.stderr)
	}
}

// diagRunefile triggers a static analyzer error (undefined variable) so the
// diagnostic renderer (file:line:col + caret) runs on stderr with exit 3.
const diagRunefile = "task t {\n  @echo {{nope}}\n}\n"

// TestUS2_ListInvariant: --list (stdout surface) is plain and identical across
// all three off-triggers.
func TestUS2_ListInvariant(t *testing.T) {
	dir := writeRunefile(t, us1Runefile)
	assertPlainInvariant(t, dir, "--list")
}

// TestUS2_EmptyListInvariant: when every task is filtered out (all private),
// the styled --list falls back to the same header-only output as today and is
// identical across off-triggers (spec Edge Cases).
func TestUS2_EmptyListInvariant(t *testing.T) {
	dir := writeRunefile(t, "_hidden:\n    @echo hi\n")
	plain := assertPlainInvariant(t, dir, "--list")
	if !strings.Contains(plain.stdout, "Available tasks:") {
		t.Errorf("empty --list should still print the header: %q", plain.stdout)
	}
	// Styled empty list strips back to the same bytes (no stray ANSI around an
	// absent row).
	styled := run(t, dir, nil, "--color=always", "--list")
	if stripANSI(styled.stdout) != plain.stdout {
		t.Errorf("styled empty --list (stripped) != plain:\n stripped=%q\n plain=%q", stripANSI(styled.stdout), plain.stdout)
	}
}

// TestUS2_DiagnosticInvariant: diagnostics (stderr surface) are plain and
// identical across off-triggers, and the exit code is unchanged (3).
func TestUS2_DiagnosticInvariant(t *testing.T) {
	dir := writeRunefile(t, diagRunefile)
	r := assertPlainInvariant(t, dir, "t")
	if r.code != 3 {
		t.Fatalf("diagnostic exit = %d, want 3 (stderr=%q)", r.code, r.stderr)
	}
}

// TestUS4_InvalidColorValueErrors: an invalid --color value is a usage error
// (exit 2) printed to stderr, and nothing runs (FR-009).
func TestUS4_InvalidColorValueErrors(t *testing.T) {
	dir := writeRunefile(t, us1Runefile)
	r := run(t, dir, nil, "--color=sometimes", "--list")
	if r.code != 2 {
		t.Errorf("invalid --color exit = %d, want 2; stderr=%q", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, "invalid --color") {
		t.Errorf("missing clear error message: %q", r.stderr)
	}
	if strings.Contains(r.stdout, "Available tasks") {
		t.Errorf("listed tasks despite invalid flag: %q", r.stdout)
	}
}

// TestUS4_AlwaysAndNever: --color=always colors a piped stdout; --color=never
// stays plain (FR-007). Per-stream resolution (stdout + stderr both colored
// under --color=always) is covered by TestUS2_ForcedColorEmitsANSI.
func TestUS4_AlwaysAndNever(t *testing.T) {
	dir := writeRunefile(t, us1Runefile)
	if got := run(t, dir, nil, "--color=always", "--list"); !hasANSI(got.stdout) {
		t.Errorf("--color=always should color stdout, got %q", got.stdout)
	}
	if got := run(t, dir, nil, "--color=never", "--list"); hasANSI(got.stdout) {
		t.Errorf("--color=never should not color, got %q", got.stdout)
	}
}

// TestUS2_ForcedColorEmitsANSI: --color=always is the one path that colors
// through a pipe, on both stdout (--list) and stderr (diagnostic) (SC-004).
func TestUS2_ForcedColorEmitsANSI(t *testing.T) {
	listDir := writeRunefile(t, us1Runefile)
	if got := run(t, listDir, nil, "--color=always", "--list"); !hasANSI(got.stdout) {
		t.Errorf("--color=always --list: expected ANSI on stdout, got %q", got.stdout)
	}
	diagDir := writeRunefile(t, diagRunefile)
	if got := run(t, diagDir, nil, "--color=always", "t"); !hasANSI(got.stderr) {
		t.Errorf("--color=always (diagnostic): expected ANSI on stderr, got %q", got.stderr)
	}
}
