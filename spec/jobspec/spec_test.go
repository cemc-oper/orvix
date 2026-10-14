package jobspec

import (
	"testing"

	"github.com/cemc-oper/orvix/internal/directive"
	"github.com/stretchr/testify/assert"
)

// TestKnownDirectivesMatchSchema is the orvix-side contract gate: the parser's
// KnownDirectives must be exactly the keys declared in the jobspec schema
// (hyphenated). A contract change that adds/renames/removes a key fails this
// test until the parser is updated in the same commit.
func TestKnownDirectivesMatchSchema(t *testing.T) {
	want := map[string]bool{}
	for _, k := range DirectiveKeys() {
		want[k] = true
	}
	assert.Equal(t, want, directive.KnownDirectives,
		"directive.KnownDirectives drifted from spec/jobspec/jobspec.schema.json; "+
			"update the schema and the parser together")
}

func TestVersionIsSet(t *testing.T) {
	assert.NotEmpty(t, Version())
}
