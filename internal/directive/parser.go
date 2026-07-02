// Package directive parses `#ORVIX ...` preprocessor directives from a script.
package directive

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/cemc-oper/orvix/internal/log"
)

// Marker is the prefix that identifies an orvix directive line.
const Marker = "#ORVIX"

// Directive is one parsed `#ORVIX key=value` (or bare `#ORVIX key`) line.
// Conditional directives carry ConditionKey/ConditionValue (e.g. scheduler=slurm).
// Empty ConditionKey means unconditional.
type Directive struct {
	Key            string
	Value          string // empty for bare keys
	Line           int
	ConditionKey   string // e.g. "scheduler"; empty means unconditional
	ConditionValue string // e.g. "slurm"
}

// KnownDirectives is the canonical set of orvix generic directives.
// Any directive whose key is not in this set is silently discarded during parsing.
var KnownDirectives = map[string]bool{
	"scheduler":       true,
	"job-name":        true,
	"output":          true,
	"error":           true,
	"nodes":           true,
	"ntasks":          true,
	"ntasks-per-node": true,
	"cpus-per-task":   true,
	"time":            true,
	"queue":           true,
	"account":         true,
	"project":         true,
	"application":     true,
	"exclusive":       true,
	"nodelist":        true,
	"job-type":        true,
	"memory":          true,
	"dependency":      true,
	"requeue":         true,
}

// Set is the parsed collection of known directives, indexed by key for quick lookup.
type Set struct {
	Items []Directive
	byKey map[string]string
}

// Get returns the value for a key, or empty string if absent.
func (s *Set) Get(key string) string {
	if s == nil {
		return ""
	}
	return s.byKey[key]
}

// Has reports whether a directive with the given key was set.
func (s *Set) Has(key string) bool {
	if s == nil {
		return false
	}
	_, ok := s.byKey[key]
	return ok
}

// GetOK returns the value for a key together with an ok flag.
// If the key is absent, it returns ("", false).
func (s *Set) GetOK(key string) (string, bool) {
	if s == nil {
		return "", false
	}
	v, ok := s.byKey[key]
	return v, ok
}

// Scheduler returns the scheduler= value, defaulting to "local".
func (s *Set) Scheduler() string {
	v := s.Get("scheduler")
	if v == "" {
		return "local"
	}
	return v
}

// IsDirectiveLine reports whether a raw script line is an `#ORVIX ...` directive.
// Leading whitespace before #ORVIX is allowed; #ORVIX must be followed by
// whitespace or end-of-line so that `#ORVIXFOO` is not mistaken for a directive.
func IsDirectiveLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, Marker) {
		return false
	}
	rest := trimmed[len(Marker):]
	return rest == "" || rest[0] == ' ' || rest[0] == '\t'
}

// Parse extracts `#ORVIX ...` directives from a script header using the
// scheduler specified in the script (defaulting to "local"). It is a
// convenience wrapper for ParseWithOverride with an empty override.
func Parse(src []byte) (*Set, error) {
	return ParseWithOverride(src, "")
}

