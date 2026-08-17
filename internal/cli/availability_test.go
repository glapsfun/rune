package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/rune-task-runner/rune/internal/config"
	"github.com/rune-task-runner/rune/internal/eval"
	"github.com/rune-task-runner/rune/internal/parser"
	"github.com/rune-task-runner/rune/internal/runtime/scheduler"
)

// availEngine builds an engine over src with an injected host OS, mirroring
// the construction in mcpAdapter.Call so availability behavior is testable
// for any platform from any platform (spec 020 FR-007).
func availEngine(t *testing.T, src, goos string, plan planMode, stdout *bytes.Buffer) *engine {
	t.Helper()
	f, diags := parser.Parse("Runefile", src)
	if diags.HasErrors() {
		t.Fatalf("parse: %v", diags)
	}
	assigns := indexAssignments(f)
	scope := eval.NewScope(assigns, map[string]string{})
	scope.GOOS = goos
	settings, _ := config.ResolveSettings(f, eval.New(scope))
	var errBuf bytes.Buffer
	return &engine{
		file:      f,
		src:       newSourceProvider("Runefile", []byte(src)),
		tasks:     indexTasks(f),
		assigns:   assigns,
		overrides: map[string]string{},
		scope:     scope,
		settings:  settings,
		workDir:   t.TempDir(),
		root:      t.TempDir(),
		opts:      Options{Stdout: stdout, Stderr: &errBuf, Quiet: true},
		plan:      plan,
		now:       func() string { return "" },
		ctx:       context.Background(),
		goos:      goos,
	}
}

const dispatchSrc = "" +
	"setup: setup-nix setup-win\n    @echo setup-done\n" +
	"[unix]\nsetup-nix:\n    @echo unix-setup\n" +
	"[windows]\nsetup-win:\n    @echo windows-setup\n"

// Direct invocation of an OS-mismatched task must fail with a ValidationError
// (exit 3, nothing executed) whose message uses the attribute vocabulary
// (spec 020 US2; contracts/availability.md).
func TestResolveRootsRejectsOSMismatchedTask(t *testing.T) {
	var out bytes.Buffer
	// All plan modes flow through resolveRoots, so run/--dry-run/--summary
	// reject identically (research D3).
	for _, plan := range []planMode{planRun, planDryRun, planSummary} {
		eng := availEngine(t, dispatchSrc, "linux", plan, &out)
		_, err := eng.resolveRoots([]rawInvocation{{name: "setup-win"}})
		if err == nil {
			t.Fatalf("plan %v: mismatched root accepted", plan)
		}
		var verr *ValidationError
		if !errors.As(err, &verr) {
			t.Errorf("plan %v: error type = %T, want *ValidationError", plan, err)
		}
		if code := CodeFor(err); code != ExitValidation {
			t.Errorf("plan %v: CodeFor = %d, want %d", plan, code, ExitValidation)
		}
		for _, part := range []string{`"setup-win"`, "not available on linux", "requires windows"} {
			if !strings.Contains(err.Error(), part) {
				t.Errorf("plan %v: message %q missing %q", plan, err.Error(), part)
			}
		}
	}
	if out.Len() != 0 {
		t.Errorf("rejected invocation produced output: %q", out.String())
	}
}

// A multi-task invocation with one mismatched root executes nothing at all:
// the gate fires during root resolution, before the scheduler starts
// (spec 020 Story 2 scenario 3).
func TestResolveRootsMismatchAbortsWholeInvocation(t *testing.T) {
	var out bytes.Buffer
	eng := availEngine(t, dispatchSrc, "linux", planRun, &out)
	invs, err := eng.resolveRoots([]rawInvocation{{name: "setup-nix"}, {name: "setup-win"}})
	if err == nil {
		t.Fatal("mixed invocation accepted")
	}
	if invs != nil {
		t.Errorf("roots returned alongside error: %v", invs)
	}
	if out.Len() != 0 {
		t.Errorf("output produced despite aborted resolution: %q", out.String())
	}
}

// The host renders in attribute vocabulary: a Mac reads "macos", never the
// internal GOOS name "darwin" (analyze remediation I1).
func TestAvailabilityErrorUsesAttributeVocabulary(t *testing.T) {
	var out bytes.Buffer
	eng := availEngine(t, dispatchSrc, "darwin", planRun, &out)
	_, err := eng.resolveRoots([]rawInvocation{{name: "setup-win"}})
	if err == nil {
		t.Fatal("mismatched root accepted")
	}
	if strings.Contains(err.Error(), "darwin") {
		t.Errorf("message leaks GOOS vocabulary: %q", err.Error())
	}
	if !strings.Contains(err.Error(), "not available on macos") {
		t.Errorf("message missing attribute-vocabulary host: %q", err.Error())
	}
}

// A [unix] task is available on darwin (spec 020 Story 2 scenario 2) and
// runs normally end to end.
func TestUnixTaskRunsOnDarwin(t *testing.T) {
	var out bytes.Buffer
	eng := availEngine(t, dispatchSrc, "darwin", planRun, &out)
	invs, err := eng.resolveRoots([]rawInvocation{{name: "setup-nix"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := scheduler.Run(eng, invs); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "unix-setup") {
		t.Errorf("unix task did not run on darwin: %q", out.String())
	}
}
