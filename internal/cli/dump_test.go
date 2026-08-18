package cli

import (
	"testing"

	"github.com/rune-task-runner/rune/internal/parser"
)

// Spec 020 US4: the JSON dump carries a computed `available` verdict per
// task, evaluated against the target OS, while ALL tasks stay listed
// (available:false is data, not a filter — mirrors the private field) and
// the raw OS attribute names remain in attributes.
func TestDumpAvailabilityVerdict(t *testing.T) {
	src := "" +
		"everywhere:\n    @echo e\n" +
		"[windows]\nwin-only:\n    @echo w\n" +
		"[private]\n[linux]\nsecret:\n    @echo s\n"
	f, diags := parser.Parse("Runefile", src)
	if diags.HasErrors() {
		t.Fatalf("parse: %v", diags)
	}

	byName := func(goos string) map[string]taskDTO {
		out := map[string]taskDTO{}
		for _, td := range toDTO(f, goos).Tasks {
			out[td.Name] = td
		}
		return out
	}

	linux := byName("linux")
	if len(linux) != 3 {
		t.Fatalf("dump dropped tasks: %v", linux)
	}
	if !linux["everywhere"].Available {
		t.Error("unrestricted task must dump available:true")
	}
	if linux["win-only"].Available {
		t.Error("[windows] task must dump available:false on linux")
	}
	if !linux["secret"].Available || !linux["secret"].Private {
		t.Errorf("private [linux] task on linux: want available:true private:true, got %+v", linux["secret"])
	}
	if got := linux["win-only"].Attributes; len(got) != 1 || got[0] != "windows" {
		t.Errorf("raw OS attributes must stay in attributes: %v", got)
	}

	windows := byName("windows")
	if !windows["win-only"].Available {
		t.Error("[windows] task must dump available:true on windows")
	}
	if windows["secret"].Available {
		t.Error("[linux] task must dump available:false on windows")
	}
}
