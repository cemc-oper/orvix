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
	src := []byte(`#ORVIX comment="hello world"
#ORVIX tags='a b c'
`)
	set, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if got := set.Get("comment"); got != "hello world" {
		t.Errorf("comment = %q, want %q", got, "hello world")
	}
	if got := set.Get("tags"); got != "a b c" {
		t.Errorf("tags = %q, want %q", got, "a b c")
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
