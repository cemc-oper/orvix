package jobspec

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cemc-oper/orvix/internal/submit"
)

// conformanceSchedulers are the backends every vector is exercised against.
// Keep in sync with the directories under conformance/golden/.
var conformanceSchedulers = []string{"slurm", "donau", "local"}

var update = flag.Bool("update", false,
	"regenerate conformance golden files from the current build instead of diffing")

// TestConformanceGolden is the native contract gate: every script in
// conformance/vectors is run through the real generate pipeline
// (parse directives -> resolve scheduler -> render) once per scheduler
// backend, and the output is diffed against conformance/golden/<scheduler>/.
//
// Any orvix change that alters generated output fails here; if the change is
// intentional, regenerate the baselines with `make regen-golden` (i.e.
// `go test ./spec/jobspec/ -run TestConformanceGolden -update`) and review
// the diff in the same commit.
func TestConformanceGolden(t *testing.T) {
	vectors, err := filepath.Glob("conformance/vectors/*.sh")
	require.NoError(t, err)
	require.NotEmpty(t, vectors, "no conformance vectors found")

	for _, vector := range vectors {
		name := strings.TrimSuffix(filepath.Base(vector), ".sh")
		src, err := os.ReadFile(vector)
		require.NoError(t, err)

		for _, schedName := range conformanceSchedulers {
			t.Run(name+"/"+schedName, func(t *testing.T) {
				tmp := t.TempDir()
				scriptPath := filepath.Join(tmp, name+".sh")
				outPath := filepath.Join(tmp, "out.submit")
				require.NoError(t, os.WriteFile(scriptPath, src, 0o644))

				_, err := submit.Generate(submit.GenerateOptions{
					ScriptPath: scriptPath,
					Scheduler:  schedName,
					OutScript:  outPath,
					Out:        io.Discard,
				})
				require.NoError(t, err)

				produced, err := os.ReadFile(outPath)
				require.NoError(t, err)

				goldenPath := filepath.Join("conformance", "golden", schedName, name+".submit")
				if *update {
					require.NoError(t, os.WriteFile(goldenPath, produced, 0o644))
					t.Logf("updated %s", goldenPath)
					return
				}

				golden, err := os.ReadFile(goldenPath)
				require.NoError(t, err)
				assert.Equal(t, string(golden), string(produced),
					"generated output drifted from %s; if intentional, run `make regen-golden`",
					goldenPath)
			})
		}
	}
}
