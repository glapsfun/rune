package integration

import (
	"runtime"
	"strings"
	"testing"
)

// Spec 020 US2: directly invoking an OS-mismatched task fails with the
// availability message, exit 3, and zero execution — asserted against the
// real binary on the real host OS. The fixture is OS-complementary so the
// test is meaningful on every CI platform: unix hosts invoke the [windows]
// task, the windows host invokes the [unix] task.
func TestOSAvailability_DirectInvokeMismatchedExits3(t *testing.T) {
	dir := writeRunefile(t, ""+
		"[windows]\nwin-task:\n    @echo ran-win\n"+
		"[unix]\nnix-task:\n    @echo ran-nix\n")

	name, marker, requires := "win-task", "ran-win", "windows"
	if runtime.GOOS == "windows" {
		name, marker, requires = "nix-task", "ran-nix", "unix"
	}

	r := run(t, dir, nil, name)
	if r.code != 3 {
		t.Fatalf("exit = %d, want 3; stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, `task "`+name+`" is not available on `) {
		t.Errorf("stderr missing availability message: %q", r.stderr)
	}
	if !strings.Contains(r.stderr, "requires "+requires) {
		t.Errorf("stderr missing required OS %q: %q", requires, r.stderr)
	}
	if strings.Contains(r.stderr, "darwin") {
		t.Errorf("stderr leaks GOOS vocabulary (want macos): %q", r.stderr)
	}
	if strings.Contains(r.stdout, marker) {
		t.Errorf("mismatched task body executed: %q", r.stdout)
	}
}
