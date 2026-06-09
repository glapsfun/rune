package agent

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func lookFalse() (string, error) { return exec.LookPath("false") }

func TestRunNotConfigured(t *testing.T) {
	_, err := CLIProvider{}.Run(context.Background(), "do something", Options{})
	var nc *NotConfiguredError
	if !errors.As(err, &nc) {
		t.Fatalf("err = %v, want NotConfiguredError (never invent credentials)", err)
	}
}

func TestRunMissingCLI(t *testing.T) {
	_, err := CLIProvider{}.Run(context.Background(), "prompt", Options{
		AgentCmd: []string{"definitely-not-a-real-agent-cli-xyz", "-p"},
	})
	var ni *NotInstalledError
	if !errors.As(err, &ni) {
		t.Fatalf("err = %v, want NotInstalledError", err)
	}
	if !strings.Contains(ni.Error(), "not found on PATH") {
		t.Errorf("message not actionable: %q", ni.Error())
	}
}

func TestRunCapturesOutput(t *testing.T) {
	// Use a real always-present binary as a stand-in "agent CLI": echo prints
	// the prompt, which we capture as the final text.
	out, err := CLIProvider{}.Run(context.Background(), "hello-prompt", Options{
		AgentCmd: []string{"echo"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "hello-prompt") {
		t.Errorf("captured output = %q, want it to contain the prompt", out)
	}
}

func TestRunFailureIsAuthError(t *testing.T) {
	// `false` exits non-zero, modeling an unauthenticated/failed agent CLI.
	if _, err := lookFalse(); err != nil {
		t.Skip("no `false` binary available")
	}
	_, err := CLIProvider{}.Run(context.Background(), "prompt", Options{AgentCmd: []string{"false"}})
	var ae *AuthError
	if !errors.As(err, &ae) {
		t.Fatalf("err = %v, want AuthError", err)
	}
}
