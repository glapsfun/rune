package lsp

import (
	"path/filepath"
	"strings"
)

// URI <-> path conversion at the protocol boundary. The analysis layer works in
// filesystem paths; LSP clients speak file:// URIs. These converters are
// intentionally hand-written (no net/url) so the package keeps zero networking
// dependencies (the safety guard forbids net/*).

// uriToPath converts a file:// URI to a filesystem path. A plain path is
// returned unchanged, so the server tolerates non-conformant clients.
func uriToPath(uri string) string {
	if !strings.HasPrefix(uri, "file://") {
		return filepath.FromSlash(percentDecode(uri))
	}
	p := strings.TrimPrefix(uri, "file://")
	// An authority component ("file://host/path") is uncommon for local files;
	// drop an empty authority ("file:///path" -> "/path").
	if strings.HasPrefix(p, "/") {
		// keep leading slash
	} else if i := strings.IndexByte(p, '/'); i >= 0 {
		p = p[i:] // strip host
	}
	p = percentDecode(p)
	// Windows drive paths arrive as /C:/... — strip the leading slash.
	if len(p) >= 3 && p[0] == '/' && isDriveLetter(p[1]) && p[2] == ':' {
		p = p[1:]
	}
	return filepath.FromSlash(p)
}

// pathToURI converts a filesystem path to a file:// URI.
func pathToURI(path string) string {
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p // Windows "C:/..." -> "/C:/..."
	}
	return "file://" + percentEncodePath(p)
}

func isDriveLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func percentDecode(s string) string {
	if !strings.ContainsRune(s, '%') {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) {
			if h := unhex(s[i+1]); h >= 0 {
				if l := unhex(s[i+2]); l >= 0 {
					b.WriteByte(byte(h<<4 | l))
					i += 2
					continue
				}
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// percentEncodePath encodes a path for a file URI, leaving path separators and
// common safe characters intact.
func percentEncodePath(p string) string {
	const safe = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_.~/:"
	var b strings.Builder
	for i := 0; i < len(p); i++ {
		c := p[i]
		if strings.IndexByte(safe, c) >= 0 {
			b.WriteByte(c)
		} else {
			const hexdigits = "0123456789ABCDEF"
			b.WriteByte('%')
			b.WriteByte(hexdigits[c>>4])
			b.WriteByte(hexdigits[c&0xF])
		}
	}
	return b.String()
}

func unhex(b byte) int {
	switch {
	case b >= '0' && b <= '9':
		return int(b - '0')
	case b >= 'a' && b <= 'f':
		return int(b-'a') + 10
	case b >= 'A' && b <= 'F':
		return int(b-'A') + 10
	default:
		return -1
	}
}
