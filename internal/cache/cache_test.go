package cache

import (
	"os"
	"path/filepath"
	"testing"
)

func specFor(root string) Spec {
	return Spec{
		Key:      "build",
		Root:     root,
		Inputs:   []string{"src/**/*.go", "go.mod"},
		Outputs:  []string{"out/bin"},
		Body:     "go build",
		Vars:     map[string]string{"target": "linux"},
		Executor: "sh",
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCacheMissThenHit(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module x")
	writeFile(t, filepath.Join(root, "src", "a.go"), "package a")
	writeFile(t, filepath.Join(root, "out", "bin"), "binary")

	spec := specFor(root)

	d1, err := Decide(spec)
	if err != nil {
		t.Fatal(err)
	}
	if d1.Skip {
		t.Fatal("first run should not skip (no stored record)")
	}
	if err := Store(spec, d1.Hash, "2026-06-08T00:00:00Z"); err != nil {
		t.Fatal(err)
	}

	d2, err := Decide(spec)
	if err != nil {
		t.Fatal(err)
	}
	if !d2.Skip {
		t.Fatal("second run with unchanged inputs should skip")
	}
}

func TestCacheInputChangeForcesRun(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module x")
	writeFile(t, filepath.Join(root, "src", "a.go"), "package a")
	writeFile(t, filepath.Join(root, "out", "bin"), "binary")
	spec := specFor(root)

	d, _ := Decide(spec)
	_ = Store(spec, d.Hash, "t")

	writeFile(t, filepath.Join(root, "src", "a.go"), "package a // changed")
	d2, _ := Decide(spec)
	if d2.Skip {
		t.Error("changed input must force a re-run")
	}
}

func TestCacheMissingOutputForcesRun(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module x")
	writeFile(t, filepath.Join(root, "out", "bin"), "binary")
	spec := specFor(root)
	d, _ := Decide(spec)
	_ = Store(spec, d.Hash, "t")

	_ = os.Remove(filepath.Join(root, "out", "bin"))
	d2, _ := Decide(spec)
	if d2.Skip {
		t.Error("missing output must force a re-run")
	}
}

func TestCacheVarChangeForcesRun(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module x")
	writeFile(t, filepath.Join(root, "out", "bin"), "binary")
	spec := specFor(root)
	d, _ := Decide(spec)
	_ = Store(spec, d.Hash, "t")

	spec.Vars["target"] = "windows"
	d2, _ := Decide(spec)
	if d2.Skip {
		t.Error("changed variable must force a re-run")
	}
}

func TestCacheCorruptionIsMiss(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module x")
	writeFile(t, filepath.Join(root, "out", "bin"), "binary")
	spec := specFor(root)
	d, _ := Decide(spec)
	_ = Store(spec, d.Hash, "t")
	// Corrupt the record.
	writeFile(t, recordPath(spec), "{not valid json")
	d2, err := Decide(spec)
	if err != nil {
		t.Fatalf("corruption must not error: %v", err)
	}
	if d2.Skip {
		t.Error("corruption must be treated as a miss")
	}
}

func TestClear(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module x")
	writeFile(t, filepath.Join(root, "out", "bin"), "binary")
	spec := specFor(root)
	d, _ := Decide(spec)
	_ = Store(spec, d.Hash, "t")
	if err := Clear(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cacheDir(root)); !os.IsNotExist(err) {
		t.Error("cache dir should be gone after Clear")
	}
}
