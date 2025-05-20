package operations

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/apiqube/cli/internal/core/manifests"
	"gopkg.in/yaml.v3"
)

func NormalizeYAML(m manifests.Manifest) ([]byte, error) {
	data, err := yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed normilize manifest: %v", err)
	}

	var raw map[string]interface{}
	if err = yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed normilize manifest: %v", err)
	}

	sorted := sortMapKeys(raw)

	return yaml.Marshal(sorted)
}

func NormalizeJSON(m manifests.Manifest) ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed normilize manifest: %v", err)
	}

	var raw map[string]interface{}
	if err = json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed normilize manifest: %v", err)
	}

	sorted := sortMapKeys(raw)

	return json.Marshal(sorted)
}

func sortMapKeys(m map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	keys := make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if nested, ok := m[k].(map[string]interface{}); ok {
			res[k] = sortMapKeys(nested)
		} else {
			res[k] = m[k]
		}
	}

	return res
}
