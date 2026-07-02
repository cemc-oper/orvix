package scheduler

import (
	"testing"

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
