package submit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
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
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	// Generate should create a .submit file next to the original.
	gotPath, err := Generate(GenerateOptions{
		ScriptPath: scriptPath,
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	wantPath := scriptPath + ".submit"
	if gotPath != wantPath {
		t.Errorf("path: got %q, want %q", gotPath, wantPath)
	}

	info, err := os.Stat(wantPath)
	if err != nil {
		t.Fatalf("stat generated script: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Errorf("generated script is not executable: %o", info.Mode().Perm())
	}

	gotBytes, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read generated script: %v", err)
	}
	got := string(gotBytes)

	// The generated script should preserve the shebang and body,
	// and strip #ORVIX lines.
	if !strings.Contains(got, "#!/bin/bash") {
		t.Error("generated script missing shebang")
	}
	if !strings.Contains(got, "echo hello") {
		t.Error("generated script missing body")
	}
	if strings.Contains(got, "#ORVIX") {
		t.Error("generated script should not contain #ORVIX lines")
	}
}

func TestGenerateCustomOutput(t *testing.T) {
	tmpDir := t.TempDir()

	scriptContent := `#!/bin/bash
#ORVIX scheduler=local
echo hello
`
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	customPath := filepath.Join(tmpDir, "out.sh")
	gotPath, err := Generate(GenerateOptions{
		ScriptPath: scriptPath,
		OutScript:  customPath,
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if gotPath != customPath {
		t.Errorf("path: got %q, want %q", gotPath, customPath)
	}

	if _, err := os.Stat(customPath); err != nil {
		t.Errorf("generated script not found at custom path: %v", err)
	}
}

func TestGenerateSchedulerOverride(t *testing.T) {
	tmpDir := t.TempDir()

	// Script has no scheduler directive.
	scriptContent := `#!/bin/bash
#ORVIX job-name=override-test
echo hello
`
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	// Default scheduler would be "local". Override to "slurm".
	_, err := Generate(GenerateOptions{
		ScriptPath: scriptPath,
		Scheduler:  "slurm",
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	gotBytes, err := os.ReadFile(scriptPath + ".submit")
	if err != nil {
		t.Fatalf("read generated script: %v", err)
	}
	got := string(gotBytes)

	// SLURM preamble should contain #SBATCH lines.
	if !strings.Contains(got, "#SBATCH") {
		t.Error("expected SLURM #SBATCH lines in generated script")
	}
}
