package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func analyzeInDir(t *testing.T, files map[string]string, path string, jsonOut bool) (string, error) {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		writeFile(t, dir, name, content)
	}
	var buf bytes.Buffer
	opts := Options{Stdout: &buf, Stderr: io.Discard, Cwd: dir}
	target := path
	if target != "" {
		target = filepath.Join(dir, path)
	}
	err := Analyze(opts, target, jsonOut)
	return buf.String(), err
}

func TestAnalyzeHumanOutputAndExit3(t *testing.T) {
	out, err := analyzeInDir(t, map[string]string{
		"Runefile": "# Build.\nbuild:\n    @echo build\n# Deploy.\ndeploy: missing\n    @echo deploy\n",
	}, "Runefile", false)

	if !strings.Contains(out, "error[RUNE2001]: unknown task: missing") {
		t.Errorf("output missing coded diagnostic:\n%s", out)
	}
	if !strings.Contains(out, "1 error, 0 warnings") {
		t.Errorf("output missing summary:\n%s", out)
	}
	// Error diagnostics -> ValidationError -> exit 3.
	if got := CodeFor(err); got != ExitValidation {
		t.Errorf("exit code = %d, want %d (validation)", got, ExitValidation)
	}
}

func TestAnalyzeCleanExit0(t *testing.T) {
	out, err := analyzeInDir(t, map[string]string{
		"Runefile": "# Build.\nbuild:\n    @echo build\n",
	}, "Runefile", false)
	if err != nil {
		t.Errorf("clean Runefile should return nil error, got %v", err)
	}
	if got := CodeFor(err); got != ExitSuccess {
		t.Errorf("exit code = %d, want 0", got)
	}
	if !strings.Contains(out, "0 errors, 0 warnings") {
		t.Errorf("summary = %q", out)
	}
}

func TestAnalyzeUndocumentedWarningDoesNotGate(t *testing.T) {
	// An undocumented public task is a warning (RUNE2010) and must NOT cause exit 3.
	out, err := analyzeInDir(t, map[string]string{
		"Runefile": "build:\n    @echo build\n",
	}, "Runefile", false)
	if !strings.Contains(out, "warning[RUNE2010]") {
		t.Errorf("expected RUNE2010 warning, got:\n%s", out)
	}
	if got := CodeFor(err); got != ExitSuccess {
		t.Errorf("warnings must not gate: exit = %d, want 0", got)
	}
}

func TestAnalyzeMissingRunefileExit2(t *testing.T) {
	var buf bytes.Buffer
	opts := Options{Stdout: &buf, Stderr: io.Discard, Cwd: t.TempDir()}
	err := Analyze(opts, "", false)
	if got := CodeFor(err); got != ExitUsage {
		t.Errorf("missing Runefile: exit = %d, want %d (usage)", got, ExitUsage)
	}
}

func TestAnalyzeJSONOutput(t *testing.T) {
	out, err := analyzeInDir(t, map[string]string{
		"Runefile": "# Build.\nbuild:\n    @echo build\n# Deploy.\ndeploy: missing\n    @echo deploy\n",
	}, "Runefile", true)
	if CodeFor(err) != ExitValidation {
		t.Errorf("json run exit = %d, want 3", CodeFor(err))
	}
	var report jsonReport
	if e := json.Unmarshal([]byte(out), &report); e != nil {
		t.Fatalf("invalid JSON: %v\n%s", e, out)
	}
	if report.Errors != 1 || report.Warnings != 0 {
		t.Errorf("summary = %+v, want 1 error / 0 warnings", report)
	}
	if len(report.Diagnostics) != 1 || report.Diagnostics[0].Code != "RUNE2001" {
		t.Errorf("diagnostics = %+v", report.Diagnostics)
	}
	if report.Diagnostics[0].Range.Start.Line != 5 {
		t.Errorf("range start line = %d, want 5", report.Diagnostics[0].Range.Start.Line)
	}
}
