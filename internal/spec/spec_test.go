package spec

import (
	"testing"

	"github.com/cemc-oper/orvix/internal/directive"
	"github.com/stretchr/testify/assert"
)

// TestKnownDirectivesMatchVendoredSchema is the orvix-side contract gate: the
// parser's KnownDirectives must be exactly the keys declared in the vendored
// takflow jobspec schema (hyphenated). If takflow adds/renames/removes a key,
// re-vendoring the schema makes this test fail until the parser is updated.
func TestKnownDirectivesMatchVendoredSchema(t *testing.T) {
	want := map[string]bool{}
	for _, k := range DirectiveKeys() {
		want[k] = true
	}
	assert.Equal(t, want, directive.KnownDirectives,
		"parser.KnownDirectives drifted from the vendored jobspec schema; "+
			"re-vendor internal/spec/jobspec.schema.json and update KnownDirectives")
}

func TestVendoredVersionIsSet(t *testing.T) {
	assert.NotEmpty(t, Version())
}
