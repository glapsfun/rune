package cli

import "testing"

// TestAdapterExcludesContextTask: the [context] hook is never a tool — its
// output is already delivered as instructions (spec 021 FR-006). Both
// transports (`rune mcp` stdio and `rune serve` HTTP) build this same
// adapter, so this covers stdio and HTTP.
func TestAdapterExcludesContextTask(t *testing.T) {
	a := adapterFor(t, "[context]\nhealth:\n    @echo h\nbuild:\n    @echo b\n")
	names := map[string]bool{}
	for _, ti := range a.Tasks() {
		names[ti.Name] = true
	}
	if names["health"] {
		t.Error("context task must not be exposed as a tool")
	}
	if !names["build"] {
		t.Errorf("ordinary tasks must remain exposed: %v", names)
	}
}
