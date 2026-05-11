// Package directive parses `#ORVIX ...` preprocessor directives from a script.
package directive

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// Marker is the prefix that identifies an orvix directive line.
const Marker = "#ORVIX"

// Directive is one parsed `#ORVIX key=value` (or bare `#ORVIX key`) line.
type Directive struct {
	Key   string
	Value string // empty for bare keys
	Line  int
}

// Set is the parsed collection of directives, indexed by key for quick lookup.
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

// Parse extracts `#ORVIX ...` directives from a script header.
//
// Each directive line carries a single key=value pair, or a bare key for flags
// without values:
//
//	#ORVIX scheduler=slurm
//	#ORVIX partition=gpu
//	#ORVIX nodes=2
//	#ORVIX comment="long running benchmark"
//	#ORVIX exclusive                          (bare key, no value)
//
// Parsing stops at the first non-blank, non-comment line.
func Parse(src []byte) (*Set, error) {
	set := &Set{byKey: make(map[string]string)}
	scanner := bufio.NewScanner(bytes.NewReader(src))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

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

		dir, err := parseDirective(body)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
		dir.Line = lineNo
		set.Items = append(set.Items, dir)
		set.byKey[dir.Key] = dir.Value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return set, nil
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
