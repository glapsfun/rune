package integration

import (
	"os/exec"
	"strings"
	"testing"
)

// TestSecretMasking_InterpExecutorMasked proves masking is
// executor-independent: the python (interp) path streams through the same
// wrapped writers as the shell path (contract §2.3 item 1).
func TestSecretMasking_InterpExecutorMasked(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available")
	}
	src := "show (python):\n    import os\n    print(\"py token is \" + os.environ.get(\"PY_DEMO_TOKEN\", \"unset\"))\n"
	dir := writeRunefile(t, src)
	r := run(t, dir, []string{"PY_DEMO_TOKEN=hunter2-py-secret"}, "show")
	if r.code != 0 {
		t.Fatalf("exit = %d, want 0; stderr=%s", r.code, r.stderr)
	}
	if strings.Contains(r.stdout, "hunter2-py-secret") {
		t.Fatalf("raw secret leaked on the interp path: %q", r.stdout)
	}
	if !strings.Contains(r.stdout, "py token is ***") {
		t.Errorf("stdout = %q, want masked token line", r.stdout)
	}
}
