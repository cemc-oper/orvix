package scheduler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cemc-oper/orvix/internal/directive"
)

func mustParse(t *testing.T, src string) *directive.Set {
	t.Helper()
	set, err := directive.ParseWithOverride([]byte(src), "donau")
	require.NoError(t, err, "parse error")
	return set
}

func TestDonauPreambleSimpleMappings(t *testing.T) {
	src := `#!/bin/bash
#ORVIX job-name=myjob
#ORVIX output=/tmp/out.log
#ORVIX error=/tmp/err.log
#ORVIX nodes=4
#ORVIX ntasks-per-node=8
#ORVIX queue=operation
#ORVIX account=operation
`
	set := mustParse(t, src)
	d := &Donau{}
	lines, err := d.PreambleFor(set)
	require.NoError(t, err)

	want := []string{
		"#DSUB -n myjob",
		"#DSUB -oo /tmp/out.log",
		"#DSUB -eo /tmp/err.log",
		"#DSUB -nn 4",
		"#DSUB -tpn 8",
		"#DSUB -q operation",
		"#DSUB -A operation",
	}
	assert.Equal(t, want, lines)
}

func TestDonauPreambleProjectApplicationCombined(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want []string
	}{
		{
			name: "both project and application",
			src: `#!/bin/bash
#ORVIX project=105-00-02
#ORVIX application=GRAPES
`,
			want: []string{`#DSUB -d "105-00-02:GRAPES"`},
		},
		{
			name: "only project",
			src: `#!/bin/bash
#ORVIX project=105-00-02
`,
			want: []string{`#DSUB -d "105-00-02"`},
		},
		{
			name: "only application",
			src: `#!/bin/bash
#ORVIX application=GRAPES
`,
			want: []string{`#DSUB -d "GRAPES"`},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			set := mustParse(t, tc.src)
			d := &Donau{}
			got, err := d.PreambleFor(set)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDonauPreambleResourceLine(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want []string
	}{
		{
			name: "cpus-per-task only",
			src: `#!/bin/bash
#ORVIX cpus-per-task=2
`,
			want: []string{`#DSUB -R "cpu=2"`},
		},
		{
			name: "memory only",
			src: `#!/bin/bash
#ORVIX memory=256
`,
			want: []string{`#DSUB -R "mem=256"`},
		},
		{
			name: "both cpus-per-task and memory",
			src: `#!/bin/bash
#ORVIX cpus-per-task=2
#ORVIX memory=256
`,
			want: []string{`#DSUB -R "cpu=2;mem=256"`},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			set := mustParse(t, tc.src)
			d := &Donau{}
			got, err := d.PreambleFor(set)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDonauPreambleExclusive(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want []string
	}{
		{
			name: "bare exclusive",
			src: `#!/bin/bash
#ORVIX exclusive
`,
			want: []string{"#DSUB --exclusive"},
		},
		{
			name: "exclusive with value",
			src: `#!/bin/bash
#ORVIX exclusive=job
`,
			want: []string{"#DSUB --exclusive job"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			set := mustParse(t, tc.src)
			d := &Donau{}
			got, err := d.PreambleFor(set)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDonauPreambleTime(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "pure seconds",
			src: `#!/bin/bash
#ORVIX time=88000
`,
			want: "#DSUB -T 88000",
		},
		{
			name: "HH:MM:SS",
			src: `#!/bin/bash
#ORVIX time=01:00:00
`,
			want: "#DSUB -T 3600",
		},
		{
			name: "MM:SS",
			src: `#!/bin/bash
#ORVIX time=05:30
`,
			want: "#DSUB -T 330",
		},
		{
			name: "duration string 8h",
			src: `#!/bin/bash
#ORVIX time=8h
`,
			want: "#DSUB -T 8h",
		},
		{
			name: "duration string 1h30m",
			src: `#!/bin/bash
#ORVIX time=1h30m
`,
			want: "#DSUB -T 1h30m",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			set := mustParse(t, tc.src)
			d := &Donau{}
			got, err := d.PreambleFor(set)
			require.NoError(t, err)
			require.Len(t, got, 1)
			assert.Equal(t, tc.want, got[0])
		})
	}
}

func TestDonauPreambleSkipsNtasks(t *testing.T) {
	src := `#!/bin/bash
#ORVIX ntasks=64
`
	set := mustParse(t, src)
	d := &Donau{}
	lines, err := d.PreambleFor(set)
	require.NoError(t, err)
	assert.Empty(t, lines, "expected 0 lines for ntasks-only")
}

func TestDonauPreambleJobType(t *testing.T) {
	src := `#!/bin/bash
#ORVIX job-type=cosched
`
	set := mustParse(t, src)
	d := &Donau{}
	lines, err := d.PreambleFor(set)
	require.NoError(t, err)
	assert.Equal(t, []string{"#DSUB --job_type cosched"}, lines)
}

func TestDonauPreambleNodelistQuoted(t *testing.T) {
	src := `#!/bin/bash
#ORVIX nodelist=rp_cme001-418
`
	set := mustParse(t, src)
	d := &Donau{}
	lines, err := d.PreambleFor(set)
	require.NoError(t, err)
	assert.Equal(t, []string{"#DSUB -pn 'rp_cme001-418'"}, lines)
}

func TestDonauPreambleFullParallel(t *testing.T) {
	// Simulates the directives needed to reproduce parallel.sh preamble.
	src := `#!/bin/bash
#ORVIX scheduler=donau
#ORVIX job-name=fcst
#ORVIX output=/g1/u/op_meso/OPER/ECFOUT/cma_meso_1km_v6_0_am/cold/00/model/fcst.1
#ORVIX error=/g1/u/op_meso/OPER/ECFOUT/cma_meso_1km_v6_0_am/cold/00/model/fcst.1.err
#ORVIX nodes=192
#ORVIX ntasks-per-node=120
#ORVIX cpus-per-task=1
#ORVIX queue=operation
#ORVIX account=operation
#ORVIX exclusive=job
#ORVIX time=8h
#ORVIX job-type=cosched
#ORVIX nodelist=rp_cme001-418
#ORVIX project=105-00-02
#ORVIX application=GRAPES
`
	set := mustParse(t, src)
	d := &Donau{}
	lines, err := d.PreambleFor(set)
	require.NoError(t, err)

	// Build a set for quick membership checks.
	have := make(map[string]bool)
	for _, l := range lines {
		have[l] = true
	}

	mustHave := []string{
		"#DSUB -n fcst",
		"#DSUB -oo /g1/u/op_meso/OPER/ECFOUT/cma_meso_1km_v6_0_am/cold/00/model/fcst.1",
		"#DSUB -eo /g1/u/op_meso/OPER/ECFOUT/cma_meso_1km_v6_0_am/cold/00/model/fcst.1.err",
		"#DSUB -nn 192",
		"#DSUB -tpn 120",
		`#DSUB -R "cpu=1"`,
		"#DSUB -q operation",
		"#DSUB -A operation",
		"#DSUB --exclusive job",
		"#DSUB -T 8h",
		"#DSUB --job_type cosched",
		"#DSUB -pn 'rp_cme001-418'",
		`#DSUB -d "105-00-02:GRAPES"`,
	}
	for _, w := range mustHave {
		assert.True(t, have[w], "missing expected line: %s", w)
	}

	// Ensure ntasks is NOT present (it has no Donau equivalent).
	for _, l := range lines {
		assert.NotContains(t, l, "ntasks", "unexpected ntasks line: %s", l)
	}
}

func TestDonauPreambleSerial(t *testing.T) {
	// Simulates the directives needed to reproduce serial.sh preamble.
	src := `#!/bin/bash
#ORVIX scheduler=donau
#ORVIX job-name=radqc_9002
#ORVIX output=/g3/op_meso/OPER/ECFOUT/obs_1h_am/02/get_RADAR/mosaic/mosaic/radqc/part1/radqc_9002.1
#ORVIX error=/g3/op_meso/OPER/ECFOUT/obs_1h_am/02/get_RADAR/mosaic/mosaic/radqc/part1/radqc_9002.1.err
#ORVIX nodes=1
#ORVIX cpus-per-task=1
#ORVIX queue=root.largemem_op
#ORVIX account=root.largemem
#ORVIX time=88000
#ORVIX job-type=cosched
#ORVIX project=105-00-02
#ORVIX application=GRAPES
`
	set := mustParse(t, src)
	d := &Donau{}
	lines, err := d.PreambleFor(set)
	require.NoError(t, err)

	have := make(map[string]bool)
	for _, l := range lines {
		have[l] = true
	}

	mustHave := []string{
		"#DSUB -n radqc_9002",
		"#DSUB -nn 1",
		`#DSUB -R "cpu=1"`,
		"#DSUB -q root.largemem_op",
		"#DSUB -A root.largemem",
		"#DSUB -T 88000",
		"#DSUB --job_type cosched",
		`#DSUB -d "105-00-02:GRAPES"`,
	}
	for _, w := range mustHave {
		assert.True(t, have[w], "missing expected line: %s", w)
	}
}

func TestDonauNormalizeState(t *testing.T) {
	cases := []struct {
		raw  string
		want JobState
	}{
		{"RUNNING", StateRunning},
		{"running", StateRunning},
		{"PENDING", StatePending},
		{"WAITING", StatePending},
		{"QUEUED", StatePending},
		{"COMPLETED", StateCompleted},
		{"DONE", StateCompleted},
		{"FINISHED", StateCompleted},
		{"FAILED", StateFailed},
		{"FAILURE", StateFailed},
		{"ABORTED", StateFailed},
		{"BOOT_FAIL", StateFailed},
		{"NODE_FAIL", StateFailed},
		{"OUT_OF_MEMORY", StateFailed},
		{"OOM", StateFailed},
		{"CANCELLED", StateCancelled},
		{"CANCELED", StateCancelled},
		{"TIMEOUT", StateTimeout},
		{"UNKNOWN", StateUnknown},
		{"whatever", StateUnknown},
	}

	d := &Donau{}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			got := d.NormalizeState(tc.raw)
			assert.Equal(t, tc.want, got)
		})
	}
}
