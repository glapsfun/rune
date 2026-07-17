package analysis

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTemp writes content to dir/name and returns the full path.
func writeTemp(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}