// ParseWithOverride extracts directives and resolves conditional filtering using
// schedulerOverride when non-empty, otherwise falling back to the script's own
// scheduler= directive. The override value is also written into the returned
// Set so that Scheduler() and downstream consumers see the overridden value.
func ParseWithOverride(src []byte, schedulerOverride string) (*Set, error) {
	scanner := bufio.NewScanner(bytes.NewReader(src))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var raw []Directive

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimRight(scanner.Text(), "\r")
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "#") {
			break
		}
		if !strings.HasPrefix(trimmed, Marker) {
			continue
		}
		rest := trimmed[len(Marker):]
		if rest == "" {
			continue
		}
		if rest[0] != ' ' && rest[0] != '\t' {
			// e.g. "#ORVIXFOO" — not our directive.
			continue
		}
		body := strings.TrimSpace(rest)
		if body == "" {
			continue
		}
		if strings.HasPrefix(body, "[") && !strings.Contains(body, "]") {
			return nil, fmt.Errorf("line %d: unclosed [ in condition", lineNo)
		}

		condKey, condValue, body := splitCondition(body)
		if body == "" {
			continue
		}

		dir, err := parseDirective(body)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
		dir.Line = lineNo
		dir.ConditionKey = condKey
		dir.ConditionValue = condValue
		raw = append(raw, dir)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	log.Debugf("[directive] parsed %d raw directive line(s)", len(raw))

	// Determine scheduler: override takes precedence.
	sched := schedulerOverride
	if sched == "" {
		for _, dir := range raw {
			if dir.Key == "scheduler" && dir.ConditionKey == "" {
				sched = dir.Value
				break
			}
		}
	}
	if sched == "" {
		sched = "local"
	}
	log.Debugf("[directive] effective scheduler: %s", sched)

	// Filter by condition and known-directive set, then build the final Set.
	set := &Set{byKey: make(map[string]string)}
	for _, dir := range raw {
		if dir.ConditionKey != "" && (dir.ConditionKey != "scheduler" || dir.ConditionValue != sched) {
			continue // skip conditionally excluded directives
		}
		if !KnownDirectives[dir.Key] {
			log.Debugf("[directive] dropping unknown directive: %s", dir.Key)
			continue // skip unknown directives (no backward-compatible passthrough)
		}
		set.Items = append(set.Items, dir)
		set.byKey[dir.Key] = dir.Value
	}
	log.Debugf("[directive] %d directive(s) after filtering", len(set.Items))

	// Apply scheduler override into the Set so downstream code sees it.
	if schedulerOverride != "" {
		set.byKey["scheduler"] = schedulerOverride
		found := false
		for i := range set.Items {
			if set.Items[i].Key == "scheduler" {
				set.Items[i].Value = schedulerOverride
				found = true
				break
			}
		}
		if !found {
			set.Items = append([]Directive{{Key: "scheduler", Value: schedulerOverride}}, set.Items...)
		}
	}

	return set, nil
}

// splitCondition checks if a directive body starts with `[key=value]`.
// If so, it returns (key, value, remaining_body).
// Otherwise it returns ("", "", body).
func splitCondition(body string) (condKey, condValue, rest string) {
	if !strings.HasPrefix(body, "[") {
		return "", "", body
	}
	close := strings.IndexByte(body, ']')
	if close < 0 {
		return "", "", body
	}
	inner := body[1:close]
	rest = strings.TrimSpace(body[close+1:])

	eq := strings.IndexByte(inner, '=')
	if eq < 0 {
		// No = inside brackets: not a valid condition, treat as normal body
		return "", "", body
	}
	return inner[:eq], inner[eq+1:], rest
}

func parseDirective(body string) (Directive, error) {
	tokens, err := tokenize(body)
	if err != nil {
		return Directive{}, err
	}
	if len(tokens) == 0 {
		return Directive{}, fmt.Errorf("empty directive")
	}
	first := tokens[0]

	if eq := strings.IndexByte(first, '='); eq >= 0 {
		if eq == 0 {
			return Directive{}, fmt.Errorf("empty key in %q", body)
		}
		key := first[:eq]
		value := first[eq+1:]
		if len(tokens) > 1 {
			value = value + " " + strings.Join(tokens[1:], " ")
		}
		return Directive{Key: key, Value: value}, nil
	}

	// Bare key: must be the only token.
	if len(tokens) > 1 {
		return Directive{}, fmt.Errorf("expected key=value or bare key, got %q", body)
	}
	return Directive{Key: first, Value: ""}, nil
}

// tokenize splits a directive body into whitespace-separated tokens, honoring
// "..." and '...' quoting. No escape processing or env expansion.
func tokenize(s string) ([]string, error) {
	var (
		tokens  []string
		cur     strings.Builder
		inQuote byte
	)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inQuote != 0 {
			if c == inQuote {
				inQuote = 0
				continue
			}
			cur.WriteByte(c)
			continue
		}
		switch c {
		case '"', '\'':
			inQuote = c
		case ' ', '\t':
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
		}
	}
	if inQuote != 0 {
		return nil, fmt.Errorf("unterminated %c-quote", inQuote)
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens, nil
}
