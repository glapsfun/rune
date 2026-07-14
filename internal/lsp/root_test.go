package lsp

import "testing"

// TestInitializeSetsWorkspaceRoot verifies the server determines and stores the
// workspace root from the client's initialize params (spec FR-021), so file
// watching can be scoped to the project.
func TestInitializeSetsWorkspaceRoot(t *testing.T) {
	srv := NewServer(nil, nil, Options{})

	// An explicit workspace folder wins over discovery.
	root := srv.detectRoot(InitializeParams{
		WorkspaceFolders: []WorkspaceFolder{{URI: "file:///home/dev/project", Name: "project"}},
	})
	if root != "/home/dev/project" {
		t.Errorf("root from workspaceFolders = %q, want /home/dev/project", root)
	}

	// rootUri is used when no workspaceFolders are given.
	root = srv.detectRoot(InitializeParams{RootURI: "file:///srv/app"})
	if root != "/srv/app" {
		t.Errorf("root from rootUri = %q, want /srv/app", root)
	}

	// With neither, DetectRoot falls back through the FR-021 order (never panics,
	// returns a usable path).
	if got := srv.detectRoot(InitializeParams{}); got == "" {
		t.Error("fallback root should be non-empty")
	}
}
