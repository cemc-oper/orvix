// Package version records the orvix CLI version.
//
// Release builds (GoReleaser) inject the tag at link time via:
//
//	-X github.com/cemc-oper/orvix/internal/version.Version=vX.Y.Z
//
// Otherwise the version is derived at runtime from the Go build info
// injected by the toolchain (Go 1.18+). When the binary is installed as a
// versioned module (go install ...@vX.Y.Z), that tag is used. For local
// development builds, the git commit hash is used instead. Falls back to
// "dev" when no version info is available.
package version

import "runtime/debug"

// Version is the orvix release version.
//
// Release builds set it at link time (see the package doc), in which case it
// is used as-is. Otherwise it is filled in by computeVersion at startup.
var Version = ""

func init() {
	if Version == "" {
		Version = computeVersion()
	}
}

func computeVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}

	var revision, modified string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value
		}
	}

	// Prefer the module version when available (e.g. from go install or tagged builds).
	// Go already appends +dirty to the module version when the working tree is dirty,
	// so we use it as-is.
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	// Fall back to the VCS revision for local development builds.
	if revision != "" {
		v := revision
		if len(v) > 12 {
			v = v[:12]
		}
		if modified == "true" {
			v += "+dirty"
		}
		return v
	}

	return "dev"
}
