package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// US2: TaskCandidates feeds dynamic shell completion. It must surface only
// completable tasks (non-private, OS-matching) with their doc summaries, and
// degrade gracefully (nil, no panic) when the Runefile is absent or unparseable.

func TestTaskCandidates_FiltersAndDocs(t *testing.T) {
	dir := t.TempDir()
	content := "# Build the binary.\nbuild:\n    @echo b\n\n[private]\nsecret:\n    @echo s\n"
	if err := os.WriteFile(filepath.Join(dir, "Runefile"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	docs := map[string]string{}
	for _, c := range TaskCandidates(Options{Cwd: dir}) {
		docs[c.Name] = c.Doc
	}

	if _, ok := docs["build"]; !ok {
		t.Fatalf("expected 'build' candidate; got %v", docs)
	}
	if !strings.Contains(docs["build"], "Build the binary") {
		t.Errorf("build doc = %q, want to contain %q", docs["build"], "Build the binary")
	}
	if _, ok := docs["secret"]; ok {
		t.Errorf("[private] task 'secret' must not be a completion candidate")
	}
}

func TestTaskCandidates_GracefulNoRunefile(t *testing.T) {
	if got := TaskCandidates(Options{Cwd: t.TempDir()}); got != nil {
		t.Errorf("want nil candidates with no Runefile, got %v", got)
	}
}

func TestTaskCandidates_GracefulParseError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Runefile"), []byte("x := \"unterminated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := TaskCandidates(Options{Cwd: dir}); got != nil {
		t.Errorf("want nil candidates on parse error, got %v", got)
	}
}
