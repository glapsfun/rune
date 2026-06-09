package docs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// markdownLinkRe matches [text](target) inline links.
var markdownLinkRe = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

// TestInternalLinksResolve fails on any relative Markdown link that does not
// point at an existing file. External (http/https/mailto) links and pure
// in-page anchors are skipped (offline, deterministic — research D4). Anchor
// (heading) resolution is intentionally deferred to a later tightening.
func TestInternalLinksResolve(t *testing.T) {
	for _, f := range docMarkdownFiles(t) {
		t.Run(relPath(f), func(t *testing.T) {
			src, err := os.ReadFile(f)
			if err != nil {
				t.Fatalf("read %s: %v", relPath(f), err)
			}
			for _, m := range markdownLinkRe.FindAllStringSubmatch(string(src), -1) {
				target := strings.TrimSpace(m[1])
				if skipLink(target) {
					continue
				}
				// Strip any #anchor or ?query suffix.
				if i := strings.IndexAny(target, "#?"); i >= 0 {
					target = target[:i]
				}
				if target == "" {
					continue // pure in-page anchor
				}
				resolved := filepath.Join(filepath.Dir(f), filepath.FromSlash(target))
				if _, err := os.Stat(resolved); err != nil {
					t.Errorf("broken internal link %q in %s → %s not found",
						m[1], relPath(f), relPath(resolved))
				}
			}
		})
	}
}

func skipLink(target string) bool {
	switch {
	case strings.HasPrefix(target, "http://"), strings.HasPrefix(target, "https://"):
		return true
	case strings.HasPrefix(target, "mailto:"):
		return true
	case strings.HasPrefix(target, "#"):
		return true
	default:
		return false
	}
}
