package cmd

import "testing"

func TestDerivePaths(t *testing.T) {
	cases := []struct {
		name       string
		orig       string
		wantScript string
		wantYaml   string
	}{
		{
			name:       "sh extension",
			orig:       "/tmp/script.sh",
			wantScript: "/tmp/script.submit.sh",
			wantYaml:   "/tmp/script.info.yaml",
		},
		{
			name:       "bash extension",
			orig:       "/tmp/run.bash",
			wantScript: "/tmp/run.submit.bash",
			wantYaml:   "/tmp/run.info.yaml",
		},
		{
			name:       "no extension defaults to .sh",
			orig:       "/tmp/runme",
			wantScript: "/tmp/runme.submit.sh",
			wantYaml:   "/tmp/runme.info.yaml",
		},
		{
			name:       "stem contains dot",
			orig:       "/tmp/a.b.sh",
			wantScript: "/tmp/a.b.submit.sh",
			wantYaml:   "/tmp/a.b.info.yaml",
		},
		{
			name:       "relative path",
			orig:       "case/job/local/orvix_local.sh",
			wantScript: "case/job/local/orvix_local.submit.sh",
			wantYaml:   "case/job/local/orvix_local.info.yaml",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotScript, gotYaml := derivePaths(tc.orig)
			if gotScript != tc.wantScript {
				t.Errorf("script: got %q, want %q", gotScript, tc.wantScript)
			}
			if gotYaml != tc.wantYaml {
				t.Errorf("yaml: got %q, want %q", gotYaml, tc.wantYaml)
			}
		})
	}
}
