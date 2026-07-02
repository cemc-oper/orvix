// Package spec vendors the canonical takflow jobspec contract and exposes it to
// the rest of orvix. The schema (jobspec.schema.json) is the single source of
// truth, authored in the takflow repo; this is a verbatim vendored copy kept in
// sync by the conformance harness. Do not hand-edit — re-vendor from takflow.
package spec

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

// Version returns the vendored contract version (semantic version string).
func Version() string {
	return strings.TrimSpace(versionRaw)
}

// PropertyKeys returns the schema property names in snake_case (e.g. "job_name").
func PropertyKeys() []string {
	var doc schemaDoc
	if err := json.Unmarshal(schemaJSON, &doc); err != nil {
		panic("vendored jobspec.schema.json is not valid JSON: " + err.Error())
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
