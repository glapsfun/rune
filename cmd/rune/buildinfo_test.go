package main

import (
	"runtime/debug"
	"testing"

	"github.com/rune-task-runner/rune/internal/config"
	"github.com/rune-task-runner/rune/internal/semver"
)

// moduleBuildInfo fakes the build info of a binary built from the module
// cache (`go install module@version`): a main-module version and no vcs.*
// settings, since module zips carry no repository metadata.
func moduleBuildInfo(mainVersion string) *debug.BuildInfo {
	return &debug.BuildInfo{
		Main:     debug.Module{Path: "github.com/rune-task-runner/rune", Version: mainVersion},
		Settings: []debug.BuildSetting{{Key: "-buildmode", Value: "exe"}},
	}
}

// checkoutBuildInfo fakes the build info of a binary built from a source
// checkout: Go ≥1.24 stamps a main-module version AND vcs.* settings.
func checkoutBuildInfo(mainVersion string) *debug.BuildInfo {
	bi := moduleBuildInfo(mainVersion)
	bi.Settings = append(bi.Settings,
		debug.BuildSetting{Key: "vcs", Value: "git"},
		debug.BuildSetting{Key: "vcs.revision", Value: "432dbb8de56bb6ced36eb6ddbb3a6a64669275af"},
		debug.BuildSetting{Key: "vcs.time", Value: "2026-08-12T09:25:50Z"},
		debug.BuildSetting{Key: "vcs.modified", Value: "false"},
	)
	return bi
}

func TestVersionFromBuildInfo(t *testing.T) {
	pseudo := "v0.4.4-0.20260812092550-432dbb8de56b"
	cases := []struct {
		name    string
		version string // ldflags-stamped value (release builds stamp a real one)
		bi      *debug.BuildInfo
		want    string
	}{
		{
			// Release artifacts: the stamped version is authoritative and
			// build info must be ignored entirely (contract C3).
			name:    "ldflags version wins over build info",
			version: "9.9.9",
			bi:      moduleBuildInfo("v0.1.0"),
			want:    "9.9.9",
		},
		{
			// The reported bug: `go install mod@v0.4.3` must report 0.4.3.
			name:    "module cache tagged release",
			version: devVersion,
			bi:      moduleBuildInfo("v0.4.3"),
			want:    "0.4.3",
		},
		{
			// `go install mod@<untagged commit>`: report the pseudo-version,
			// visibly not a clean release (contract C5).
			name:    "module cache pseudo-version",
			version: devVersion,
			bi:      moduleBuildInfo(pseudo),
			want:    "0.4.4-0.20260812092550-432dbb8de56b",
		},
		{
			name:    "no build info stays dev",
			version: devVersion,
			bi:      nil,
			want:    devVersion,
		},
		{
			name:    "empty module version stays dev",
			version: devVersion,
			bi:      moduleBuildInfo(""),
			want:    devVersion,
		},
		{
			name:    "(devel) stays dev",
			version: devVersion,
			bi:      moduleBuildInfo("(devel)"),
			want:    devVersion,
		},
		{
			// The dangerous case (research D3): a clean checkout at a tag is
			// stamped with the release version verbatim; the vcs.* settings
			// are what mark it as a source build, not a release.
			name:    "checkout at clean tag stays dev",
			version: devVersion,
			bi:      checkoutBuildInfo("v0.4.3"),
			want:    devVersion,
		},
		{
			name:    "dirty checkout pseudo-version stays dev",
			version: devVersion,
			bi:      checkoutBuildInfo("v0.4.4-0.20260812092550-432dbb8de56b+dirty"),
			want:    devVersion,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := versionFromBuildInfo(c.version, c.bi); got != c.want {
				t.Errorf("versionFromBuildInfo(%q, …) = %q, want %q", c.version, got, c.want)
			}
		})
	}
}

// TestResolvedVersionsSatisfyGate ties the resolved strings to the
// minimum_version gate: go-install binaries must be compared as real
// releases (US2), while "dev" keeps its deliberate bypass (US3 / C8).
func TestResolvedVersionsSatisfyGate(t *testing.T) {
	reqVersion, err := semver.Parse("0.4.0")
	if err != nil {
		t.Fatal(err)
	}
	req := config.MinimumRequirement{Present: true, Raw: "0.4.0", Version: reqVersion}

	cases := []struct {
		name     string
		bi       *debug.BuildInfo
		wantOK   bool
		wantDev  bool
		resolved string // sanity-check of the input under test
	}{
		{
			name:     "module release newer than required passes",
			bi:       moduleBuildInfo("v0.4.3"),
			wantOK:   true,
			wantDev:  false,
			resolved: "0.4.3",
		},
		{
			name:     "module release older than required is blocked",
			bi:       moduleBuildInfo("v0.3.0"),
			wantOK:   false,
			wantDev:  false,
			resolved: "0.3.0",
		},
		{
			name:     "module pseudo-version newer than required passes",
			bi:       moduleBuildInfo("v0.4.4-0.20260812092550-432dbb8de56b"),
			wantOK:   true,
			wantDev:  false,
			resolved: "0.4.4-0.20260812092550-432dbb8de56b",
		},
		{
			name:     "checkout build stays dev and bypasses the gate",
			bi:       checkoutBuildInfo("v9.9.9"),
			wantOK:   true,
			wantDev:  true,
			resolved: devVersion,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := versionFromBuildInfo(devVersion, c.bi)
			if v != c.resolved {
				t.Fatalf("resolved version = %q, want %q", v, c.resolved)
			}
			ok, dev := req.Satisfied(v)
			if ok != c.wantOK || dev != c.wantDev {
				t.Errorf("Satisfied(%q) = (ok=%v, dev=%v), want (ok=%v, dev=%v)", v, ok, dev, c.wantOK, c.wantDev)
			}
		})
	}
}

// TestPseudoVersionBelowPinnedTagIsBlocked locks in the decided edge
// (research D4): a pseudo-version is a semver prerelease of the NEXT tag, so
// a Runefile pinning exactly that tag blocks a `go install @<untagged-commit>`
// binary even if the commit already contains the tag's features. Such builds
// previously slipped through as "dev"; blocking them is intentional.
func TestPseudoVersionBelowPinnedTagIsBlocked(t *testing.T) {
	reqVersion, err := semver.Parse("0.4.4")
	if err != nil {
		t.Fatal(err)
	}
	req := config.MinimumRequirement{Present: true, Raw: "0.4.4", Version: reqVersion}

	v := versionFromBuildInfo(devVersion, moduleBuildInfo("v0.4.4-0.20260812092550-432dbb8de56b"))
	ok, dev := req.Satisfied(v)
	if ok || dev {
		t.Errorf("Satisfied(%q) = (ok=%v, dev=%v), want (ok=false, dev=false)", v, ok, dev)
	}
}
