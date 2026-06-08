package version

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	// Version should never be empty.
	require.NotEmpty(t, Version, "Version is empty")
	t.Logf("Version = %q", Version)

	// Verify the computation logic by inspecting build info.
	info, ok := debug.ReadBuildInfo()
	require.True(t, ok, "debug.ReadBuildInfo failed")
	t.Logf("Main.Version = %q", info.Main.Version)
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" || s.Key == "vcs.modified" {
			t.Logf("Setting %s = %q", s.Key, s.Value)
		}
	}
}
