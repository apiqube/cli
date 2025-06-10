package operations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/apiqube/cli/internal/core/manifests"
	"gopkg.in/yaml.v3"
)

func NormalizeJSON(m manifests.Manifest) ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize manifest: %v", err)
	}
	var raw interface{}
	if err = json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to normalize manifest: %v", err)
	}
	sorted := sortAny(raw)
	// Compact encoding: no spaces, tabs, or newlines
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "") // no indent
	if err = enc.Encode(sorted); err != nil {
		return nil, fmt.Errorf("failed to encode normalized manifest: %v", err)
	}
	// Remove trailing newline added by Encoder
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

func NormalizeYAML(m manifests.Manifest) ([]byte, error) {
	// Marshal to JSON first for canonicalization
	jsonData, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize manifest: %v", err)
	}
	var raw interface{}
	if err = json.Unmarshal(jsonData, &raw); err != nil {
		return nil, fmt.Errorf("failed to normalize manifest: %v", err)
	}
	sorted := sortAny(raw)
	// Marshal to YAML
	data, err := yaml.Marshal(sorted)
	if err != nil {
		return nil, fmt.Errorf("failed to encode normalized manifest: %v", err)
	}
	// Remove trailing spaces and extra newlines
	lines := strings.Split(string(data), "\n")
	var compactLines []string
	for _, line := range lines {
		l := strings.TrimRight(line, " \t")
		if l != "" {
			compactLines = append(compactLines, l)
		}
	}
	return []byte(strings.Join(compactLines, "\n")), nil
}

// Recursively sort all map keys and arrays of maps for canonical output
func sortAny(v any) any {
	switch val := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		res := make(map[string]any, len(val))
		for _, k := range keys {
			res[k] = sortAny(val[k])
		}
		return res
	case []any:
		for i := range val {
			val[i] = sortAny(val[i])
		}
		return val
	default:
		return val
	}
}
