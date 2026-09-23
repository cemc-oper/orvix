package scheduler

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cemc-oper/orvix/internal/directive"
)

func mustParseSlurm(t *testing.T, src string) *directive.Set {
	t.Helper()
	set, err := directive.ParseWithOverride([]byte(src), "slurm")
	require.NoError(t, err, "parse error")
	return set
}

func TestSlurmPreambleSimpleMappings(t *testing.T) {
	src := `#!/bin/bash
#ORVIX job-name=myjob
#ORVIX nodes=24
#ORVIX ntasks-per-node=64
#ORVIX time=00:30:00
#ORVIX queue=normal
#ORVIX project=op_mcv
#ORVIX application=mcv
#ORVIX memory=25G
`
	lines, err := (&SLURM{}).PreambleFor(mustParseSlurm(t, src))
	require.NoError(t, err)
	want := []string{
		"#SBATCH --job-name=myjob",
		"#SBATCH --nodes=24",
		"#SBATCH --ntasks-per-node=64",
		"#SBATCH --time=00:30:00",
		"#SBATCH --partition=normal",
		"#SBATCH --wckey=op_mcv",
		"#SBATCH --comment=mcv",
		"#SBATCH --mem=25G",
	}
	assert.Equal(t, want, lines)
}

func TestSlurmRequeueFalseEmitsNoRequeue(t *testing.T) {
	lines, err := (&SLURM{}).PreambleFor(mustParseSlurm(t, "#ORVIX requeue=false\n"))
	require.NoError(t, err)
	assert.Equal(t, []string{"#SBATCH --no-requeue"}, lines)
}

func TestSlurmRequeueTrueEmitsNothing(t *testing.T) {
	lines, err := (&SLURM{}).PreambleFor(mustParseSlurm(t, "#ORVIX requeue=true\n"))
	require.NoError(t, err)
	assert.Empty(t, lines)
}

func TestSlurmRequeueAbsentEmitsNothing(t *testing.T) {
	lines, err := (&SLURM{}).PreambleFor(mustParseSlurm(t, "#ORVIX nodes=2\n"))
	require.NoError(t, err)
	assert.Equal(t, []string{"#SBATCH --nodes=2"}, lines)
}

// --- Kill tests --------------------------------------------------------------

// stubBin installs fake scheduler CLI stubs in a temp dir prepended to PATH.
// Every stub first logs its invocation as "<name> <args>" to $STUB_LOG.
type stubBin struct {
	dir string
	log string
}

func newStubBin(t *testing.T) *stubBin {
	t.Helper()
	dir := t.TempDir()
	s := &stubBin{dir: dir, log: filepath.Join(dir, "calls.log")}
	t.Setenv("STUB_LOG", s.log)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return s
}

// add installs a stub executable; body is bash executed after the invocation
// log line (e.g. "echo RUNNING", "exit 1").
func (s *stubBin) add(t *testing.T, name, body string) {
	t.Helper()
	script := "#!/bin/bash\necho \"" + name + ` $*"` + " >> \"$STUB_LOG\"\n" + body + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(s.dir, name), []byte(script), 0o755))
}

// calls returns the logged invocations, one per line.
func (s *stubBin) calls(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(s.log)
	if os.IsNotExist(err) {
		return nil
	}
	require.NoError(t, err)
	return strings.Split(strings.TrimRight(string(data), "\n"), "\n")
}

func (s *stubBin) scancelCalls(t *testing.T) []string {
	t.Helper()
	var out []string
	for _, c := range s.calls(t) {
		if strings.HasPrefix(c, "scancel") {
			out = append(out, c)
		}
	}
	return out
}

func TestSlurmKillDefaultJobEndsDuringGrace(t *testing.T) {
	stubs := newStubBin(t)
	t.Setenv("ORVIX_SLURM_KILL_GRACE", "5")
	stubs.add(t, "scancel", "exit 0")
	stubs.add(t, "squeue", "exit 1") // job already left the queue
	stubs.add(t, "sacct", "echo CANCELLED")

	require.NoError(t, (&SLURM{}).Kill("123", nil))

	calls := stubs.calls(t)
	require.NotEmpty(t, calls)
	assert.Equal(t, "scancel --full --signal=TERM 123", calls[0])
	assert.Empty(t, stubs.scancelCalls(t)[1:], "fallback scancel must not run once the job is terminal")
}

func TestSlurmKillDefaultEscalatesAfterGrace(t *testing.T) {
	stubs := newStubBin(t)
	t.Setenv("ORVIX_SLURM_KILL_GRACE", "2")
	stubs.add(t, "scancel", "exit 0")
	stubs.add(t, "squeue", "echo RUNNING")

	require.NoError(t, (&SLURM{}).Kill("123", nil))

	calls := stubs.scancelCalls(t)
	require.Len(t, calls, 2)
	assert.Equal(t, "scancel --full --signal=TERM 123", calls[0])
	assert.Equal(t, "scancel 123", calls[1])
}

func TestSlurmKillDefaultZeroGraceEscalatesImmediately(t *testing.T) {
	stubs := newStubBin(t)
	t.Setenv("ORVIX_SLURM_KILL_GRACE", "0")
	stubs.add(t, "scancel", "exit 0")
	stubs.add(t, "squeue", "echo RUNNING")

	require.NoError(t, (&SLURM{}).Kill("123", nil))

	calls := stubs.scancelCalls(t)
	require.Len(t, calls, 2)
	assert.Equal(t, "scancel --full --signal=TERM 123", calls[0])
	assert.Equal(t, "scancel 123", calls[1])
}

func TestSlurmKillExplicitSignalAddsFull(t *testing.T) {
	stubs := newStubBin(t)
	stubs.add(t, "scancel", "exit 0")

	require.NoError(t, (&SLURM{}).Kill("123", syscall.SIGKILL))
	assert.Equal(t, []string{"scancel --full --signal=9 123"}, stubs.calls(t))
}

func TestSlurmKillAlreadyFinishedIsSuccess(t *testing.T) {
	stubs := newStubBin(t)
	t.Setenv("ORVIX_SLURM_KILL_GRACE", "0")
	stubs.add(t, "scancel",
		`echo "scancel: error: Kill job error on job id 123: Invalid job id specified" >&2; exit 1`)
	stubs.add(t, "squeue", "exit 1")
	stubs.add(t, "sacct", "echo COMPLETED")

	require.NoError(t, (&SLURM{}).Kill("123", nil))
}

func TestSlurmKillErrorPropagatesWhileJobAlive(t *testing.T) {
	stubs := newStubBin(t)
	t.Setenv("ORVIX_SLURM_KILL_GRACE", "0")
	stubs.add(t, "scancel", `echo "scancel: error: some unexpected failure" >&2; exit 1`)
	stubs.add(t, "squeue", "echo RUNNING")

	assert.Error(t, (&SLURM{}).Kill("123", nil))
}

func TestSlurmKillGraceEnv(t *testing.T) {
	t.Setenv("ORVIX_SLURM_KILL_GRACE", "7")
	assert.Equal(t, 7*time.Second, slurmKillGrace())
	t.Setenv("ORVIX_SLURM_KILL_GRACE", "junk")
	assert.Equal(t, 30*time.Second, slurmKillGrace())
	t.Setenv("ORVIX_SLURM_KILL_GRACE", "-3")
	assert.Equal(t, 30*time.Second, slurmKillGrace())
}
