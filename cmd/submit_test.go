package cmd

import "testing"

func TestDerivePaths(t *testing.T) {
	cases := []struct {
		name       string
		orig       string
		outScript  string
		outInfo    string
		wantScript string
		wantYaml   string
	}{
		{
			name:       "sh extension defaults",
			orig:       "/tmp/script.sh",
			outScript:  "",
			outInfo:    "",
			wantScript: "/tmp/script.sh.submit",
			wantYaml:   "/tmp/script.sh.info.yaml",
		},
		{
			name:       "bash extension defaults",
			orig:       "/tmp/run.bash",
			outScript:  "",
			outInfo:    "",
			wantScript: "/tmp/run.bash.submit",
			wantYaml:   "/tmp/run.bash.info.yaml",
		},
		{
			name:       "no extension",
			orig:       "/tmp/runme",
			outScript:  "",
			outInfo:    "",
			wantScript: "/tmp/runme.submit",
			wantYaml:   "/tmp/runme.info.yaml",
		},
		{
			name:       "stem contains dot",
			orig:       "/tmp/a.b.sh",
			outScript:  "",
			outInfo:    "",
			wantScript: "/tmp/a.b.sh.submit",
			wantYaml:   "/tmp/a.b.sh.info.yaml",
		},
		{
			name:       "relative path",
			orig:       "case/job/local/orvix_local.sh",
			outScript:  "",
			outInfo:    "",
			wantScript: "case/job/local/orvix_local.sh.submit",
			wantYaml:   "case/job/local/orvix_local.sh.info.yaml",
		},
		{
			name:       "custom output script",
			orig:       "/tmp/script.sh",
			outScript:  "/custom/submit.sh",
			outInfo:    "",
			wantScript: "/custom/submit.sh",
			wantYaml:   "/tmp/script.sh.info.yaml",
		},
		{
			name:       "custom output info",
			orig:       "/tmp/script.sh",
			outScript:  "",
			outInfo:    "/custom/info.yaml",
			wantScript: "/tmp/script.sh.submit",
			wantYaml:   "/custom/info.yaml",
		},
		{
			name:       "custom both",
			orig:       "/tmp/script.sh",
			outScript:  "/custom/submit.sh",
			outInfo:    "/custom/info.yaml",
			wantScript: "/custom/submit.sh",
			wantYaml:   "/custom/info.yaml",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotScript, gotYaml := derivePaths(tc.orig, tc.outScript, tc.outInfo)
			if gotScript != tc.wantScript {
				t.Errorf("script: got %q, want %q", gotScript, tc.wantScript)
			}
			if gotYaml != tc.wantYaml {
				t.Errorf("yaml: got %q, want %q", gotYaml, tc.wantYaml)
			}
		})
	}
}
