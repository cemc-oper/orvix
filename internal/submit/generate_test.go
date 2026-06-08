package submit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveScriptPath(t *testing.T) {
	cases := []struct {
		name      string
		orig      string
		outScript string
		want      string
	}{
		{
			name:      "default extension",
			orig:      "/tmp/script.sh",
			outScript: "",
			want:      "/tmp/script.sh.submit",
		},
		{
			name:      "no extension",
			orig:      "/tmp/runme",
			outScript: "",
			want:      "/tmp/runme.submit",
		},
		{
			name:      "stem contains dot",
			orig:      "/tmp/a.b.sh",
			outScript: "",
			want:      "/tmp/a.b.sh.submit",
		},
		{
			name:      "custom output",
			orig:      "/tmp/script.sh",
			outScript: "/custom/submit.sh",
			want:      "/custom/submit.sh",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := deriveScriptPath(tc.orig, tc.outScript)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGenerateLocal(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a script with #ORVIX directives.
	scriptContent := `#!/bin/bash
#ORVIX scheduler=local
#ORVIX job-name=test-generate
echo hello
`
	scriptPath := filepath.Join(tmpDir, "test.sh")
	require.NoError(t, os.WriteFile(scriptPath, []byte(scriptContent), 0o644))

	// Generate should create a .submit file next to the original.
	gotPath, err := Generate(GenerateOptions{
		ScriptPath: scriptPath,
	})
	require.NoError(t, err)

	wantPath := scriptPath + ".submit"
	assert.Equal(t, wantPath, gotPath)

	info, err := os.Stat(wantPath)
	require.NoError(t, err)
	assert.NotZero(t, info.Mode().Perm()&0o111, "generated script should be executable")

	gotBytes, err := os.ReadFile(wantPath)
	require.NoError(t, err)
	got := string(gotBytes)

	// The generated script should preserve the shebang and body,
	// and strip #ORVIX lines.
	assert.Contains(t, got, "#!/bin/bash")
	assert.Contains(t, got, "echo hello")
	assert.NotContains(t, got, "#ORVIX")
}

func TestGenerateCustomOutput(t *testing.T) {
	tmpDir := t.TempDir()

	scriptContent := `#!/bin/bash
#ORVIX scheduler=local
echo hello
`
	scriptPath := filepath.Join(tmpDir, "test.sh")
	require.NoError(t, os.WriteFile(scriptPath, []byte(scriptContent), 0o644))

	customPath := filepath.Join(tmpDir, "out.sh")
	gotPath, err := Generate(GenerateOptions{
		ScriptPath: scriptPath,
		OutScript:  customPath,
	})
	require.NoError(t, err)
	assert.Equal(t, customPath, gotPath)

	_, err = os.Stat(customPath)
	require.NoError(t, err, "generated script not found at custom path")
}

func TestGenerateSchedulerOverride(t *testing.T) {
	tmpDir := t.TempDir()

	// Script has no scheduler directive.
	scriptContent := `#!/bin/bash
#ORVIX job-name=override-test
echo hello
`
	scriptPath := filepath.Join(tmpDir, "test.sh")
	require.NoError(t, os.WriteFile(scriptPath, []byte(scriptContent), 0o644))

	// Default scheduler would be "local". Override to "slurm".
	_, err := Generate(GenerateOptions{
		ScriptPath: scriptPath,
		Scheduler:  "slurm",
	})
	require.NoError(t, err)

	gotBytes, err := os.ReadFile(scriptPath + ".submit")
	require.NoError(t, err)
	got := string(gotBytes)

	// SLURM preamble should contain #SBATCH lines.
	assert.Contains(t, got, "#SBATCH")
}
