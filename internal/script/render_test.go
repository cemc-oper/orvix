package script

import (
	"strings"
	"testing"

	"github.com/cemc-oper/orvix/internal/directive"
	"github.com/cemc-oper/orvix/internal/scheduler"
)

func renderWithSlurm(t *testing.T, src string) string {
	t.Helper()
	d, err := directive.Parse([]byte(src))
	if err != nil {
		t.Fatalf("parse directives: %v", err)
	}
	out, err := Render([]byte(src), d, &scheduler.SLURM{})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return string(out)
}

func TestRenderShebangFirstLine(t *testing.T) {
	src := "#!/bin/bash\n#ORVIX time=00:15:00\n\necho hi\n"
	got := renderWithSlurm(t, src)
	want := "#!/bin/bash\n#SBATCH --time=00:15:00\n\necho hi\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

// ecFlow-created job files may carry leading blank line(s) before the
// shebang; the shebang must still be hoisted ahead of the #SBATCH preamble.
func TestRenderShebangAfterLeadingBlankLines(t *testing.T) {
	src := "\n\n#! /usr/bin/env bash\n#\n#ORVIX time=5:00\n#ORVIX queue=serial\n\necho hi\n"
	got := renderWithSlurm(t, src)
	want := "#! /usr/bin/env bash\n#SBATCH --time=5:00\n#SBATCH --partition=serial\n\n#\n\necho hi\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
	if !strings.HasPrefix(got, "#!") {
		t.Errorf("first line must be the shebang, got: %.40q", got)
	}
}

func TestRenderNoShebang(t *testing.T) {
	src := "#ORVIX time=5:00\necho hi\n"
	got := renderWithSlurm(t, src)
	want := "#SBATCH --time=5:00\n\necho hi\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

// A shebang that is not the first non-empty line is an ordinary body line.
func TestRenderShebangAfterCommentNotHoisted(t *testing.T) {
	src := "# comment\n#!/bin/bash\n#ORVIX time=5:00\necho hi\n"
	got := renderWithSlurm(t, src)
	want := "#SBATCH --time=5:00\n\n# comment\n#!/bin/bash\necho hi\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}
