package directive

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBasic(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX nodes=2
#ORVIX time=01:00:00

echo "hello"
`)
	set, err := Parse(src)
	require.NoError(t, err)
	assert.Equal(t, "slurm", set.Scheduler())
	assert.Equal(t, "2", set.Get("nodes"))
	assert.Equal(t, "01:00:00", set.Get("time"))
}

func TestParseBareKey(t *testing.T) {
	src := []byte(`#ORVIX exclusive
#ORVIX nodes=2
`)
	set, err := Parse(src)
	require.NoError(t, err)
	assert.True(t, set.Has("exclusive"))
	assert.Equal(t, "", set.Get("exclusive"))
	assert.Equal(t, "2", set.Get("nodes"))
}

func TestParseQuotedValue(t *testing.T) {
	src := []byte(`#ORVIX application="hello world"
#ORVIX queue='a b c'
`)
	set, err := Parse(src)
	require.NoError(t, err)
	assert.Equal(t, "hello world", set.Get("application"))
	assert.Equal(t, "a b c", set.Get("queue"))
}

func TestParseStopsAtFirstNonCommentLine(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX nodes=2

echo before
#ORVIX time=99
`)
	set, err := Parse(src)
	require.NoError(t, err)
	assert.Equal(t, "2", set.Get("nodes"))
	assert.False(t, set.Has("time"), "time should not be parsed (after non-comment line)")
}

func TestParseDefaultSchedulerIsLocal(t *testing.T) {
	set, err := Parse([]byte("#!/bin/bash\necho hi\n"))
	require.NoError(t, err)
	assert.Equal(t, "local", set.Scheduler())
}

func TestParseSkipsRegularComments(t *testing.T) {
	src := []byte(`#!/bin/bash
# this is a normal comment
#ORVIX nodes=2
# orvix lowercase=ignored
#ORVIXNOSPACE=alsoignored
`)
	set, err := Parse(src)
	require.NoError(t, err)
	assert.Equal(t, "2", set.Get("nodes"))
	assert.False(t, set.Has("lowercase"), "lowercase orvix should not be matched")
	assert.False(t, set.Has("NOSPACE"), "#ORVIXNOSPACE should not be matched (no whitespace after marker)")
}

func TestParseRejectsBarePlusValue(t *testing.T) {
	src := []byte(`#ORVIX nodes 2
`)
	_, err := Parse(src)
	require.Error(t, err, "expected error for `key value` (no =)")
}

func TestParseRejectsEmptyKey(t *testing.T) {
	src := []byte(`#ORVIX =foo
`)
	_, err := Parse(src)
	require.Error(t, err, "expected error for empty key")
}

func TestParseConditionalKeepsMatching(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX scheduler=donau
#ORVIX nodes=2
#ORVIX [scheduler=donau] nodelist=rp_cme001-418
#ORVIX [scheduler=slurm] queue=normal
`)
	set, err := Parse(src)
	require.NoError(t, err)
	assert.True(t, set.Has("nodes"), "nodes should be present (unconditional)")
	assert.True(t, set.Has("nodelist"), "nodelist should be present ([scheduler=donau] matches)")
	assert.False(t, set.Has("queue"), "queue should be excluded ([scheduler=slurm] does not match scheduler=donau)")
	assert.Equal(t, "rp_cme001-418", set.Get("nodelist"))
}

func TestParseConditionalDropsNonMatching(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX nodes=2
#ORVIX [scheduler=donau] nodelist=rp_cme001-418
#ORVIX [scheduler=slurm] queue=normal
`)
	set, err := Parse(src)
	require.NoError(t, err)
	assert.True(t, set.Has("nodes"), "nodes should be present (unconditional)")
	assert.False(t, set.Has("nodelist"), "nodelist should be excluded ([scheduler=donau] does not match scheduler=slurm)")
	assert.True(t, set.Has("queue"), "queue should be present ([scheduler=slurm] matches scheduler=slurm)")
}

func TestParseConditionalBareKey(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX [scheduler=slurm] exclusive
#ORVIX [scheduler=donau] cosched
`)
	set, err := Parse(src)
	require.NoError(t, err)
	assert.True(t, set.Has("exclusive"), "exclusive should be present ([scheduler=slurm] matches)")
	assert.False(t, set.Has("cosched"), "cosched should be excluded ([scheduler=donau] does not match)")
}

func TestParseConditionalNoSpace(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX scheduler=donau
#ORVIX [scheduler=donau]nodelist=rp_cme001-418
`)
	set, err := Parse(src)
	require.NoError(t, err)
	assert.True(t, set.Has("nodelist"), "nodelist should be present (no space after ] is ok)")
}

func TestParseConditionalEmptyBrackets(t *testing.T) {
	// Empty [] is not a valid condition (no = inside), treated as part of the key.
	src := []byte(`#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX [] nodes=4
`)
	_, err := Parse(src)
	require.Error(t, err, "expected error for [] (no key=value inside brackets)")
}

func TestParseConditionalUnclosedBracket(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX [scheduler=donau nodelist=rp_cme001-418
`)
	// Unclosed [ is treated as part of the key, so parseDirective will error
	_, err := Parse(src)
	require.Error(t, err, "expected error for unclosed [")
}

func TestParseConditionalMixedWithUnconditional(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX scheduler=donau
#ORVIX time=01:00:00
#ORVIX [scheduler=donau] time=120
`)
	set, err := Parse(src)
	require.NoError(t, err)
	// The conditional [scheduler=donau] time=120 should override the unconditional time=01:00:00
	assert.Equal(t, "120", set.Get("time"), "conditional should override unconditional")
}

func TestParseWithOverride(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX nodes=2
#ORVIX [scheduler=donau] nodelist=rp_cme001-418
#ORVIX [scheduler=slurm] queue=normal
`)
	// Override to donau: condition filtering uses "donau", and Set.Scheduler() returns "donau"
	set, err := ParseWithOverride(src, "donau")
	require.NoError(t, err)
	assert.Equal(t, "donau", set.Scheduler())
	assert.False(t, set.Has("queue"), "queue should be excluded (condition matches slurm, not donau)")
	assert.True(t, set.Has("nodelist"), "nodelist should be present (condition matches donau)")
}

func TestParseWithOverrideDefaultLocal(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX nodes=2
`)
	set, err := ParseWithOverride(src, "")
	require.NoError(t, err)
	assert.Equal(t, "local", set.Scheduler())
}

func TestParseWithOverrideSetsScheduler(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX nodes=2
`)
	set, err := ParseWithOverride(src, "slurm")
	require.NoError(t, err)
	assert.Equal(t, "slurm", set.Scheduler())
	assert.True(t, set.Has("scheduler"), "scheduler should be injected into the Set")
}

func TestIsDirectiveLine(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"#ORVIX nodes=2", true},
		{"  #ORVIX nodes=2", true},
		{"#ORVIX", true},
		{"#ORVIX\t", true},
		{"#ORVIXFOO", false},
		{"# ORVIX nodes=2", false},
		{"# orvix nodes=2", false},
		{"#SBATCH --nodes=2", false},
		{"echo hi", false},
		{"", false},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, IsDirectiveLine(c.line), "IsDirectiveLine(%q)", c.line)
	}
}
