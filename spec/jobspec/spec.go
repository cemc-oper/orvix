// Package jobspec is the canonical home of the jobspec contract: the
// language-neutral job-run resource description shared by workflow generators
// (takflow and its applications) and orvix. The schema (jobspec.schema.json),
// the contract VERSION, and the conformance vectors/goldens live here and are
// authoritative; takflow vendors a read-only copy pinned to an orvix ref.
//
// The package exposes the schema to the rest of orvix via go:embed of the
// files next to it, so there is exactly one copy of the contract in this
// repository.
package jobspec

import (
	_ "embed"
	"encoding/json"
	"sort"
	"strings"
)

//go:embed jobspec.schema.json
var schemaJSON []byte

//go:embed VERSION
var versionRaw string

type schemaDoc struct {
	Properties map[string]json.RawMessage `json:"properties"`
}

// Version returns the contract version (semantic version string).
func Version() string {
	return strings.TrimSpace(versionRaw)
}

// PropertyKeys returns the schema property names in snake_case (e.g. "job_name").
func PropertyKeys() []string {
	var doc schemaDoc
	if err := json.Unmarshal(schemaJSON, &doc); err != nil {
		panic("jobspec.schema.json is not valid JSON: " + err.Error())
	}
	keys := make([]string, 0, len(doc.Properties))
	for k := range doc.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// DirectiveKeys returns the schema property names as #ORVIX directive keys
// (hyphenated, e.g. "job-name"). This is the set orvix's parser must recognize.
func DirectiveKeys() []string {
	keys := PropertyKeys()
	out := make([]string, len(keys))
	for i, k := range keys {
		out[i] = strings.ReplaceAll(k, "_", "-")
	}
	return out
}
