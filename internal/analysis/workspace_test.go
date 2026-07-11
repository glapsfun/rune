package analysis

import "testing"

func TestDetectRootOrder(t *testing.T) {
	// Explicit workspace folder always wins.
	if got := DetectRoot("/explicit", "/a/b/Runefile", func(string) bool { return true }); got != "/explicit" {
		t.Errorf("explicit: got %q", got)
	}

	// Nearest Runefile beats .git.
	present := map[string]bool{"/proj/Runefile": true, "/proj/sub/.git": true}
	exists := func(p string) bool { return present[p] }
	if got := DetectRoot("", "/proj/sub/deep/Runefile", exists); got != "/proj" {
		t.Errorf("nearest Runefile: got %q, want /proj", got)
	}

	// No Runefile: fall back to nearest .git.
	present = map[string]bool{"/repo/.git": true}
	if got := DetectRoot("", "/repo/pkg/Runefile", exists); got != "/repo" {
		t.Errorf("nearest .git: got %q, want /repo", got)
	}

	// Neither: the document's own directory.
	present = map[string]bool{}
	if got := DetectRoot("", "/lonely/Runefile", exists); got != "/lonely" {
		t.Errorf("fallback: got %q, want /lonely", got)
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
