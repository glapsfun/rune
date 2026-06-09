package integration

import (
	"os/exec"
	"strings"
	"testing"
)

const us3Runefile = `build_dir := "dist"

analyze (python):
    print("coverage from {{build_dir}}")

bundle (node):
    console.log("bundling " + "{{build_dir}}")
`

func TestUS3_PythonBody(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available")
	}
	dir := writeRunefile(t, us3Runefile)
	r := run(t, dir, nil, "analyze")
	if r.code != 0 {
		t.Fatalf("exit = %d, stderr=%s", r.code, r.stderr)
	}
	if strings.TrimSpace(r.stdout) != "coverage from dist" {
		t.Errorf("stdout = %q, want 'coverage from dist'", r.stdout)
	}
}

func TestUS3_NodeBody(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not available")
	}
	dir := writeRunefile(t, us3Runefile)
	r := run(t, dir, nil, "bundle")
	if r.code != 0 {
		t.Fatalf("exit = %d, stderr=%s", r.code, r.stderr)
	}
	if strings.TrimSpace(r.stdout) != "bundling dist" {
		t.Errorf("stdout = %q, want 'bundling dist'", r.stdout)
	}
}

func TestUS3_MissingInterpreter(t *testing.T) {
	dir := writeRunefile(t, "weird (no_such_interp_xyz):\n    do something\n")
	r := run(t, dir, nil, "weird")
	if r.code != 1 {
		t.Fatalf("exit = %d, want 1; stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, "not found on PATH") {
		t.Errorf("stderr not actionable: %q", r.stderr)
	}
}

func TestUS3_ShellAndPythonCoexist(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available")
	}
	src := "hello:\n    @echo shell-says hi\n\nanalyze (python):\n    print(\"python-says hi\")\n"
	dir := writeRunefile(t, src)
	if r := run(t, dir, nil, "hello"); r.code != 0 || !strings.Contains(r.stdout, "shell-says hi") {
		t.Errorf("shell task: code=%d stdout=%q", r.code, r.stdout)
	}
	if r := run(t, dir, nil, "analyze"); r.code != 0 || !strings.Contains(r.stdout, "python-says hi") {
		t.Errorf("python task: code=%d stdout=%q", r.code, r.stdout)
	}
}
