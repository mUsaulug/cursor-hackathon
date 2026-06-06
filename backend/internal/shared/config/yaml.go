// Package config loads the CivicLens rule files (ontology, priority policy,
// confidence thresholds) without any third-party dependency. The backend.mdc
// rule forbids non-stdlib packages, so this file implements a deliberately tiny
// YAML reader that supports exactly the subset our rule files use:
//
//	version: "1.0"          # top-level scalar
//	some_map:               # 2-level string map
//	  key: value
//	some_list:              # list of scalars
//	  - item
//
// It is NOT a general YAML parser. Anything beyond this subset is unsupported.
package config

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// yamlDoc is the parsed representation of one rule file.
type yamlDoc struct {
	scalars map[string]string
	maps    map[string]map[string]string
	lists   map[string][]string
}

func newYAMLDoc() *yamlDoc {
	return &yamlDoc{
		scalars: map[string]string{},
		maps:    map[string]map[string]string{},
		lists:   map[string][]string{},
	}
}

// parseYAML reads the supported subset described in the package doc.
func parseYAML(data []byte) (*yamlDoc, error) {
	doc := newYAMLDoc()
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var currentKey string // the most recent top-level "key:" with no inline value
	var currentKind byte  // 'm' map, 'l' list, 0 unknown-yet
	lineNo := 0

	for sc.Scan() {
		lineNo++
		raw := sc.Text()
		line := stripComment(raw)
		if strings.TrimSpace(line) == "" {
			continue
		}

		indented := len(line) > 0 && (line[0] == ' ' || line[0] == '\t')
		trimmed := strings.TrimSpace(line)

		if !indented {
			// Top-level entry. Reset any open block.
			currentKey = ""
			currentKind = 0

			key, val, ok := splitKeyValue(trimmed)
			if !ok {
				return nil, fmt.Errorf("config: line %d: expected 'key:' or 'key: value', got %q", lineNo, trimmed)
			}
			if val == "" {
				// Opens a nested block (map or list); kind decided by children.
				currentKey = key
				continue
			}
			doc.scalars[key] = unquote(val)
			continue
		}

		// Indented child line; must belong to an open top-level key.
		if currentKey == "" {
			return nil, fmt.Errorf("config: line %d: indented entry without a parent key", lineNo)
		}

		if strings.HasPrefix(trimmed, "- ") || trimmed == "-" {
			if currentKind == 'm' {
				return nil, fmt.Errorf("config: line %d: list item inside a map block %q", lineNo, currentKey)
			}
			currentKind = 'l'
			item := unquote(strings.TrimSpace(strings.TrimPrefix(trimmed, "-")))
			doc.lists[currentKey] = append(doc.lists[currentKey], item)
			continue
		}

		key, val, ok := splitKeyValue(trimmed)
		if !ok {
			return nil, fmt.Errorf("config: line %d: expected 'key: value', got %q", lineNo, trimmed)
		}
		if currentKind == 'l' {
			return nil, fmt.Errorf("config: line %d: map entry inside a list block %q", lineNo, currentKey)
		}
		currentKind = 'm'
		if doc.maps[currentKey] == nil {
			doc.maps[currentKey] = map[string]string{}
		}
		doc.maps[currentKey][key] = unquote(val)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("config: scan: %w", err)
	}
	return doc, nil
}

func (d *yamlDoc) scalar(key string) (string, bool) {
	v, ok := d.scalars[key]
	return v, ok
}

func (d *yamlDoc) float(key string) (float64, bool, error) {
	v, ok := d.scalars[key]
	if !ok {
		return 0, false, nil
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil {
		return 0, true, fmt.Errorf("config: key %q is not a number: %w", key, err)
	}
	return f, true, nil
}

// parseFloat trims and parses a float scalar.
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

// stripComment removes a trailing "#..." comment. Our rule files never quote a
// literal '#', so a simple split is safe for this subset.
func stripComment(line string) string {
	if i := strings.Index(line, "#"); i >= 0 {
		return line[:i]
	}
	return line
}

// splitKeyValue parses "key: value" / "key:" into (key, value, ok).
func splitKeyValue(s string) (string, string, bool) {
	i := strings.Index(s, ":")
	if i < 0 {
		return "", "", false
	}
	key := strings.TrimSpace(s[:i])
	val := strings.TrimSpace(s[i+1:])
	if key == "" {
		return "", "", false
	}
	return key, val, true
}

// unquote removes surrounding single or double quotes if present.
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
