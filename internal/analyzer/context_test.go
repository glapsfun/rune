package analyzer

import (
	"strings"
	"testing"

	"github.com/rune-task-runner/rune/internal/diag"
)

func containsMsg(diags diag.List, substr string) bool {
	for _, d := range diags {
		if strings.Contains(d.Message, substr) {
			return true
		}
	}
	return false
}

func TestContextValid(t *testing.T) {
	diags := analyzeDiags(t, "[context]\n[linux, macos]\nhealth:\n    @echo ok\nbuild:\n    @echo b\n")
	if diags.HasErrors() {
		t.Fatalf("valid [context] task should analyze cleanly: %v", diags)
	}
}

func TestContextDuplicateRejected(t *testing.T) {
	diags := analyzeDiags(t, "[context]\na:\n    @echo a\n[context]\nb:\n    @echo b\n")
	if !diags.HasErrors() || !containsMsg(diags, "duplicate [context]") {
		t.Fatalf("want duplicate [context] error, got: %v", diags)
	}
}

func TestContextConfirmRejected(t *testing.T) {
	diags := analyzeDiags(t, "[context]\n[confirm(\"sure?\")]\na:\n    @echo a\n")
	if !diags.HasErrors() || !containsMsg(diags, "[confirm]") {
		t.Fatalf("want [confirm] incompatibility error, got: %v", diags)
	}
}

func TestContextRequiredParamRejected(t *testing.T) {
	diags := analyzeDiags(t, "[context]\na target:\n    @echo {{target}}\n")
	if !diags.HasErrors() || !containsMsg(diags, "must have a default") {
		t.Fatalf("want defaultless-parameter error, got: %v", diags)
	}
}

func TestContextDefaultedParamAllowed(t *testing.T) {
	diags := analyzeDiags(t, "[context]\na target=\"all\":\n    @echo {{target}}\n")
	if diags.HasErrors() {
		t.Fatalf("defaulted parameter should be allowed: %v", diags)
	}
}

func TestContextAgentExecutorRejected(t *testing.T) {
	diags := analyzeDiags(t, "[context]\na (agent):\n    Summarize the repo.\n")
	if !diags.HasErrors() || !containsMsg(diags, "agent executor") {
		t.Fatalf("want agent-executor rejection, got: %v", diags)
	}
}

// The unattended rules cover the hook's dependency closure: a [confirm] dep
// auto-declines (the hook engine has no stdin) and would doom every run.
func TestContextConfirmDepRejected(t *testing.T) {
	diags := analyzeDiags(t, "[confirm]\nclean:\n    @echo c\n[context]\nhealth: clean\n    @echo ok\n")
	if !diags.HasErrors() || !containsMsg(diags, `dependency "clean"`) {
		t.Fatalf("want [confirm]-dependency rejection, got: %v", diags)
	}
}

// Same for the agent executor anywhere beneath the hook — even transitively.
func TestContextAgentDepRejectedTransitively(t *testing.T) {
	diags := analyzeDiags(t, "summarize (agent):\n    Summarize.\nmid: summarize\n    @echo m\n[context]\nhealth: mid\n    @echo ok\n")
	if !diags.HasErrors() || !containsMsg(diags, `dependency "summarize"`) {
		t.Fatalf("want transitive agent-dependency rejection, got: %v", diags)
	}
}

// Post-hooks are part of the closure too.
func TestContextConfirmPostHookRejected(t *testing.T) {
	diags := analyzeDiags(t, "prep:\n    @echo p\n[confirm]\nnotify:\n    @echo n\n[context]\nhealth: prep && notify\n    @echo ok\n")
	if !diags.HasErrors() || !containsMsg(diags, `dependency "notify"`) {
		t.Fatalf("want [confirm]-post-hook rejection, got: %v", diags)
	}
}

// A plain dependency stays legal.
func TestContextPlainDepAllowed(t *testing.T) {
	diags := analyzeDiags(t, "clean:\n    @echo c\n[context]\nhealth: clean\n    @echo ok\n")
	if diags.HasErrors() {
		t.Fatalf("plain dependency should be allowed: %v", diags)
	}
}
