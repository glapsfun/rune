package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestGatherContextSuccess(t *testing.T) {
	a := adapterFor(t, "[context]\nhealth:\n    @echo on-branch-main\nbuild:\n    @echo b\n")
	var errb bytes.Buffer
	text := a.gatherContext(context.Background(), &errb)
	if text != "on-branch-main" {
		t.Errorf("text = %q, want on-branch-main", text)
	}
	if errb.Len() != 0 {
		t.Errorf("no warning expected, got %q", errb.String())
	}
}

func TestGatherContextNoHook(t *testing.T) {
	a := adapterFor(t, "build:\n    @echo b\n")
	if text := a.gatherContext(context.Background(), &bytes.Buffer{}); text != "" {
		t.Errorf("no hook: got %q, want empty", text)
	}
}

func TestGatherContextOSMismatchTreatedAsAbsent(t *testing.T) {
	a := adapterFor(t, "[context]\n[windows]\nhealth:\n    @echo w\n")
	a.goos = "linux"
	if text := a.gatherContext(context.Background(), &bytes.Buffer{}); text != "" {
		t.Errorf("mismatched hook: got %q, want empty", text)
	}
}

// A hook that succeeds but prints nothing yields no context at all, so no
// surface injects a bare header (FR-009).
func TestGatherContextEmptyOutputTreatedAsAbsent(t *testing.T) {
	a := adapterFor(t, "[context]\nhealth:\n    @true\n")
	var errb bytes.Buffer
	if text := a.gatherContext(context.Background(), &errb); text != "" {
		t.Errorf("empty-output hook: got %q, want empty", text)
	}
	if errb.Len() != 0 {
		t.Errorf("empty output is not a failure; no warning expected, got %q", errb.String())
	}
}

func TestGatherContextFailureDegrades(t *testing.T) {
	a := adapterFor(t, "[context]\nhealth:\n    @exit 7\n")
	var errb bytes.Buffer
	text := a.gatherContext(context.Background(), &errb)
	want := `(context hook "health" failed; proceeding without project context)`
	if text != want {
		t.Errorf("text = %q, want %q", text, want)
	}
	if !strings.Contains(errb.String(), "context hook") {
		t.Errorf("stderr should carry a warning, got %q", errb.String())
	}
}

func TestGatherContextTimeoutDegrades(t *testing.T) {
	a := adapterFor(t, "[context]\nhealth:\n    @sleep 5\n")
	a.contextTimeout = 50 * time.Millisecond
	var errb bytes.Buffer
	start := time.Now()
	text := a.gatherContext(context.Background(), &errb)
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Fatalf("timeout did not cut the hook off (took %v) — context propagation broken?", elapsed)
	}
	if !strings.Contains(text, "failed; proceeding without project context") {
		t.Errorf("timeout should degrade to the notice, got %q", text)
	}
	// The shell's exec handler holds a 2s kill-grace timer after a context
	// cancel; wait it out so the package goleak check sees a clean state.
	time.Sleep(2100 * time.Millisecond)
}

func TestGatherContextTruncates(t *testing.T) {
	// Emit ~9000 bytes; the cap is 8192 bytes of content plus the marker.
	// A single builtin (no pipeline) keeps the capture path single-writer:
	// piped segments share the captured stderr and race in os/exec copiers —
	// a pre-existing Call() hazard tracked outside this feature.
	a := adapterFor(t, "[context]\nhealth:\n    @printf '%09000d' 0\n")
	text := a.gatherContext(context.Background(), &bytes.Buffer{})
	if text == "" {
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

// The byte-wise cap must never split a multi-byte rune: the cut backs up to
// a rune boundary so the injected text stays valid UTF-8 for the JSON-encoded
// MCP instructions and the provider prompt.
func TestGatherContextTruncatesOnRuneBoundary(t *testing.T) {
	// 3-byte runes (…, U+2026) straddle the 8192-byte cap: 8192 % 3 != 0.
	a := adapterFor(t, "[context]\nhealth:\n    @awk 'BEGIN { for (i = 0; i < 3000; i++) printf \"…\" }'\n")
	text := a.gatherContext(context.Background(), &bytes.Buffer{})
	if !strings.HasSuffix(text, "[truncated]") {
		t.Fatalf("expected truncation, got %d bytes", len(text))
	}
	if !utf8.ValidString(text) {
		t.Fatal("truncation split a rune: context text is not valid UTF-8")
	}
}

func TestGatherContextMasksSecrets(t *testing.T) {
	a := adapterFor(t, "[context]\nhealth:\n    @echo token=$MY_API_TOKEN\n")
	a.baseEnv = []string{"MY_API_TOKEN=hunter2-super-secret"}
	a.maskSet = deriveMaskSet(a.baseEnv, a.tasks, a.settings.Secrets, a.settings.Unmasked)
	text := a.gatherContext(context.Background(), &bytes.Buffer{})
	if strings.Contains(text, "hunter2-super-secret") {
		t.Fatalf("secret leaked into context text: %q", text)
	}
	if !strings.Contains(text, "***") {
		t.Errorf("masked placeholder expected, got %q", text)
	}
}
