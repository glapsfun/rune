package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestGatherContextSuccess(t *testing.T) {
	a := adapterFor(t, "[context]\nhealth:\n    @echo on-branch-main\nbuild:\n    @echo b\n")
	var errb bytes.Buffer
	text, ok := a.gatherContext(context.Background(), &errb)
	if !ok {
		t.Fatal("hook should be detected")
	}
	if text != "on-branch-main" {
		t.Errorf("text = %q, want on-branch-main", text)
	}
	if errb.Len() != 0 {
		t.Errorf("no warning expected, got %q", errb.String())
	}
}

func TestGatherContextNoHook(t *testing.T) {
	a := adapterFor(t, "build:\n    @echo b\n")
	text, ok := a.gatherContext(context.Background(), &bytes.Buffer{})
	if ok || text != "" {
		t.Errorf("no hook: got (%q, %v), want (\"\", false)", text, ok)
	}
}

func TestGatherContextOSMismatchTreatedAsAbsent(t *testing.T) {
	a := adapterFor(t, "[context]\n[windows]\nhealth:\n    @echo w\n")
	a.goos = "linux"
	text, ok := a.gatherContext(context.Background(), &bytes.Buffer{})
	if ok || text != "" {
		t.Errorf("mismatched hook: got (%q, %v), want (\"\", false)", text, ok)
	}
}

func TestGatherContextFailureDegrades(t *testing.T) {
	a := adapterFor(t, "[context]\nhealth:\n    @exit 7\n")
	var errb bytes.Buffer
	text, ok := a.gatherContext(context.Background(), &errb)
	if !ok {
		t.Fatal("hook exists even when it fails")
	}
	want := `(context hook "health" failed; proceeding without project context)`
	if text != want {
		t.Errorf("text = %q, want %q", text, want)
	}
	if !strings.Contains(errb.String(), "context hook") {
		t.Errorf("stderr should carry a warning, got %q", errb.String())
	}
}

func TestGatherContextTimeoutDegrades(t *testing.T) {
	old := contextTimeout
	contextTimeout = 50 * time.Millisecond
	defer func() { contextTimeout = old }()

	a := adapterFor(t, "[context]\nhealth:\n    @sleep 5\n")
	var errb bytes.Buffer
	start := time.Now()
	text, ok := a.gatherContext(context.Background(), &errb)
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Fatalf("timeout did not cut the hook off (took %v) — context propagation broken?", elapsed)
	}
	if !ok || !strings.Contains(text, "failed; proceeding without project context") {
		t.Errorf("timeout should degrade to the notice, got (%q, %v)", text, ok)
	}
	// The shell's exec handler holds a 2s kill-grace timer after a context
	// cancel; wait it out so the package goleak check sees a clean state.
	time.Sleep(2100 * time.Millisecond)
}

func TestGatherContextTruncates(t *testing.T) {
	// Emit ~9000 bytes; the cap is 8192 bytes of content plus the marker.
	a := adapterFor(t, "[context]\nhealth:\n    @head -c 9000 /dev/zero | tr '\\0' 'x'\n")
	text, ok := a.gatherContext(context.Background(), &bytes.Buffer{})
	if !ok {
		t.Fatal("hook should run")
	}
	if !strings.HasSuffix(text, "[truncated]") {
		tail := text
		if len(tail) > 20 {
			tail = tail[len(tail)-20:]
		}
		t.Errorf("truncated output must end with the marker, got tail %q", tail)
	}
	if len(text) > contextMaxBytes+len("\n[truncated]") {
		t.Errorf("text length %d exceeds cap", len(text))
	}
}

func TestGatherContextMasksSecrets(t *testing.T) {
	a := adapterFor(t, "[context]\nhealth:\n    @echo token=$MY_API_TOKEN\n")
	a.baseEnv = []string{"MY_API_TOKEN=hunter2-super-secret"}
	a.maskSet = deriveMaskSet(a.baseEnv, a.tasks, a.settings.Secrets, a.settings.Unmasked)
	text, _ := a.gatherContext(context.Background(), &bytes.Buffer{})
	if strings.Contains(text, "hunter2-super-secret") {
		t.Fatalf("secret leaked into context text: %q", text)
	}
	if !strings.Contains(text, "***") {
		t.Errorf("masked placeholder expected, got %q", text)
	}
}
