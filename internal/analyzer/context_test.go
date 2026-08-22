package analyzer

import (
	"strings"
	"testing"

	"github.com/rune-task-runner/rune/internal/diag"
	"github.com/rune-task-runner/rune/internal/parser"
)

func analyzeSrc(t *testing.T, src string) diag.List {
	t.Helper()
	f, pdiags := parser.Parse("Runefile", src)
	if pdiags.HasErrors() {
		t.Fatalf("parse: %v", pdiags)
	}
	return Analyze(f)
}

func containsMsg(diags diag.List, substr string) bool {
	for _, d := range diags {
		if strings.Contains(d.Message, substr) {
			return true
		}
	}
	return false
}

func TestContextValid(t *testing.T) {
	diags := analyzeSrc(t, "[context]\n[linux, macos]\nhealth:\n    @echo ok\nbuild:\n    @echo b\n")
	if diags.HasErrors() {
		t.Fatalf("valid [context] task should analyze cleanly: %v", diags)
	}
}

func TestContextDuplicateRejected(t *testing.T) {
	diags := analyzeSrc(t, "[context]\na:\n    @echo a\n[context]\nb:\n    @echo b\n")
	if !diags.HasErrors() || !containsMsg(diags, "duplicate [context]") {
		t.Fatalf("want duplicate [context] error, got: %v", diags)
	}
}

func TestContextConfirmRejected(t *testing.T) {
	diags := analyzeSrc(t, "[context]\n[confirm(\"sure?\")]\na:\n    @echo a\n")
	if !diags.HasErrors() || !containsMsg(diags, "[confirm]") {
		t.Fatalf("want [confirm] incompatibility error, got: %v", diags)
	}
}

func TestContextRequiredParamRejected(t *testing.T) {
	diags := analyzeSrc(t, "[context]\na target:\n    @echo {{target}}\n")
	if !diags.HasErrors() || !containsMsg(diags, "must have a default") {
		t.Fatalf("want defaultless-parameter error, got: %v", diags)
	}
}

func TestContextDefaultedParamAllowed(t *testing.T) {
	diags := analyzeSrc(t, "[context]\na target=\"all\":\n    @echo {{target}}\n")
	if diags.HasErrors() {
		t.Fatalf("defaulted parameter should be allowed: %v", diags)
	}
}

func TestContextAgentExecutorRejected(t *testing.T) {
	diags := analyzeSrc(t, "[context]\na (agent):\n    Summarize the repo.\n")
	if !diags.HasErrors() || !containsMsg(diags, "agent executor") {
		t.Fatalf("want agent-executor rejection, got: %v", diags)
	}
}
