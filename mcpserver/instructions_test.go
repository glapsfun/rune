package mcpserver

import "testing"

// TestInstructionsDelivered: text passed as Options.Instructions reaches the
// client in the initialize result (spec 021 FR-002).
func TestInstructionsDelivered(t *testing.T) {
	srv := New(sampleEngine(), Options{Instructions: "on branch main; lint clean"})
	cs := connect(t, srv)
	if got := cs.InitializeResult().Instructions; got != "on branch main; lint clean" {
		t.Errorf("instructions = %q, want the injected context", got)
	}
}

// TestNoInstructionsByDefault: without Options.Instructions the initialize
// result carries none (spec 021 FR-009).
func TestNoInstructionsByDefault(t *testing.T) {
	srv := New(sampleEngine(), Options{})
	cs := connect(t, srv)
	if got := cs.InitializeResult().Instructions; got != "" {
		t.Errorf("instructions = %q, want empty", got)
	}
}
