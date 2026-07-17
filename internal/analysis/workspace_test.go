package analysis

import (
	"path/filepath"
	"testing"
)

func TestDetectRootOrder(t *testing.T) {
	// DetectRoot works in OS-native paths (filepath), so build the fixture paths
	// and expectations the same way to keep the test portable to Windows.
	fs := filepath.FromSlash

	// Explicit workspace folder always wins (returned verbatim).
	if got := DetectRoot("/explicit", fs("/a/b/Runefile"), func(string) bool { return true }); got != "/explicit" {
		t.Errorf("explicit: got %q", got)
	}

	// Nearest Runefile beats .git.
	present := map[string]bool{fs("/proj/Runefile"): true, fs("/proj/sub/.git"): true}
	exists := func(p string) bool { return present[p] }
	if got := DetectRoot("", fs("/proj/sub/deep/Runefile"), exists); got != fs("/proj") {
		t.Errorf("nearest Runefile: got %q, want %q", got, fs("/proj"))
	}

	// No Runefile: fall back to nearest .git.
	present = map[string]bool{fs("/repo/.git"): true}
	if got := DetectRoot("", fs("/repo/pkg/Runefile"), exists); got != fs("/repo") {
		t.Errorf("nearest .git: got %q, want %q", got, fs("/repo"))
	}

	// Neither: the document's own directory.
	present = map[string]bool{}
	if got := DetectRoot("", fs("/lonely/Runefile"), exists); got != fs("/lonely") {
		t.Errorf("fallback: got %q, want %q", got, fs("/lonely"))
	}
}

func TestTransitiveImporters(t *testing.T) {
	g := NewImportGraph()
	g.AddEdge("root", "mid")
	g.AddEdge("mid", "leaf")
	g.AddEdge("other", "leaf")

	imp := g.TransitiveImporters("leaf")
	set := map[string]bool{}
	for _, u := range imp {
		set[u] = true
	}
	for _, want := range []string{"mid", "root", "other"} {
		if !set[want] {
			t.Errorf("TransitiveImporters(leaf) missing %q (got %v)", want, imp)
		}
	}
}
