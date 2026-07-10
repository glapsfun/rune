package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/rune-task-runner/rune/internal/config"
	"github.com/rune-task-runner/rune/internal/parser"
)

// TestRenderVersionMismatchGolden pins the incompatibility diagnostic (caret on
// the value literal, installed/required/upgrade notes, trailer). The Runefile
// path is fixed so the output is deterministic.
func TestRenderVersionMismatchGolden(t *testing.T) {
	src := "set minimum_version := \"0.8.0\"\nbuild:\n    @echo hi\n"
	file, _ := parser.Parse("Runefile", src)
	req, diags := config.MinimumVersion(file)
	if diags.HasErrors() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	var buf bytes.Buffer
	opts := Options{Stderr: &buf, Version: "0.7.2"} // ColorStderr false => plain theme
	renderVersionMismatch(opts, req, newSourceProvider("Runefile", []byte(src)))

	golden := filepath.Join("testdata", "minimum_version_incompatible.golden")
	if *update {
		if err := os.WriteFile(golden, buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden (run with -update to create): %v", err)
	}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("rendered mismatch:\n--- got ---\n%s\n--- want ---\n%s", buf.String(), want)
	}
}
