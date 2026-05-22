package version

import (
	"runtime/debug"
	"testing"
)

func TestVersion(t *testing.T) {
	// Version should never be empty.
	if Version == "" {
		t.Fatal("Version is empty")
	}
	t.Logf("Version = %q", Version)

	// Verify the computation logic by inspecting build info.
	info, ok := debug.ReadBuildInfo()
	if !ok {
		t.Fatal("debug.ReadBuildInfo failed")
	}
	t.Logf("Main.Version = %q", info.Main.Version)
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" || s.Key == "vcs.modified" {
			t.Logf("Setting %s = %q", s.Key, s.Value)
		}
	}
}
