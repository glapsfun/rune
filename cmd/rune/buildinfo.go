package main

import (
	"runtime/debug"
	"strings"
)

// devVersion is the placeholder main.version carries until release ldflags
// stamp a real version; resolveVersion treats it as "not stamped".
const devVersion = "dev"

// resolveVersion returns the version the binary should report. Release
// artifacts stamp main.version via -ldflags and that value is authoritative;
// the fallback below exists so the documented
// `go install github.com/rune-task-runner/rune/cmd/rune@latest` path — which
// cannot receive ldflags — still reports the release it was built from
// instead of the "dev" placeholder (and is therefore no longer waved through
// the minimum_version gate as a development build).
func resolveVersion(version string) string {
	if version != devVersion {
		return version
	}
	bi, _ := debug.ReadBuildInfo()
	return versionFromBuildInfo(version, bi)
}

// versionFromBuildInfo resolves the reported version from the ldflags value
// and the toolchain-embedded build info. The module version is trusted only
// for module-cache builds (`go install module@version`), recognized by the
// absence of vcs.* build settings: Go ≥1.24 stamps a version into checkout
// builds too, and a clean checkout at a tag would otherwise claim that
// release verbatim. Checkout builds therefore keep reporting "dev", which
// also preserves their deliberate minimum_version-gate bypass.
func versionFromBuildInfo(version string, bi *debug.BuildInfo) string {
	if version != devVersion {
		return version
	}
	if bi == nil {
		return version
	}
	mv := bi.Main.Version
	if mv == "" || mv == "(devel)" {
		return version
	}
	for _, s := range bi.Settings {
		if s.Key == "vcs" || strings.HasPrefix(s.Key, "vcs.") {
			return version
		}
	}
	// Module versions carry a leading "v" (v0.4.3); the CLI reports 0.4.3.
	return strings.TrimPrefix(mv, "v")
}
