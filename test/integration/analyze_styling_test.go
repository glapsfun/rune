package integration

import (
	"strings"
	"testing"
)

// 014 US2 (C4): `rune analyze` renders diagnostics with the shared diagnostic
// renderer — severity color, faint locator, caret span — matching the run
// path's presentation, while keeping its coded severity token
// (error[RUNE2001]) and summary line. JSON output stays byte-frozen.

const analyzeBrokenRunefile = "# Deploy.\ndeploy: missing\n    @echo deploy\n"

// TestAnalyzeIsEnvIndependent: analyze is a static tool whose text output must
// be a pure function of the Runefile — it must not value-mask from the ambient
// environment (which would make output depend on the caller's shell and break
// reproducibility). A secret-named env var whose value equals a Runefile
// literal on a diagnostic line must NOT alter analyze's output.
func TestAnalyzeIsEnvIndependent(t *testing.T) {
	const literal = "AKIAsecretVALUE123"
	dir := writeRunefile(t, "deploy:\n    @echo "+literal+" {{nope}}\n")

	withEnv := run(t, dir, []string{"API_TOKEN=" + literal}, "analyze")
	withoutEnv := run(t, dir, nil, "analyze")
	if withEnv.stdout != withoutEnv.stdout {
		t.Errorf("analyze output depends on the environment:\n withEnv=%q\n withoutEnv=%q", withEnv.stdout, withoutEnv.stdout)
	}
	if !strings.Contains(withoutEnv.stdout, literal) {
		t.Errorf("analyze should echo the source literal verbatim: %q", withoutEnv.stdout)
	}
}

// TestAnalyzeTextMatchesRunRendering: the snippet (source line + caret) that a
// normal run prints for a broken Runefile appears identically in `rune
// analyze` output (modulo stream and the coded severity token).
func TestAnalyzeTextMatchesRunRendering(t *testing.T) {
	dir := writeRunefile(t, analyzeBrokenRunefile)

	runRes := run(t, dir, nil, "deploy")
	if runRes.code != 3 {
		t.Fatalf("run exit = %d, want 3", runRes.code)
	}
	anaRes := run(t, dir, nil, "analyze")
	if anaRes.code != 3 {
		t.Fatalf("analyze exit = %d, want 3; stdout=%q", anaRes.code, anaRes.stdout)
	}

	// The caret snippet lines must be shared verbatim between the two surfaces.
	for _, want := range []string{"2 | deploy: missing", "^^^^^^^"} {
		if !strings.Contains(runRes.stderr, want) {
			t.Errorf("run diagnostics missing %q: %q", want, runRes.stderr)
		}
		if !strings.Contains(anaRes.stdout, want) {
			t.Errorf("analyze diagnostics missing %q: %q", want, anaRes.stdout)
		}
	}
	// analyze keeps its coded severity token and summary.
	if !strings.Contains(anaRes.stdout, "error[RUNE2001]: unknown task: missing") {
		t.Errorf("analyze lost the coded diagnostic: %q", anaRes.stdout)
	}
	if !strings.Contains(anaRes.stdout, "1 error, 0 warnings") {
		t.Errorf("analyze lost the summary: %q", anaRes.stdout)
	}
}

// TestAnalyzeStyled: --color=always colors analyze's stdout (severity word,
// locator, caret, summary count) and stripping recovers the plain bytes.
// Both flag positions must work: --color is a persistent flag (C8/FR-008).
func TestAnalyzeStyled(t *testing.T) {
	dir := writeRunefile(t, analyzeBrokenRunefile)
	plain := run(t, dir, nil, "--color=never", "analyze")
	for name, styled := range map[string]result{
		"root position": run(t, dir, nil, "--color=always", "analyze"),
		"sub position":  run(t, dir, nil, "analyze", "--color=always"),
	} {
		if styled.code != 3 {
			t.Fatalf("%s: exit = %d, want 3; stderr=%q", name, styled.code, styled.stderr)
		}
		if !hasANSI(styled.stdout) {
			t.Errorf("%s: expected ANSI on styled analyze stdout: %q", name, styled.stdout)
		}
		if got := stripANSI(styled.stdout); got != plain.stdout {
			t.Errorf("%s: styled analyze (stripped) != plain:\n stripped=%q\n plain=%q", name, got, plain.stdout)
		}
	}
}

// TestAnalyzeCleanSummaryUnstyled: with zero diagnostics the summary's severity
// words carry no color even under --color=always (styled only when count > 0).
func TestAnalyzeCleanSummaryUnstyled(t *testing.T) {
	dir := writeRunefile(t, "# Build.\nbuild:\n    @echo build\n")
	r := run(t, dir, nil, "--color=always", "analyze")
	if r.code != 0 {
		t.Fatalf("clean analyze exit = %d; stdout=%q", r.code, r.stdout)
	}
	if !strings.Contains(r.stdout, "0 errors, 0 warnings") {
		t.Fatalf("summary changed: %q", r.stdout)
	}
	if hasANSI(r.stdout) {
		t.Errorf("zero-count summary should stay plain: %q", r.stdout)
	}
}

// TestAnalyzePlainInvariant: analyze text output is identical across the three
// color-off triggers with unchanged exit code.
func TestAnalyzePlainInvariant(t *testing.T) {
	dir := writeRunefile(t, analyzeBrokenRunefile)
	r := assertPlainInvariant(t, dir, "analyze")
	if r.code != 3 {
		t.Fatalf("exit = %d, want 3", r.code)
	}
}

// TestAnalyzeJSONByteFrozen: analyze --json is machine output — byte-identical
// across every color mode, never carrying ANSI (C7).
func TestAnalyzeJSONByteFrozen(t *testing.T) {
	dir := writeRunefile(t, analyzeBrokenRunefile)
	base := run(t, dir, nil, "analyze", "--json")
	always := run(t, dir, nil, "--color=always", "analyze", "--json")
	if hasANSI(base.stdout) || hasANSI(always.stdout) {
		t.Fatalf("JSON output carried ANSI")
	}
	if base.stdout != always.stdout {
		t.Errorf("analyze --json diverged across color modes:\n base=%q\n always=%q", base.stdout, always.stdout)
	}
	if base.code != 3 || always.code != 3 {
		t.Errorf("exit codes: base=%d always=%d, want 3", base.code, always.code)
	}
}
