//go:build runetest

package main

import "os"

// installedVersion lets integration tests (which build the binary with
// `-tags runetest`) inject the reported installed version via RUNE_TEST_VERSION,
// so the minimum_version gate can be exercised end-to-end without release-time
// ldflags. This file is NOT compiled into production binaries, so a released
// `rune` never honors the env var and the installed version stays authoritative.
func installedVersion(version string) string {
	if tv := os.Getenv("RUNE_TEST_VERSION"); tv != "" {
		return tv
	}
	return version
}
