package directive

import "testing"

func TestParseBasic(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX nodes=2
#ORVIX time=01:00:00

echo "hello"
`)
	set, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if got := set.Scheduler(); got != "slurm" {
		t.Errorf("scheduler = %q, want slurm", got)
	}
	if got := set.Get("nodes"); got != "2" {
		t.Errorf("nodes = %q, want 2", got)
	}
	if got := set.Get("time"); got != "01:00:00" {
		t.Errorf("time = %q, want 01:00:00", got)
	}
}

func TestParseBareKey(t *testing.T) {
	src := []byte(`#ORVIX exclusive
#ORVIX nodes=2
`)
	set, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if !set.Has("exclusive") {
		t.Error("exclusive should be set")
	}
	if got := set.Get("exclusive"); got != "" {
		t.Errorf("exclusive value = %q, want empty", got)
	}
	if got := set.Get("nodes"); got != "2" {
		t.Errorf("nodes = %q, want 2", got)
	}
}

func TestParseQuotedValue(t *testing.T) {
	src := []byte(`#ORVIX application="hello world"
#ORVIX partition='a b c'
`)
	set, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if got := set.Get("application"); got != "hello world" {
		t.Errorf("application = %q, want %q", got, "hello world")
	}
	if got := set.Get("partition"); got != "a b c" {
		t.Errorf("partition = %q, want %q", got, "a b c")
	}
}

func TestParseStopsAtFirstNonCommentLine(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX nodes=2

echo before
#ORVIX time=99
`)
	set, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if got := set.Get("nodes"); got != "2" {
		t.Errorf("nodes = %q, want 2", got)
	}
	if set.Has("time") {
		t.Error("time should not be parsed (after non-comment line)")
	}
}

func TestParseDefaultSchedulerIsLocal(t *testing.T) {
	set, err := Parse([]byte("#!/bin/bash\necho hi\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got := set.Scheduler(); got != "local" {
		t.Errorf("scheduler = %q, want local", got)
	}
}

func TestParseSkipsRegularComments(t *testing.T) {
	src := []byte(`#!/bin/bash
# this is a normal comment
#ORVIX nodes=2
# orvix lowercase=ignored
#ORVIXNOSPACE=alsoignored
`)
	set, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if got := set.Get("nodes"); got != "2" {
		t.Errorf("nodes = %q, want 2", got)
	}
	if set.Has("lowercase") {
		t.Error("lowercase orvix should not be matched")
	}
	if set.Has("NOSPACE") {
		t.Error("#ORVIXNOSPACE should not be matched (no whitespace after marker)")
	}
}

func TestParseRejectsBarePlusValue(t *testing.T) {
	src := []byte(`#ORVIX nodes 2
`)
	_, err := Parse(src)
	if err == nil {
		t.Error("expected error for `key value` (no =)")
	}
}

func TestParseRejectsEmptyKey(t *testing.T) {
	src := []byte(`#ORVIX =foo
`)
	_, err := Parse(src)
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestParseConditionalKeepsMatching(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX scheduler=donau
#ORVIX nodes=2
#ORVIX [scheduler=donau] nodelist=rp_cme001-418
#ORVIX [scheduler=slurm] partition=normal
`)
	set, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if !set.Has("nodes") {
		t.Error("nodes should be present (unconditional)")
	}
	if !set.Has("nodelist") {
		t.Error("nodelist should be present ([scheduler=donau] matches)")
	}
	if set.Has("partition") {
		t.Error("partition should be excluded ([scheduler=slurm] does not match scheduler=donau)")
	}
	if got := set.Get("nodelist"); got != "rp_cme001-418" {
		t.Errorf("nodelist = %q, want rp_cme001-418", got)
	}
}

func TestParseConditionalDropsNonMatching(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX nodes=2
#ORVIX [scheduler=donau] nodelist=rp_cme001-418
#ORVIX [scheduler=slurm] partition=normal
`)
	set, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if !set.Has("nodes") {
		t.Error("nodes should be present (unconditional)")
	}
	if set.Has("nodelist") {
		t.Error("nodelist should be excluded ([scheduler=donau] does not match scheduler=slurm)")
	}
	if !set.Has("partition") {
		t.Error("partition should be present ([scheduler=slurm] matches scheduler=slurm)")
	}
}

func TestParseConditionalBareKey(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX [scheduler=slurm] exclusive
#ORVIX [scheduler=donau] cosched
`)
	set, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if !set.Has("exclusive") {
		t.Error("exclusive should be present ([scheduler=slurm] matches)")
	}
	if set.Has("cosched") {
		t.Error("cosched should be excluded ([scheduler=donau] does not match)")
	}
}

func TestParseConditionalNoSpace(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX scheduler=donau
#ORVIX [scheduler=donau]nodelist=rp_cme001-418
`)
	set, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if !set.Has("nodelist") {
		t.Error("nodelist should be present (no space after ] is ok)")
	}
}

func TestParseConditionalEmptyBrackets(t *testing.T) {
	// Empty [] is not a valid condition (no = inside), treated as part of the key.
	src := []byte(`#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX [] nodes=4
`)
	_, err := Parse(src)
	if err == nil {
		t.Error("expected error for [] (no key=value inside brackets)")
	}
}

func TestParseConditionalUnclosedBracket(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX [scheduler=donau nodelist=rp_cme001-418
`)
	// Unclosed [ is treated as part of the key, so parseDirective will error
	_, err := Parse(src)
	if err == nil {
		t.Error("expected error for unclosed [")
	}
}

func TestParseConditionalMixedWithUnconditional(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX scheduler=donau
#ORVIX time=01:00:00
#ORVIX [scheduler=donau] time=120
`)
	set, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	// The conditional [scheduler=donau] time=120 should override the unconditional time=01:00:00
	if got := set.Get("time"); got != "120" {
		t.Errorf("time = %q, want 120 (conditional overrides unconditional)", got)
	}
}

func TestParseWithOverride(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX nodes=2
#ORVIX [scheduler=donau] nodelist=rp_cme001-418
#ORVIX [scheduler=slurm] partition=normal
`)
	// Override to donau: condition filtering uses "donau", and Set.Scheduler() returns "donau"
	set, err := ParseWithOverride(src, "donau")
	if err != nil {
		t.Fatal(err)
	}
	if got := set.Scheduler(); got != "donau" {
		t.Errorf("scheduler = %q, want donau", got)
	}
	if set.Has("partition") {
		t.Error("partition should be excluded (condition matches slurm, not donau)")
	}
	if !set.Has("nodelist") {
		t.Error("nodelist should be present (condition matches donau)")
	}
}

func TestParseWithOverrideDefaultLocal(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX nodes=2
`)
	set, err := ParseWithOverride(src, "")
	if err != nil {
		t.Fatal(err)
	}
	if got := set.Scheduler(); got != "local" {
		t.Errorf("scheduler = %q, want local", got)
	}
}

func TestParseWithOverrideSetsScheduler(t *testing.T) {
	src := []byte(`#!/bin/bash
#ORVIX nodes=2
`)
	set, err := ParseWithOverride(src, "slurm")
	if err != nil {
		t.Fatal(err)
	}
	if got := set.Scheduler(); got != "slurm" {
		t.Errorf("scheduler = %q, want slurm", got)
	}
	if !set.Has("scheduler") {
		t.Error("scheduler should be injected into the Set")
	}
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
		if got := IsDirectiveLine(c.line); got != c.want {
			t.Errorf("IsDirectiveLine(%q) = %v, want %v", c.line, got, c.want)
		}
	}
}
